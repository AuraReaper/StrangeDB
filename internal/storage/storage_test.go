package storage

import (
	"os"
	"testing"

	"github.com/AuraReaper/strangedb/internal/hlc"
)

func setupTestStorage(t *testing.T) *BadgerStorage {
	dir, err := os.MkdirTemp("", "strangedb-test-*")
	if err != nil {
		t.Fatal(err)
	}

	storage := NewBadgerStorage(dir)
	if err := storage.Open(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		storage.Close()
		os.RemoveAll(dir)
	})

	return storage
}

func TestSetAndGet(t *testing.T) {
	storage := setupTestStorage(t)
	clock := hlc.NewClock("test-node")

	record := &Record{
		Key:       "test-key",
		Value:     []byte("test-value"),
		Timestamp: clock.Now(),
		Tombstone: false,
	}

	err := storage.Set(record)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := storage.Get("test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved.Value) != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", string(retrieved.Value))
	}
}

func TestDelete(t *testing.T) {
	storage := setupTestStorage(t)
	clock := hlc.NewClock("test-node")

	record := &Record{
		Key:       "delete-me",
		Value:     []byte("value"),
		Timestamp: clock.Now(),
		Tombstone: false,
	}
	storage.Set(record)

	err := storage.Delete("delete-me", clock.Now())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Get("delete-me")
	if err != ErrKeyDeleted {
		t.Errorf("Expected ErrKeyDeleted, got %v", err)
	}
}

func TestKeyNotFound(t *testing.T) {
	storage := setupTestStorage(t)

	_, err := storage.Get("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}
