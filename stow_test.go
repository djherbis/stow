package stow

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/djherbis/buffer"
)

func TestBufferPoolStore(t *testing.T) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my.db")
	defer db.Close()

	store := NewBufferPoolStore(db, []byte("pools"))

	key := []byte("hello")

	pool := buffer.NewFilePool(10, ".")
	b := pool.Get()
	b.Write([]byte("hello world"))

	store.Put(key, pool)

	p, err := store.Get(key)
	if err != nil {
		t.Error(err.Error())
	}

	p.Put(b)

	if b.Len() > 0 {
		t.Errorf("buffer should be empty")
	}
}

func TestBufferStore(t *testing.T) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my.db")
	defer db.Close()

	store := NewBufferStore(db, []byte("buckets"))

	key := []byte("hello")
	input := []byte("hello world")

	b := buffer.NewPartition(buffer.NewFilePool(10, "."))
	b.Write(input)

	store.Put(key, b)
	buf, err := store.Get(key)
	if err != nil {
		t.Error(err.Error())
	}

	data, err := ioutil.ReadAll(buf)

	if !bytes.Equal(data, input) {
		t.Errorf("expected %s, got %s", input, data)
	}

	store.DeleteAll()
}
