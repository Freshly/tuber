package db

import (
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
	DBIndexes() (map[string]string, map[string]bool, map[string]int)
	DBRoot() string
	DBUnmarshal(data []byte) (Model, error)
	DBMarshal() ([]byte, error)
	DBKey() string
}

func (d *DB) Save(m Model) error {
	key := m.DBKey()
	if key == "" {
		return fmt.Errorf("save failed, model key nil")
	}

	strings, bools, ints := m.DBIndexes()
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

	marshalled, err := m.DBMarshal()
	if err != nil {
		return err
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.DBRoot()))
		mb, err := rootb.CreateBucketIfNotExists([]byte(m.DBKey()))
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

type Query struct {
	Strings map[string]string
	Bools   map[string]bool
	Ints    map[string]int
}

func (q Query) String(k string, v string) Query {
	q.Strings[k] = v
	return q
}

func (q Query) Bool(k string, v bool) Query {
	q.Bools[k] = v
	return q
}

func (q Query) Int(k string, v int) Query {
	q.Ints[k] = v
	return q
}

type queryResult struct {
	key        string
	marshalled []byte
}

func (d *DB) Get(m Model, query Query) ([]Model, error) {
	normalizedQueryValues, err := validateAndNormalize(m, query)
	if err != nil {
		return nil, err
	}

	var marshalled []queryResult

	err = d.db.View(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.DBRoot()))
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
				marshalled = append(marshalled, queryResult{key: string(k), marshalled: rootbentryb.Get([]byte(marshalledKey))})
			}
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	var results []Model
	for _, marshalledResult := range marshalled {
		result, err := m.DBUnmarshal(marshalledResult.marshalled)
		if err != nil {
			return nil, fmt.Errorf("unmarshal failed for %s/%s/: %v", m.DBRoot(), marshalledResult.key, err)
		}
		results = append(results, result)
	}

	return results, nil
}

type NotFoundError struct {
	err error
}

func (e NotFoundError) Error() string {
	return e.err.Error()
}

func (d *DB) Find(m Model, key string) (Model, error) {
	var marshalled []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		root := m.DBRoot()
		rootb := tx.Bucket([]byte(root))
		rootbentryb := rootb.Bucket([]byte(key))
		if rootbentryb == nil {
			return NotFoundError{err: fmt.Errorf("key %s not found in %s/", key, root)}
		}

		marshalled = rootbentryb.Get([]byte(marshalledKey))
		return nil
	})
	if err != nil {
		return nil, err
	}

	model, err := m.DBUnmarshal(marshalled)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed for %s/%s/: %v", m.DBRoot(), key, err)
	}

	return model, nil
}

func (d *DB) Exists(m Model, key string) bool {
	var found bool
	d.db.View(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.DBRoot()))
		rootbentryb := rootb.Bucket([]byte(key))
		if rootbentryb != nil {
			found = true
		}
		return nil
	})
	return found
}

func (d *DB) Delete(m Model, key string) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		rootb := tx.Bucket([]byte(m.DBRoot()))
		err := rootb.DeleteBucket([]byte(key))
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

func validateAndNormalize(m Model, query Query) (map[string]string, error) {
	var convertedQueryVals map[string]string
	var queryKeys []string

	for k, v := range query.Strings {
		queryKeys = append(queryKeys, k)
		convertedQueryVals[k] = v
	}
	for k, v := range query.Bools {
		queryKeys = append(queryKeys, k)
		convertedQueryVals[k] = strconv.FormatBool(v)
	}
	for k, v := range query.Ints {
		queryKeys = append(queryKeys, k)
		convertedQueryVals[k] = strconv.Itoa(v)
	}

	strings, bools, ints := m.DBIndexes()
	var indexKeys []string
	for k, _ := range strings {
		indexKeys = append(indexKeys, k)
	}
	for k, _ := range bools {
		indexKeys = append(indexKeys, k)
	}
	for k, _ := range ints {
		indexKeys = append(indexKeys, k)
	}

	for _, k := range queryKeys {
		var found bool
		for _, i := range indexKeys {
			if k == i {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("query invalid, %s must be an index", k)
		}
	}
	return convertedQueryVals, nil
}
