// stow is used to persist objects to a bolt.DB database.
package stow

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/boltdb/bolt"
)

// ErrNotFound indicates object is not in database.
var ErrNotFound = errors.New("not found")

// Store manages objects persistance.
type Store struct {
	db     *bolt.DB
	bucket []byte
}

// NewStore creates a new Store, using the underlying
// bolt.DB "bucket" to persist objects.
func NewStore(db *bolt.DB, bucket []byte) *Store {
	return &Store{db: db, bucket: bucket}
}

// Put will store b with key "key"
func (s *Store) Put(key []byte, b interface{}) error {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&b); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		objects, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return err
		}
		objects.Put(key, buf.Bytes())
		return nil
	})
}

// Get will retreive b with key "key", and removes it from the store.
func (s *Store) Get(key []byte, b interface{}) error {
	var data []byte
	err := s.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket(s.bucket)
		if objects == nil {
			return ErrNotFound
		}
		data = objects.Get(key)
		if data == nil {
			return ErrNotFound
		}

		objects.Delete(key)
		return nil
	})

	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(b)

	return err
}

// ForEach will run do on each object in the store
func (s *Store) ForEach(do func(i interface{})) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket(s.bucket)
		if objects == nil {
			return nil
		}

		objects.ForEach(func(k, v []byte) error {
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			var i interface{}
			if err := dec.Decode(&i); err == nil {
				do(i)
			}
			return nil
		})
		return nil
	})
}

// DeleteAll empties the store
func (s *Store) DeleteAll() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(s.bucket)
	})
}