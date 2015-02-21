// stow is used to persist buffer.Buffer's to a bolt.DB database.
package stow

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/djherbis/buffer"
)

// ErrNotFound indicates buffer is not in database.
var ErrNotFound = errors.New("buffer not found")

// BufferStore manages buffer.Buffer persistance.
type BufferStore struct {
	db     *bolt.DB
	bucket []byte
}

// NewBufferStore creates a new BufferStore, using the underlying
// bolt.DB "bucket" to persist buffers.
func NewBufferStore(db *bolt.DB, bucket []byte) *BufferStore {
	return &BufferStore{db: db, bucket: bucket}
}

// Put persists buffer.Buffer b, modifying b after Put will not affect the stored buffer.
// This will fail if the buffer is not gob-encodable.
func (s *BufferStore) Put(key []byte, b buffer.Buffer) error {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&b); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		buffers, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return err
		}
		buffers.Put(key, buf.Bytes())
		return nil
	})
}

// Get will retrive (and removes) the buffer stored under "key".
// ErrNotFound indicates that the buffer was not found in the database,
// other errors indicate gob-decoding failures.
func (s *BufferStore) Get(key []byte) (b buffer.Buffer, err error) {

	var data []byte
	err = s.db.Update(func(tx *bolt.Tx) error {
		buffers := tx.Bucket(s.bucket)
		if buffers == nil {
			return ErrNotFound
		}
		data = buffers.Get(key)
		if data == nil {
			return ErrNotFound
		}

		buffers.Delete(key)
		return nil
	})

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&b)

	return b, err
}

// Wipe will call Reset() on every buffer managed by this store,
// and then delete the buffer from the store.
func (s *BufferStore) Wipe() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buffers := tx.Bucket(s.bucket)
		if buffers == nil {
			return nil
		}

		buffers.ForEach(func(k, v []byte) error {
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			var b buffer.Buffer
			if err := dec.Decode(&b); err == nil {
				b.Reset()
			}
			return nil
		})

		tx.DeleteBucket(s.bucket)
		return nil
	})
}
