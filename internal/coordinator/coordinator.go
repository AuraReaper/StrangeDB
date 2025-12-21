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
}

func New(nodeURL string, ring *ring.ConsistentHashRing, storage storage.Storage, clock *hlc.Clock,
	grpcClient *grpcTransport.Client, replicationN, readQuorum, writeQuorum int) *Coordinator {
	return &Coordinator{
		nodeURL:      nodeURL,
		ring:         ring,
		storage:      storage,
		grpcClient:   grpcClient,
		replicationN: replicationN,
		readQuorum:   readQuorum,
		writeQuorum:  writeQuorum,
	}
}

func (c *Coordinator) Get(ctx context.Context, key string) (*storage.Record, error) {
	replicas := c.ring.GetReplicas(key, c.replicationN)
	if len(replicas) == 0 {
		return nil, ErrNoNodesAvailable
	}

	type result struct {
		record *storage.Record
		err    error
	}

	resultCh := make(chan result, len(replicas))
	var wg sync.WaitGroup

	for _, replica := range replicas {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			var r *storage.Record
			var err error

			if addr == c.nodeURL {
				// local read
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

			resultCh <- result{
				record: r,
				err:    err,
			}
		}(replica)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var records []*storage.Record
	var successCount int

	for res := range resultCh {
		if res.err == nil {
			records = append(records, res.record)
			successCount++
		}

		if successCount > c.readQuorum {
			return c.findLatest(records), nil
		}
	}

	if successCount > 0 {
		return c.findLatest(records), nil
	}

	return nil, ErrQuorumNotReached
}

func (c *Coordinator) Set(ctx context.Context, key string, value []byte) (*storage.Record, error) {
	replicas := c.ring.GetReplicas(key, c.replicationN)
	if len(replicas) == 0 {
		return nil, ErrNoNodesAvailable
	}

	ts := c.clock.Now()
	record := &storage.Record{
		Key:       key,
		Value:     value,
		Timestamp: ts,
		Tombstone: false,
	}

	successCh := make(chan bool, len(replicas))
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

			successCh <- (err == nil)
		}(replica)
	}

	go func() {
		wg.Wait()
		close(successCh)
	}()

	successCount := 0
	for success := range successCh {
		if success {
			successCount++
		}

		if successCount > c.writeQuorum {
			return record, nil
		}
	}

	if successCount > 0 {
		return record, nil
	}

	return nil, ErrQuorumNotReached
}

func (c *Coordinator) Delete(ctx context.Context, key string) error {
	replicas := c.ring.GetReplicas(key, c.replicationN)
	if len(replicas) == 0 {
		return ErrNoNodesAvailable
	}

	ts := c.clock.Now()

	successCh := make(chan bool, len(replicas))
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

			successCh <- (err == nil)
		}(replica)
	}

	go func() {
		wg.Wait()
		close(successCh)
	}()

	successCount := 0
	for success := range successCh {
		if success {
			successCount++
		}

		if successCount > c.writeQuorum {
			return nil
		}
	}

	if successCount > 0 {
		return nil
	}

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
