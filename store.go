// Package stow is used to persist objects to a bolt.DB database.
package stow

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/boltdb/bolt"
)

var pool *sync.Pool = &sync.Pool{
	New: func() interface{} { return bytes.NewBuffer(nil) },
}

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

func (s *Store) marshal(val interface{}) (data []byte, err error) {
	buf := pool.Get().(*bytes.Buffer)
	err = s.codec.NewEncoder(buf).Encode(val)
	data = buf.Bytes()
	buf.Reset()
	pool.Put(buf)

	return data, err
}

func (s *Store) unmarshal(data []byte, val interface{}) (err error) {
	return s.codec.NewDecoder(bytes.NewReader(data)).Decode(val)
}

func (s *Store) toBytes(key interface{}) (keyBytes []byte, err error) {
	switch k := key.(type) {
	case string:
		return []byte(k), nil
	case []byte:
		return k, nil
	default:
		return s.marshal(key)
	}
}

// PutKey will store b with key "key"
func (s *Store) PutKey(key interface{}, b interface{}) error {
	keyBytes, err := s.toBytes(key)
	if err != nil {
		return err
	}
	return s.Put(keyBytes, b)
}

// Put will store b with key "key"
func (s *Store) Put(key []byte, b interface{}) (err error) {
	var data []byte
	data, err = s.marshal(b)

	return s.db.Update(func(tx *bolt.Tx) error {
		objects, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return err
		}
		objects.Put(key, data)
		return nil
	})
}

// PullKey will retreive b with key "key", and removes it from the store.
func (s *Store) PullKey(key interface{}, b interface{}) error {
	keyBytes, err := s.toBytes(key)
	if err != nil {
		return err
	}
	return s.Pull(keyBytes, b)
}

// Pull will retreive b with key "key", and removes it from the store.
func (s *Store) Pull(key []byte, b interface{}) error {
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

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

	return s.unmarshal(buf.Bytes(), b)
}

// GetKey will retreive b with key "key"
func (s *Store) GetKey(key interface{}, b interface{}) error {
	keyBytes, err := s.toBytes(key)
	if err != nil {
		return err
	}
	return s.Get(keyBytes, b)
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

	return s.unmarshal(buf.Bytes(), b)
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
			i := reflect.New(argtype)
			err := s.unmarshal(v, i.Interface())
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
