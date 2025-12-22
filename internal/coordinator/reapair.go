package coordinator

import (
	"context"
	"sync"

	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/storage"
	pb "github.com/AuraReaper/strangedb/internal/transport/grpc/proto"
)

type ReadRepair struct {
	coordinator *Coordinator
}

func NewReadRepair(coordinator *Coordinator) *ReadRepair {
	return &ReadRepair{
		coordinator: coordinator,
	}
}

type RepairResults struct {
	Latest    *storage.Record
	Stale     map[string]*storage.Record
	Addresses map[string]*storage.Record
}

func (rr *ReadRepair) CheckAndRepair(ctx context.Context, results *RepairResults) {
	if results.Latest == nil || len(results.Stale) == 0 {
		return
	}

	go func() {
		var wg sync.WaitGroup

		for addr := range results.Stale {
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				rr.readReplica(context.Background(), addr, results.Latest)
			}(addr)
		}

		wg.Wait()
	}()
}

func (rr *ReadRepair) readReplica(ctx context.Context, address string, record *storage.Record) error {
	if address == rr.coordinator.nodeURL {
		// local repair
		return rr.coordinator.storage.Set(record)
	}

	// remote repair
	_, err := rr.coordinator.grpcClient.Set(ctx, address, &pb.Record{
		Key:   record.Key,
		Value: record.Value,
		Timestamp: &pb.Timestamp{
			WallTime: record.Timestamp.WallTime,
			Logical:  record.Timestamp.Logical,
			NodeId:   record.Timestamp.NodeID,
		},
		Tombstone: record.Tombstone,
	})

	return err
}

func (rr *ReadRepair) AnalyzeResponses(responses map[string]*storage.Record, latest *storage.Record) *RepairResults {
	results := &RepairResults{
		Latest:    latest,
		Stale:     make(map[string]*storage.Record),
		Addresses: responses,
	}

	if latest == nil {
		return results
	}

	for addr, record := range responses {
		if record == nil {
			// missing on replica
			results.Stale[addr] = nil
		} else if hlc.IsBefore(record.Timestamp, latest.Timestamp) {
			// stale on this replica
			results.Stale[addr] = record
		}
	}

	return results
}
