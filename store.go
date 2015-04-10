// Package stow is used to persist objects to a bolt.DB database.
package stow

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
)

// ErrNotFound indicates object is not in database.
var ErrNotFound = errors.New("not found")

// Store manages objects persistance.
type Store struct {
	db     *bolt.DB
	bucket []byte
	codec  Codec
}

// NewStore creates a new Store, using the underlying
// bolt.DB "bucket" to persist objects.
// NewStore uses GobEncoding, your objects must be registered
// via gob.Register() for this encoding to work.
func NewStore(db *bolt.DB, bucket []byte) *Store {
	return NewCustomStore(db, bucket, GobCodec{})
}

// NewJSONStore creates a new Store, using the underlying
// bolt.DB "bucket" to persist objects as json.
func NewJSONStore(db *bolt.DB, bucket []byte) *Store {
	return NewCustomStore(db, bucket, JSONCodec{})
}

// NewXMLStore creates a new Store, using the underlying
// bolt.DB "bucket" to persist objects as xml.
func NewXMLStore(db *bolt.DB, bucket []byte) *Store {
	return NewCustomStore(db, bucket, XMLCodec{})
}

// NewCustomStore allows you to create a store with
// a custom underlying Encoding
func NewCustomStore(db *bolt.DB, bucket []byte, codec Codec) *Store {
	return &Store{db: db, bucket: bucket, codec: codec}
}

// Put will store b with key "key"
func (s *Store) Put(key []byte, b interface{}) error {
	buf := bytes.NewBuffer(nil)
	enc := s.codec.NewEncoder(buf)
	if err := enc.Encode(b); err != nil {
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

// Pull will retreive b with key "key", and removes it from the store.
func (s *Store) Pull(key []byte, b interface{}) error {
	buf := bytes.NewBuffer(nil)
	err := s.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket(s.bucket)
		if objects == nil {
			return ErrNotFound
		}

		data := objects.Get(key)
		if data == nil {
			return ErrNotFound
		}

		buf.Write(data)
		objects.Delete(key)
		return nil
	})

	if err != nil {
		return err
	}

	dec := s.codec.NewDecoder(buf)
	err = dec.Decode(b)

	return err
}

// Get will retreive b with key "key"
func (s *Store) Get(key []byte, b interface{}) error {
	buf := bytes.NewBuffer(nil)
	err := s.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket(s.bucket)
		if objects == nil {
			return ErrNotFound
		}
		data := objects.Get(key)
		if data == nil {
			return ErrNotFound
		}
		buf.Write(data)
		return nil
	})

	if err != nil {
		return err
	}

	dec := s.codec.NewDecoder(buf)
	err = dec.Decode(b)

	return err
}

// ForEach will run do on each object in the store.
// do must be a function which takes exactly one parameter, which is
// decodeable.
func (s *Store) ForEach(do interface{}) error {
	fv := reflect.ValueOf(do)
	if fv.Kind() != reflect.Func {
		return fmt.Errorf("do is not a func()")
	}
	ft := fv.Type()
	if ft.NumIn() != 1 {
		return fmt.Errorf("do must take exactly one param")
	}
	argtype := ft.In(0)
	isPtr := argtype.Kind() == reflect.Ptr
	if isPtr {
		argtype = argtype.Elem()
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket(s.bucket)
		if objects == nil {
			return nil
		}

		err := objects.ForEach(func(k, v []byte) error {
			buf := bytes.NewBuffer(v)
			dec := s.codec.NewDecoder(buf)
			i := reflect.New(argtype)
			err := dec.Decode(i.Interface())
			if err == nil {
				if !isPtr {
					if i.IsValid() {
						i = i.Elem()
					} else {
						i = reflect.Zero(ft.In(0))
					}
				}
				fv.Call([]reflect.Value{i})
			}
			return err
		})

		return err
	})
}

// DeleteAll empties the store
func (s *Store) DeleteAll() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(s.bucket)
	})
}
