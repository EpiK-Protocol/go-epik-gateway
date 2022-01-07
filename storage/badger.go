package storage

import (
	"sync"

	byteutils "github.com/EpiK-Protocol/go-epik-gateway/utils/bytesutils"
	"golang.org/x/xerrors"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
)

type BadgerStorage struct {
	db          *badger.DB
	enableBatch bool
	mutex       sync.Mutex
	batchOpts   map[string]*batchOpt
}

type batchOpt struct {
	key     []byte
	value   []byte
	deleted bool
}

// NewDiskStorage init a storage
func NewBadgerStorage(path string) (*BadgerStorage, error) {
	if len(path) == 0 {
		return nil, xerrors.Errorf("need storage path.")
	}
	opts := badger.DefaultOptions("").
		WithNumVersionsToKeep(1).
		WithSyncWrites(true).
		WithTruncate(true).
		WithValueLogLoadingMode(options.FileIO).
		WithNumMemtables(1).
		WithNumLevelZeroTables(1).
		WithNumLevelZeroTablesStall(2).
		WithTableLoadingMode(options.FileIO).
		WithMaxTableSize(10 * 1024 * 1024).
		WithValueLogFileSize(20 * 1024 * 1024)

	db, err := badger.Open(opts.WithDir(path).WithValueDir(path))
	if err != nil {
		return nil, err
	}

	return &BadgerStorage{
		db:          db,
		enableBatch: false,
		batchOpts:   make(map[string]*batchOpt),
	}, nil
}

// Get return value to the key in Storage
func (storage *BadgerStorage) Get(key []byte) ([]byte, error) {
	var value []byte
	err := storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrKeyNotFound
			}
			return err
		}
		return item.Value(func(v []byte) (err error) {
			value = v
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Put put the key-value entry to Storage
func (storage *BadgerStorage) Put(key []byte, value []byte) error {
	if storage.enableBatch {
		storage.mutex.Lock()
		defer storage.mutex.Unlock()

		storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
			key:     key,
			value:   value,
			deleted: false,
		}

		return nil
	}

	return storage.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Del delete the key in Storage.
func (storage *BadgerStorage) Del(key []byte) error {
	if storage.enableBatch {
		storage.mutex.Lock()
		defer storage.mutex.Unlock()

		storage.batchOpts[byteutils.Hex(key)] = &batchOpt{
			key:     key,
			deleted: true,
		}

		return nil
	}

	return storage.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close levelDB
func (storage *BadgerStorage) Close() error {
	return storage.db.Close()
}

// EnableBatch enable batch write.
func (storage *BadgerStorage) EnableBatch() {
	storage.enableBatch = true
}

// Flush write and flush pending batch write.
func (storage *BadgerStorage) Flush() error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if !storage.enableBatch {
		return nil
	}

	return storage.db.Update(func(txn *badger.Txn) error {
		for _, opt := range storage.batchOpts {
			if opt.deleted {
				if err := txn.Delete(opt.key); err != nil {
					return err
				}
			} else {
				if err := txn.Set(opt.key, opt.value); err != nil {
					return err
				}
			}
		}
		storage.batchOpts = make(map[string]*batchOpt)
		return nil
	})
}

// DisableBatch disable batch write.
func (storage *BadgerStorage) DisableBatch() {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	storage.batchOpts = make(map[string]*batchOpt)
	storage.enableBatch = false
}
