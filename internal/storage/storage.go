package storage

import "github.com/AuraReaper/strangedb/internal/hlc"

type Record struct {
	Key       string        `json:"key"`
	Value     []byte        `json:"value"`
	Timestamp hlc.Timestamp `json:"timestamp"`
	Tombstone bool          `json:"tombstone"`
}

type Storage interface {
	Open() error
	Close() error
	Get(key string) (*Record, error)
	Set(record *Record) error
	Delete(key string, timestamp hlc.Timestamp) error
	Exists(key string) (bool, error)
}
