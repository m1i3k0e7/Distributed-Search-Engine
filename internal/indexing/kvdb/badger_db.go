package kvdb

import (
	"errors"
	"os"
	"path"
	"sync/atomic"

	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	"github.com/dgraph-io/badger/v4"
)

type Badger struct {
	db   *badger.DB
	path string
}

func (s *Badger) WithDataPath(path string) *Badger {
	s.path = path
	return s
}

func (s *Badger) Open() error {
	DataDir := s.GetDbPath()
	if err := os.MkdirAll(path.Dir(DataDir), os.ModePerm); err != nil {
		return err
	}

	option := badger.DefaultOptions(DataDir).WithNumVersionsToKeep(1).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(option) // Only one instance of badger can be opened at a time, so if you open it again, it will return an error
	if err != nil {
		return err
	} else {
		s.db = db
		return nil
	}
}

func (s *Badger) GetDbPath() string {
	return s.path
}

func (s *Badger) CheckAndGC() {
	lsmSize1, vlogSize1 := s.db.Size()
	for {
		if err := s.db.RunValueLogGC(0.5); err == badger.ErrNoRewrite || err == badger.ErrRejected {
			break
		}
	}

	lsmSize2, vlogSize2 := s.db.Size()
	if vlogSize2 < vlogSize1 {
		logger.Log.Printf("badger before GC, LSM %d, vlog %d. after GC, LSM %d, vlog %d", lsmSize1, vlogSize1, lsmSize2, vlogSize2)
	} else {
		logger.Log.Printf("collect zero garbage")
	}
}

func (s *Badger) Set(k, v []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		//duration := time.Hour * 87600
		return txn.Set(k, v)
	})

	return err
}

func (s *Badger) BatchSet(keys, values [][]byte) error {
	if len(keys) != len(values) {
		return errors.New("key value not the same length")
	}

	var err error
	txn := s.db.NewTransaction(true)
	for i, key := range keys {
		value := values[i]
		//duration := time.Hour * 87600
		//util.util.Log.Debugf("duration",duration)
		if err = txn.Set(key, value); err != nil {
			_ = txn.Commit() // commit old transaction if there is an error, then start a new transaction to retry
			txn = s.db.NewTransaction(true)
			_ = txn.Set(key, value)
		}
	}
	txn.Commit()

	return err
}

func (s *Badger) Get(k []byte) ([]byte, error) {
	var ival []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		//buffer := make([]byte, badgerOptions.ValueLogMaxEntries)
		//ival, err = item.ValueCopy(buffer)
		err = item.Value(func(val []byte) error {
			ival = val
			return nil
		})

		return err
	})

	return ival, err
}

func (s *Badger) BatchGet(keys [][]byte) ([][]byte, error) {
	var err error
	txn := s.db.NewTransaction(false) // read-only transaction
	values := make([][]byte, len(keys))
	for i, key := range keys {
		var item *badger.Item
		item, err = txn.Get(key)
		if err == nil {
			//buffer := make([]byte, badgerOptions.ValueLogMaxEntries)
			var ival []byte
			//ival, err = item.ValueCopy(buffer)
			err = item.Value(func(val []byte) error {
				ival = val
				return nil
			})

			if err == nil {
				values[i] = ival
			} else {
				values[i] = []byte{} // set empty slice if there is an error reading the value
			}
		} else { // Fail to get the item, set an empty slice
			values[i] = []byte{}
			if err != badger.ErrKeyNotFound {
				txn.Discard()
				txn = s.db.NewTransaction(false)
			}
		}
	}
	txn.Discard()

	return values, err
}

func (s *Badger) Delete(k []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(k)
	})

	return err
}

func (s *Badger) BatchDelete(keys [][]byte) error {
	var err error
	txn := s.db.NewTransaction(true)
	for _, key := range keys {
		if err = txn.Delete(key); err != nil {
			_ = txn.Commit() // commit old transaction if there is an error, then start a new transaction to retry
			txn = s.db.NewTransaction(true)
			_ = txn.Delete(key)
		}
	}
	txn.Commit()

	return err
}

func (s *Badger) Has(k []byte) bool {
	var exists = false
	s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(k)
		if err != nil {
			return err
		} else {
			exists = true // if no error, the key exists
		}

		return err
	})

	return exists
}

func (s *Badger) IterDB(fn func(k, v []byte) error) int64 {
	var total int64
	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			var ival []byte
			//var err error
			//buffer := make([]byte, badgerOptions.ValueLogMaxEntries)
			//ival, err = item.ValueCopy(buffer)

			err := item.Value(func(val []byte) error {
				ival = val
				return nil
			})

			if err != nil {
				continue
			}
			if err := fn(key, ival); err == nil {
				atomic.AddInt64(&total, 1)
			}
		}

		return nil
	})

	return atomic.LoadInt64(&total)
}

// IterKey, only iterate keys, not values
func (s *Badger) IterKey(fn func(k []byte) error) int64 {
	var total int64
	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // set to false to only iterate keys
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if err := fn(k); err == nil {
				atomic.AddInt64(&total, 1)
			}
		}

		return nil
	})
	
	return atomic.LoadInt64(&total)
}

// Close, closes the database connection, it is important to call this method to release resources
func (s *Badger) Close() error {
	return s.db.Close()
}
