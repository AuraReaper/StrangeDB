package storage

import (
	"encoding/json"
	"errors"

	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/dgraph-io/badger/v4"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyDeleted  = errors.New("key deleted")
)

type BadgerStorage struct {
	db      *badger.DB
	dataDir string
}

func NewBadgerStorage(dataDir string) *BadgerStorage {
	return &BadgerStorage{
		dataDir: dataDir,
	}
}

func (s *BadgerStorage) Open() error {
	opts := badger.DefaultOptions(s.dataDir)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *BadgerStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}

	return nil
}

const dataPrefix = "d:"

func dataKey(key string) []byte {
	return []byte(dataPrefix + key)
}

func (s *BadgerStorage) Get(key string) (*Record, error) {
	var record Record

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(dataKey(key))
		if err == badger.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &record)
		})
	})
	if err != nil {
		return nil, err
	}

	if record.Tombstone {
		return nil, ErrKeyDeleted
	}

	return &record, nil
}

func (s *BadgerStorage) Set(record *Record) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(dataKey(record.Key), data)
	})
}

func (s *BadgerStorage) Delete(key string, timestamp hlc.Timestamp) error {
	record := &Record{
		Key:       key,
		Value:     nil,
		Timestamp: timestamp,
		Tombstone: true,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(dataKey(key), data)
	})
}

func (s *BadgerStorage) Exists(key string) (bool, error) {
	var exists bool

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(dataKey(key))
		if err == badger.ErrKeyNotFound {
			exists = false
			return nil
		}
		if err != nil {
			return err
		}

		var record Record
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &record)
		})
		if err != nil {
			return err
		}

		exists = !record.Tombstone
		return nil
	})

	return exists, err
}

func (s *BadgerStorage) GetRaw(key string) (*Record, error) {
	var record Record

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(dataKey(key))
		if err == badger.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &record)
		})
	})

	if err != nil {
		return nil, err
	}

	return &record, err
}
