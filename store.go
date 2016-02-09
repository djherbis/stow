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
	data = append(data, buf.Bytes()...)
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

// PutKey will store b with key "key". If key is []byte or string it uses the key
// directly. Otherwise, it marshals the given type into bytes using the stores Encoder.
func (s *Store) PutKey(key interface{}, b interface{}) error {
	keyBytes, err := s.toBytes(key)
	if err != nil {
		return err
	}
	return s.Put(keyBytes, b)
}

// Put will store b with key "key". If key is []byte or string it uses the key
// directly. Otherwise, it marshals the given type into bytes using the stores Encoder.
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

// GetKey will retreive b with key "key". If key is []byte or string it uses the key
// directly. Otherwise, it marshals the given type into bytes using the stores Encoder.
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
// do can be a function which takes either: 1 param which will take on each "value"
// or 2 params where the first param is the "key" and the second is the "value".
func (s *Store) ForEach(do interface{}) error {
	fc, err := newFuncCall(s, do)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		objects := tx.Bucket(s.bucket)
		if objects == nil {
			return nil
		}
		return objects.ForEach(fc.call)
	})
}

// DeleteAll empties the store
func (s *Store) DeleteAll() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(s.bucket)
	})
}

type funcCall struct {
	s *Store

	Value reflect.Value
	Type  reflect.Type

	hasKey  bool
	keyType reflect.Type

	valType reflect.Type
}

func newFuncCall(s *Store, fn interface{}) (fc funcCall, err error) {
	fc.s = s
	fc.Value = reflect.ValueOf(fn)
	fc.Type = fc.Value.Type()
	if fc.Value.Kind() != reflect.Func {
		return fc, fmt.Errorf("fn is not a func()")
	}

	if fc.Type.NumIn() == 1 {
		fc.setValue(fc.Type.In(0))
	} else if fc.Type.NumIn() == 2 {
		fc.setKey(fc.Type.In(0))
		fc.setValue(fc.Type.In(1))
	} else {
		return fc, fmt.Errorf("bad number of args in ForEach fn.")
	}

	return fc, nil
}

func isPtr(typ reflect.Type) bool { return typ.Kind() == reflect.Ptr }

func (fc *funcCall) setValue(typ reflect.Type) {
	fc.valType = typ
	if isPtr(fc.valType) {
		fc.valType = fc.valType.Elem()
	}
}

func (fc *funcCall) getKey(v []byte) (key reflect.Value, err error) {
	if fc.keyType.Kind() == reflect.String {
		return reflect.ValueOf(string(v)), nil
	} else if fc.keyType.Kind() == reflect.Slice && fc.keyType.Elem().Kind() == reflect.Uint8 {
		return reflect.ValueOf(v), nil
	}

	key = reflect.New(fc.valType)

	if err := fc.s.unmarshal(v, key.Interface()); err != nil {
		return key, err
	}

	if !isPtr(fc.keyType) {
		key = deref(key)
	}

	return key, err
}

func (fc *funcCall) getValue(v []byte) (val reflect.Value, err error) {
	val = reflect.New(fc.valType)

	if err := fc.s.unmarshal(v, val.Interface()); err != nil {
		return val, err
	}

	if !isPtr(fc.valType) {
		val = deref(val)
	}

	return val, err
}

func (fc *funcCall) setKey(typ reflect.Type) {
	fc.hasKey = true
	fc.keyType = typ
	isPtr := fc.keyType.Kind() == reflect.Ptr
	if isPtr {
		fc.keyType = fc.keyType.Elem()
	}
}

func (fc *funcCall) call(k, v []byte) error {
	val, err := fc.getValue(v)
	if err != nil {
		return err
	}

	if !fc.hasKey {
		fc.Value.Call([]reflect.Value{val})
		return nil
	}

	key, err := fc.getKey(k)
	if err != nil {
		return err
	}
	fc.Value.Call([]reflect.Value{key, val})
	return nil
}

func deref(val reflect.Value) reflect.Value {
	if val.IsValid() {
		return val.Elem()
	}
	return reflect.Zero(val.Type())
}
