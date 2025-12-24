package coordinator

import (
	"context"
	"errors"
	"sync"

	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/ring"
	"github.com/AuraReaper/strangedb/internal/storage"
	grpcTransport "github.com/AuraReaper/strangedb/internal/transport/grpc"
	pb "github.com/AuraReaper/strangedb/internal/transport/grpc/proto"
	"github.com/rs/zerolog"
)

var (
	ErrQuorumNotReached = errors.New("quorum not reached")
	ErrNoNodesAvailable = errors.New("no nodes available")
)

type Coordinator struct {
	nodeURL      string
	ring         *ring.ConsistentHashRing
	storage      storage.Storage
	clock        *hlc.Clock
	grpcClient   *grpcTransport.Client
	replicationN int
	readQuorum   int
	writeQuorum  int
	log          zerolog.Logger
	readRepair   *ReadRepair
	hintStore    *HintStore
}

func New(nodeURL string, ring *ring.ConsistentHashRing, storage storage.Storage, clock *hlc.Clock,
	grpcClient *grpcTransport.Client, replicationN, readQuorum, writeQuorum int, log zerolog.Logger) *Coordinator {
	return &Coordinator{
		nodeURL:      nodeURL,
		ring:         ring,
		storage:      storage,
		clock:        clock,
		grpcClient:   grpcClient,
		replicationN: replicationN,
		readQuorum:   readQuorum,
		writeQuorum:  writeQuorum,
		log:          log,
	}
}

func (c *Coordinator) SetReadRepair(rr *ReadRepair) {
	c.readRepair = rr
}

func (c *Coordinator) SetHintStore(hs *HintStore) {
	c.hintStore = hs
}

func (c *Coordinator) Storage() storage.Storage {
	return c.storage
}

func (c *Coordinator) Get(ctx context.Context, key string) (*storage.Record, error) {
	replicas := c.ring.GetReplicas(key, c.replicationN)
	if len(replicas) == 0 {
		return nil, ErrNoNodesAvailable
	}

	log := c.log.With().
		Str("key", key).
		Str("operation", "GET").
		Strs("replicas", replicas).
		Int("quorum_required", c.readQuorum).
		Logger()

	log.Info().Msg("performing get operation")

	type getResult struct {
		record *storage.Record
		err    error
		node   string
	}

	resultCh := make(chan getResult, len(replicas))
	var wg sync.WaitGroup

	for _, replica := range replicas {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			var (
				r   *storage.Record
				err error
			)

			if addr == c.nodeURL {
				// local read
				r, err = c.storage.Get(key)
			} else {
				// remote read
				resp, e := c.grpcClient.Get(ctx, addr, key)
				if e != nil {
					err = e
				} else if resp.Found {
					r = &storage.Record{
						Key:   resp.Record.Key,
						Value: resp.Record.Value,
						Timestamp: hlc.Timestamp{
							WallTime: resp.Record.Timestamp.WallTime,
							Logical:  resp.Record.Timestamp.Logical,
							NodeID:   resp.Record.Timestamp.NodeId,
						},
						Tombstone: resp.Record.Tombstone,
					}
				} else {
					err = storage.ErrKeyNotFound
				}
			}

			resultCh <- getResult{
				record: r,
				err:    err,
				node:   addr,
			}
		}(replica)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	responsesByAddr := make(map[string]*storage.Record)
	var records []*storage.Record
	var failedNodes []string
	successCount := 0

	for res := range resultCh {
		if res.err == nil && res.record != nil {
			responsesByAddr[res.node] = res.record
			records = append(records, res.record)
			successCount++
		} else {
			responsesByAddr[res.node] = nil
		}
	}

	log = log.With().
		Int("acks_received", successCount).
		Strs("failed_nodes", failedNodes).
		Logger()

	if successCount == 0 {
		log.Error().Msg("get failed: no replicas responded")
		return nil, ErrQuorumNotReached
	}

	latest := c.findLatest(records)

	if c.readRepair != nil && latest != nil {
		results := c.readRepair.AnalyzeResponses(responsesByAddr, latest)
		go c.readRepair.CheckAndRepair(context.Background(), results)
	}

	// Quorum check
	if successCount >= c.readQuorum {
		if latest == nil {
			log.Info().Msg("key not found")
			return nil, storage.ErrKeyNotFound
		}
		log.Info().Msg("get operation successful")
		return latest, nil
	}

	log.Warn().Msg("quorum not reached, returning partial result")
	if latest == nil {
		return nil, storage.ErrKeyNotFound
	}
	return latest, nil
}

