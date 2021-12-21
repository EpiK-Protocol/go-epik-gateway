package storage

import (
	"sync"

	byteutils "github.com/EpiK-Protocol/go-epik-data/utils/bytesutils"
)

// MemoryStorage the nodes in trie.
type MemoryStorage struct {
	data *sync.Map
}

// kv entry
type kv struct{ k, v []byte }

// MemoryBatch do batch task in memory storage
type MemoryBatch struct {
	db      *MemoryStorage
	entries []*kv
}

// NewMemoryStorage init a storage
func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{
		data: new(sync.Map),
	}, nil
}

// Get return value to the key in Storage
func (db *MemoryStorage) Get(key []byte) ([]byte, error) {
	if entry, ok := db.data.Load(byteutils.Hex(key)); ok {
		return entry.([]byte), nil
	}
	return nil, ErrKeyNotFound
}

// Put put the key-value entry to Storage
func (db *MemoryStorage) Put(key []byte, value []byte) error {
	db.data.Store(byteutils.Hex(key), value)
	return nil
}

// Del delete the key in Storage.
func (db *MemoryStorage) Del(key []byte) error {
	db.data.Delete(byteutils.Hex(key))
	return nil
}

// EnableBatch enable batch write.
func (db *MemoryStorage) EnableBatch() {
}

// Flush write and flush pending batch write.
func (db *MemoryStorage) Flush() error {
	return nil
}

// DisableBatch disable batch write.
func (db *MemoryStorage) DisableBatch() {
}
