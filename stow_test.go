package stow

import (
	"encoding/gob"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

type MyType struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
}

func init() {
	gob.Register(MyType{})
}

func testStore(t *testing.T, store *Store) {
	store.Put([]byte("hello"), &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(name MyType) {
		found = true
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if !found {
		t.Errorf("ForEach failed!")
	}

	if err != nil {
		t.Error(err.Error())
	}

	var name MyType
	store.Pull([]byte("hello"), &name)

	if name.FirstName != "Derek" || name.LastName != "Kered" {
		t.Errorf("Unexpected name: %v", name)
	}
}

func TestJSON(t *testing.T) {
	db, err := bolt.Open("my1.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my1.db")
	defer db.Close()

	testStore(t, NewJSONStore(db, []byte("json")))
}

func TestXML(t *testing.T) {
	db, err := bolt.Open("my2.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my2.db")
	defer db.Close()

	testStore(t, NewXMLStore(db, []byte("xml")))
}

func TestGob(t *testing.T) {
	db, err := bolt.Open("my3.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my3.db")
	defer db.Close()

	testStore(t, NewStore(db, []byte("gob")))
}