func (c *Coordinator) Set(ctx context.Context, key string, value []byte) (*storage.Record, error) {
	replicas := c.ring.GetReplicas(key, c.replicationN)
	if len(replicas) == 0 {
		return nil, ErrNoNodesAvailable
	}

	log := c.log.With().
		Str("key", key).
		Str("operation", "SET").
		Strs("replicas", replicas).
		Int("quorum_required", c.writeQuorum).
		Logger()

	log.Info().Msg("performing set operation")

	ts := c.clock.Now()
	record := &storage.Record{
		Key:       key,
		Value:     value,
		Timestamp: ts,
		Tombstone: false,
	}

	type setResult struct {
		err  error
		node string
	}

	resultCh := make(chan setResult, len(replicas))
	var wg sync.WaitGroup

	for _, replica := range replicas {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			var err error
			if addr == c.nodeURL {
				// local
				err = c.storage.Set(record)
			} else {
				// remote
				_, err = c.grpcClient.Set(ctx, addr, &pb.Record{
					Key:   record.Key,
					Value: record.Value,
					Timestamp: &pb.Timestamp{
						WallTime: record.Timestamp.WallTime,
						Logical:  record.Timestamp.Logical,
						NodeId:   record.Timestamp.NodeID,
					},
					Tombstone: false,
				})
			}

			resultCh <- setResult{err: err, node: addr}
		}(replica)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var failedNodes []string
	var successCount int
	for res := range resultCh {
		if res.err == nil {
			successCount++
		} else {
			failedNodes = append(failedNodes, res.node)
		}
	}

	log = log.With().
		Int("acks_received", successCount).
		Strs("failed_nodes", failedNodes).
		Logger()

	if successCount >= c.writeQuorum {
		log.Info().Msg("set operation successful")
		return record, nil
	}

	if successCount > 0 {
		log.Warn().Msg("quorum not reached, but returning partial results")
		return record, nil
	}

	log.Error().Msg("quorum not reached, set operation failed")
	return nil, ErrQuorumNotReached
}

func (c *Coordinator) Delete(ctx context.Context, key string) error {
	replicas := c.ring.GetReplicas(key, c.replicationN)
	if len(replicas) == 0 {
		return ErrNoNodesAvailable
	}

	log := c.log.With().
		Str("key", key).
		Str("operation", "DELETE").
		Strs("replicas", replicas).
		Int("quorum_required", c.writeQuorum).
		Logger()

	log.Info().Msg("performing delete operation")

	ts := c.clock.Now()

	type deleteResult struct {
		err  error
		node string
	}

	resultCh := make(chan deleteResult, len(replicas))
	var wg sync.WaitGroup

	for _, replica := range replicas {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			var err error
			if addr == c.nodeURL {
				// local
				err = c.storage.Delete(key, ts)
			} else {
				// remote
				_, err = c.grpcClient.Delete(ctx, addr, key, &pb.Timestamp{
					WallTime: ts.WallTime,
					Logical:  ts.Logical,
					NodeId:   ts.NodeID,
				})
			}

			resultCh <- deleteResult{err: err, node: addr}
		}(replica)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var failedNodes []string
	var successCount int
	for res := range resultCh {
		if res.err == nil {
			successCount++
		} else {
			failedNodes = append(failedNodes, res.node)
		}
	}

	log = log.With().
		Int("acks_received", successCount).
		Strs("failed_nodes", failedNodes).
		Logger()

	if successCount >= c.writeQuorum {
		log.Info().Msg("delete operation successful")
		return nil
	}

	if successCount > 0 {
		log.Warn().Msg("quorum not reached, but returning partial results")
		return nil
	}

	log.Error().Msg("quorum not reached, delete operation failed")
	return ErrQuorumNotReached
}

func (c *Coordinator) findLatest(records []*storage.Record) *storage.Record {
	if len(records) == 0 {
		return nil
	}

	latest := records[0]
	for _, r := range records {
		if hlc.IsAfter(r.Timestamp, latest.Timestamp) {
			latest = r
		}
	}

	return latest
}
