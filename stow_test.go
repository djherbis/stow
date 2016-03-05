package stow

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

type MyType struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
}

func (t *MyType) String() string {
	return fmt.Sprintf("%s %s", t.FirstName, t.LastName)
}

func init() {
	Register(&MyType{})
	RegisterName("stow.YourType", &YourType{})
}

const stowDbFilename = "stowtest.db"

var db *bolt.DB

func TestMain(m *testing.M) {
	flag.Parse()
	var err error
	db, err = bolt.Open(stowDbFilename, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	result := m.Run()
	db.Close()
	os.Remove(stowDbFilename)
	os.Exit(result)
}

type YourType struct {
	FirstName string `json:"first"`
}

func TestChangeType(t *testing.T) {
	s := NewStore(db, []byte("interface"))

	s.Put([]byte("test"), &YourType{"DJ"})

	var v MyType
	s.Get([]byte("test"), &v)

	if v.String() != "DJ " {
		t.Errorf("unexpected response name %s", v.String())
	}
}

func TestInterfaces(t *testing.T) {
	s := NewStore(db, []byte("interface"))

	var j fmt.Stringer = &MyType{"First", "Last"}
	s.Put([]byte("test"), &j)

	err := s.ForEach(func(str fmt.Stringer) {
		if str.String() != "First Last" {
			t.Errorf("unexpected string %s", str)
		}
	})
	if err != nil {
		t.Error(err.Error())
	}

	var i fmt.Stringer
	err = s.Get([]byte("test"), &i)
	if err != nil {
		t.Error(err.Error())
	} else {
		if i.String() != "First Last" {
			t.Errorf("unexpected string %s", i)
		}
	}
}

func testForEachByteKeys(t testing.TB, store *Store) {
	oKey := []byte("hello")

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key []byte, name MyType) {
		found = true
		if !bytes.Equal(key, oKey) {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEachStringKeys(t testing.TB, store *Store) {
	oKey := "hello"

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key string, name MyType) {
		found = true
		if key != oKey {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEachPtrKeys(t testing.TB, store *Store) {
	oKey := &MyType{FirstName: "D"}

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key *MyType, name MyType) {
		found = true
		if *key != *oKey {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEachKeys(t testing.TB, store *Store) {
	oKey := MyType{FirstName: "D"}

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key MyType, name *MyType) {
		found = true
		if key != oKey {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEach(t testing.TB, store *Store) {
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
}

func testStore(t testing.TB, store *Store) {
	testForEachPtrKeys(t, store)
	store.DeleteAll()

	testForEachKeys(t, store)
	store.DeleteAll()

	testForEachStringKeys(t, store)
	store.DeleteAll()

	testForEachByteKeys(t, store)
	store.DeleteAll()

	var name MyType
	if store.Get("hello", &name) != ErrNotFound {
		t.Errorf("key should not be found.")
	}

	testForEach(t, store)

	store.Get("hello", &name)

	if name.FirstName != "Derek" || name.LastName != "Kered" {
		t.Errorf("Unexpected name: %v", name)
	}

	var name2 MyType
	store.Pull("hello", &name2)

	if name2.FirstName != "Derek" || name2.LastName != "Kered" {
		t.Errorf("Unexpected name2: %v", name2)
	}

	var name3 MyType
	err := store.Pull([]byte("hello"), &name3)
	if err != ErrNotFound {
		t.Errorf("pull failed to remove the name!")
	}

	store.Put([]byte("hello"), &MyType{"Friend", "person"})

	var name5 MyType
	err = store.Get([]byte("hello world"), &name5)
	if err != ErrNotFound {
		t.Errorf("Should have been NotFound!")
	}

	store.Delete("hello")

	var name4 MyType
	err = store.Pull([]byte("hello"), &name4)
	if err != ErrNotFound {
		t.Errorf("Delete failed!")
	}

	if err := store.DeleteAll(); err != nil {
		t.Errorf("DeleteAll should have returned nil err %s", err.Error())
	}

	if err := store.Delete("hello"); err != nil {
		t.Errorf("Delete should have returned nil err %s", err.Error())
	}
}

func TestNestedJSON(t *testing.T) {
	parent := NewJSONStore(db, []byte("json_parent"))
	parent.Put("hello", "world")
	testStore(t, parent.NewNestedStore([]byte("json_child")))
	var worldValue string
	if err := parent.Pull("hello", &worldValue); err != nil || worldValue != "world" {
		t.Error("child actions affected parent!", err, worldValue)
	}
}

func TestJSON(t *testing.T) {
	testStore(t, NewJSONStore(db, []byte("json")))
}

func TestXML(t *testing.T) {
	testStore(t, NewXMLStore(db, []byte("xml")))
}

func TestGob(t *testing.T) {
	testStore(t, NewStore(db, []byte("gob")))
}

func TestFunc(t *testing.T) {
	if _, err := newFuncCall(nil, 1); err == nil {
		t.Errorf("expected bad func error")
	}

	if _, err := newFuncCall(nil, func() {}); err == nil {
		t.Errorf("expected bad # of args func error")
	}
}
