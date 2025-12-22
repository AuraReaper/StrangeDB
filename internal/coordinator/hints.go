package coordinator

import (
	"context"
	"sync"
	"time"

	"github.com/AuraReaper/strangedb/internal/storage"
	grpcTransport "github.com/AuraReaper/strangedb/internal/transport/grpc"
	pb "github.com/AuraReaper/strangedb/internal/transport/grpc/proto"
)

type Hint struct {
	TargetNode string          `json:"target_node"`
	Record     *storage.Record `json:"record"`
	CreatedAt  time.Time       `json:"created_at"`
	Attempts   int             `json:"attempts"`
}

type HintStore struct {
	mu       sync.RWMutex
	hints    map[string][]*Hint //node -> hints
	maxHints int
	ttl      time.Duration
}

func NewHintStore(maxHints int, ttl time.Duration) *HintStore {
	hs := &HintStore{
		hints:    make(map[string][]*Hint),
		maxHints: maxHints,
		ttl:      ttl,
	}

	go hs.cleanupLoop()

	return hs
}

func (hs *HintStore) AddHint(targetNode string, record *storage.Record) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hint := &Hint{
		TargetNode: targetNode,
		Record:     record,
		CreatedAt:  time.Now(),
		Attempts:   0,
	}

	hints := hs.hints[targetNode]

	if len(hints) >= hs.maxHints {
		hints = hints[1:]
	}

	hs.hints[targetNode] = append(hints, hint)
}

func (hs *HintStore) GetHints(targetNode string) []*Hint {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	hints := hs.hints[targetNode]
	result := make([]*Hint, len(hints))
	copy(result, hints)

	return result
}

func (hs *HintStore) RemoveHint(targetNode string, record *storage.Record) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hints := hs.hints[targetNode]
	filtered := make([]*Hint, 0, len(hints))

	for _, h := range hints {
		if h.Record.Key != record.Key {
			filtered = append(filtered, h)
		}
	}

	hs.hints[targetNode] = filtered
}

func (hs *HintStore) ClearHints(targetNode string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	delete(hs.hints, targetNode)
}

func (hs *HintStore) cleanupLoop() {
	ticker := time.NewTimer(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		hs.cleanupExpired()
	}
}

func (hs *HintStore) cleanupExpired() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	now := time.Now()

	for node, hints := range hs.hints {
		filtered := make([]*Hint, 0, len(hints))

		for _, h := range hints {
			if now.Sub(h.CreatedAt) < hs.ttl {
				filtered = append(filtered, h)
			}
		}

		hs.hints[node] = filtered
	}
}

func (hs *HintStore) Nodes() []string {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	nodes := make([]string, 0, len(hs.hints))
	for node := range hs.hints {
		nodes = append(nodes, node)
	}
	return nodes
}

type HintedHandoff struct {
	store      *HintStore
	grpcClient grpcTransport.Client
	interval   time.Duration
	stopCh     chan struct{}
}

func (hh *HintedHandoff) Start() {
	go hh.replayLoop()
}

func (hh *HintedHandoff) replayOnce() {
	nodes := hh.store.Nodes()

	for _, node := range nodes {
		hints := hh.store.GetHints(node)

		for _, hint := range hints {
			hh.replayHint(node, hint)
		}
	}
}

func (hh *HintedHandoff) replayHint(node string, hint *Hint) {
	var err error

	if hint.Record.Tombstone {
		_, err = hh.grpcClient.Delete(
			context.Background(), node, hint.Record.Key,
			&pb.Timestamp{
				WallTime: hint.Record.Timestamp.WallTime,
				Logical:  hint.Record.Timestamp.Logical,
				NodeId:   hint.Record.Timestamp.NodeID,
			},
		)
	} else {
		_, err = hh.grpcClient.Set(
			context.Background(),
			node,
			&pb.Record{
				Key:   hint.Record.Key,
				Value: hint.Record.Value,
				Timestamp: &pb.Timestamp{
					WallTime: hint.Record.Timestamp.WallTime,
					Logical:  hint.Record.Timestamp.Logical,
					NodeId:   hint.Record.Timestamp.NodeID,
				},
				Tombstone: hint.Record.Tombstone,
			},
		)
	}

	if err == nil {
		hh.store.RemoveHint(node, hint.Record)
	} else {
		hint.Attempts++
	}
}

func (hh *HintedHandoff) replayLoop() {
	ticker := time.NewTicker(hh.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hh.replayOnce()
		case <-hh.stopCh:
			return
		}
	}
}
