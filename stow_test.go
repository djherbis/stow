package stow

import (
	"fmt"
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

type YourType struct {
	FirstName string `json:"first"`
}

func TestChangeType(t *testing.T) {
	db, err := bolt.Open("my4.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my4.db")
	defer db.Close()

	s := NewStore(db, []byte("interface"))

	s.Put([]byte("test"), &YourType{"DJ"})

	var v MyType
	s.Get([]byte("test"), &v)

	if v.String() != "DJ " {
		t.Errorf("unexpected response name %s", v.String())
	}
}

func TestInterfaces(t *testing.T) {
	db, err := bolt.Open("my4.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my4.db")
	defer db.Close()

	s := NewStore(db, []byte("interface"))

	var j fmt.Stringer = &MyType{"First", "Last"}
	s.Put([]byte("test"), &j)

	err = s.ForEach(func(str fmt.Stringer) {
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
	store.Get([]byte("hello"), &name)

	if name.FirstName != "Derek" || name.LastName != "Kered" {
		t.Errorf("Unexpected name: %v", name)
	}

	var name2 MyType
	store.Pull([]byte("hello"), &name2)

	if name2.FirstName != "Derek" || name2.LastName != "Kered" {
		t.Errorf("Unexpected name2: %v", name2)
	}

	var name3 MyType
	err = store.Pull([]byte("hello"), &name3)
	if err != ErrNotFound {
		t.Errorf("pull failed to remove the name!")
	}

	store.Put([]byte("hello"), &MyType{"Friend", "person"})
	store.DeleteAll()

	var name4 MyType
	err = store.Pull([]byte("hello"), &name4)
	if err != ErrNotFound {
		t.Errorf("DeleteAll failed!")
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
