package stow

import (
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

func TestJson(t *testing.T) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove("my.db")
	defer db.Close()

	type MyType struct {
		FirstName string `json:"first"`
		LastName  string `json:"last"`
	}

	store := NewJSONStore(db, []byte("json"))
	store.Put([]byte("hello"), &MyType{"Derek", "Kered"})

	var found bool
	err = store.ForEach(func(name MyType) {
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
