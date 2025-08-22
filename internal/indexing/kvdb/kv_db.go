package kvdb

import (
	"os"
	"strings"

	
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
)

type IKeyValueDB interface {
	Open() error                              // Initialize the database, create the directory if it does not exist
	GetDbPath() string                        // Get the path to the database file
	Set(k, v []byte) error                    // Set a key-value pair
	BatchSet(keys, values [][]byte) error     // Batch set, keys and values must be the same length
	Get(k []byte) ([]byte, error)             // Get a value by key
	BatchGet(keys [][]byte) ([][]byte, error) // Batch get, keys must be the same length, order is not guaranteed
	Delete(k []byte) error                    // Delete a key-value pair
	BatchDelete(keys [][]byte) error          // Batch delete, keys must be the same length
	Has(k []byte) bool                        // Check if a key exists
	IterDB(fn func(k, v []byte) error) int64  // Iterate through all key-value pairs, return the number of pairs
	IterKey(fn func(k []byte) error) int64    // Iterate through all keys, return the number of keys
	SetBucket(bucket string) error		 	  // Set the bucket for the database, return the database instance
	Close() error                             // Flush memory to disk and close the database
}

// Factory function to get a key-value database instance based on the type and path
func GetKvDb(dbtype int, path string) (IKeyValueDB, error) {
	paths := strings.Split(path, "/")
	parentPath := strings.Join(paths[0:len(paths)-1], "/")

	info, err := os.Stat(parentPath)
	if os.IsNotExist(err) { // if parentPath does not exist, create it
		logger.Log.Printf("create dir %s", parentPath)
		os.MkdirAll(parentPath, os.ModePerm)
	} else {
		if info.Mode().IsRegular() { // if parentPath is a regular file, delete it
			logger.Log.Printf("%s is a regular file, will delete it", parentPath)
			os.Remove(parentPath)
		}
	}

	var db IKeyValueDB
	switch dbtype {
	case BADGER:
		db = new(Badger).WithDataPath(path)
	default: // default to BoltDB
		db = new(Bolt).WithDataPath(path).WithBucket("radic")
	}
	err = db.Open()
	
	return db, err
}
