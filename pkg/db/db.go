package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

type DB struct {
	db *bolt.DB
}

const DefaultFilemode = 0666
const DefaultStartupTimeout = time.Second

func NewDefaultDB(path string, roots ...string) (*DB, error) {
	return NewDB(path, DefaultFilemode, nil, DefaultStartupTimeout, roots...)
}

func NewDB(path string, filemode os.FileMode, boltOptions *bolt.Options, startupTimeout time.Duration, roots ...string) (*DB, error) {
	dbChan := make(chan *bolt.DB)
	errChan := make(chan error)
	var db *bolt.DB
	go func(dbChan chan *bolt.DB, errChan chan error) {
		db, err := bolt.Open(path, filemode, boltOptions)
		if err != nil {
			errChan <- err
		}
		dbChan <- db
		return
	}(dbChan, errChan)

	timeoutChan := make(chan bool)
	go func(startupTimeout time.Duration, timeoutChan chan bool) {
		time.Sleep(startupTimeout)
		timeoutChan <- true
	}(startupTimeout, timeoutChan)

	select {
	case dbr := <-dbChan:
		db = dbr
	case dbErr := <-errChan:
		return nil, dbErr
	case <-timeoutChan:
		return nil, fmt.Errorf("timeout opening database - check for other running processes that access the file")
	}

	if db == nil {
		panic("nil db")
	}

	err := db.Update(func(tx *bolt.Tx) error {
		for _, root := range roots {
			_, err := tx.CreateBucketIfNotExists([]byte(root))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() {
	d.db.Close()
}

const marshalledKey = "marshalled"

type Model interface {
	Indexes() (map[string]string, map[string]bool, map[string]int)
	BucketName() string
	Key() string
}

func (d *DB) Save(m Model) error {
	key := m.Key()
	if key == "" {
		return fmt.Errorf("save failed, model key nil")
	}

	strings, bools, ints := m.Indexes()
	var err error

	for k, v := range bools {
		if strings[k] != "" {
			return fmt.Errorf("model misconfigured, index keys overlap across type")
		}
		strings[k] = strconv.FormatBool(v)
	}

	for k, v := range ints {
		if strings[k] != "" {
			return fmt.Errorf("model misconfigured, index keys overlap across type")
		}
		strings[k] = strconv.Itoa(v)
	}

	marshalled, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.BucketName()))
		mb, err := rootb.CreateBucketIfNotExists([]byte(m.Key()))
		if err != nil {
			return err
		}
		err = mb.Put([]byte(marshalledKey), marshalled)
		if err != nil {
			return err
		}
		for k, v := range strings {
			err = mb.Put([]byte(k), []byte(v))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type queryResult struct {
	key        string
	marshalled []byte
}

func (d *DB) Get(bucketName string, results interface{}, query Query) error {
	normalizedQueryValues := query.normalize()

	marshalled := []byte("[")
	i := 0

	if err := d.db.View(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(bucketName))
		_ = rootb.ForEach(func(k, v []byte) error {
			rootbentryb := rootb.Bucket(k)
			if rootbentryb == nil {
				return nil
			}
			var match = true
			for queryKey, queryVal := range normalizedQueryValues {
				storedVal := rootbentryb.Get([]byte(queryKey))
				if string(storedVal) != queryVal {
					match = false
					break
				}
			}
			if match {
				if i != 0 {
					marshalled = append(marshalled, byte(','))
				}
				i++
				marshalled = append(marshalled, rootbentryb.Get([]byte(marshalledKey))...)
			}

			return nil
		})

		return nil
	}); err != nil {
		return err
	}

	marshalled = append(marshalled, byte(']'))

	if err := json.Unmarshal(marshalled, results); err != nil {
		return err
	}

	return nil
}

type NotFoundError struct {
	err error
}

func (e NotFoundError) Error() string {
	return e.err.Error()
}

func (d *DB) Find(bucketName string, key string, result interface{}) error {
	var marshalled []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		root := bucketName
		rootb := tx.Bucket([]byte(root))
		rootbentryb := rootb.Bucket([]byte(key))
		if rootbentryb == nil {
			return NotFoundError{err: fmt.Errorf("key %s not found in %s/", key, root)}
		}

		marshalled = rootbentryb.Get([]byte(marshalledKey))
		return nil
	})
	if err != nil {
		return err
	}

	if err := json.Unmarshal(marshalled, result); err != nil {
		return fmt.Errorf("unmarshal failed for %s/%s/: %v", bucketName, key, err)
	}

	return nil
}

func (d *DB) Exists(m Model) bool {
	var found bool
	d.db.View(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.BucketName()))
		rootbentryb := rootb.Bucket([]byte(m.Key()))
		if rootbentryb != nil {
			found = true
		}
		return nil
	})
	return found
}

func (d *DB) Delete(m Model) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.BucketName()))
		err := rootb.DeleteBucket([]byte(m.Key()))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
