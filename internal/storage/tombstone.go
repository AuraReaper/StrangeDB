package storage

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type TombstoneCollector struct {
	db       *badger.DB
	ttl      time.Duration
	interval time.Duration
	stopcCh  chan struct{}
}

func NewTombstoneCollector(db *badger.DB, ttl, interval time.Duration) *TombstoneCollector {
	return &TombstoneCollector{
		db:       db,
		ttl:      ttl,
		interval: interval,
		stopcCh:  make(chan struct{}),
	}
}

func (tc *TombstoneCollector) Start() {
	go tc.collectLoop()
}

func (tc *TombstoneCollector) Stop() {
	close(tc.stopcCh)
}

func (tc *TombstoneCollector) collectLoop() {
	ticker := time.NewTimer(tc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tc.collect()
		case <-tc.stopcCh:
			return
		}
	}
}

func (tc *TombstoneCollector) collect() {
	now := time.Now()
	threshold := now.Add(-tc.ttl).UnixNano()

	keyToDelete := [][]byte{}

	tc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(dataPrefix)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var record Record
				if err := json.Unmarshal(val, &record); err != nil {
					return nil
				}

				if record.Tombstone && record.Timestamp.WallTime < threshold {
					keyToDelete = append(keyToDelete, item.KeyCopy(nil))
				}

				return nil
			})

			if err != nil {
				continue
			}
		}

		return nil
	})

	if len(keyToDelete) > 0 {
		tc.db.Update(func(txn *badger.Txn) error {
			for _, key := range keyToDelete {
				txn.Delete(key)
			}

			return nil
		})
	}
}
