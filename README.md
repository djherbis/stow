Stow [![GoDoc](https://godoc.org/github.com/djherbis/stow?status.svg)](https://godoc.org/github.com/djherbis/stow)
==========

Usage
------------

This package provides a persistence manager for objects backed by boltdb.

```go
package stow

import (
  "os"
  "testing"

  "github.com/boltdb/bolt"
)

func TestJson(t *testing.T) {

  // Create a boltdb database
  db, err := bolt.Open("my.db", 0600, nil)
  if err != nil {
    t.Error(err.Error())
  }
  defer os.Remove("my.db")
  defer db.Close()


  // Defined our test Type
  type MyType struct {
    FirstName string `json:"first"`
    LastName  string `json:"last"`
  }

  // Create a Json-encoded Store, Xml and Gob are also built-in
  store := NewJsonStore(db, []byte("json"))

  // Store the object
  store.Put([]byte("hello"), &MyType{"Derek", "Kered"})

  // Get the object
  var name MyType
  store.Pull([]byte("hello"), &name)

  // Verify
  if name.FirstName != "Derek" || name.LastName != "Kered" {
    t.Errorf("Unexpected name: %v", name)
  }
}
```

Installation
------------
```sh
go get github.com/djherbis/stow
```
