package storage

import "errors"

// const
var (
	ErrKeyNotFound = errors.New("Key not found")
)

// Storage interface of Storage.
type Storage interface {
	// Get return the value to the key in Storage.
	Get(key []byte) ([]byte, error)

	// Put put the key-value entry to Storage.
	Put(key []byte, value []byte) error

	// Del delete the key entry in Storage.
	Del(key []byte) error

	// EnableBatch enable batch write.
	EnableBatch()

	// DisableBatch disable batch write.
	DisableBatch()

	// Flush write and flush pending batch write.
	Flush() error
}
