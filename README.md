Stow 
==========

[![GoDoc](https://godoc.org/github.com/djherbis/stow?status.svg)](https://godoc.org/github.com/djherbis/stow/v4)
[![Release](https://img.shields.io/github/release/djherbis/stow.svg)](https://github.com/djherbis/stow/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.txt)
[![go test](https://github.com/djherbis/stow/actions/workflows/go-test.yml/badge.svg)](https://github.com/djherbis/stow/actions/workflows/go-test.yml)
[![Coverage Status](https://coveralls.io/repos/djherbis/stow/badge.svg?branch=master)](https://coveralls.io/r/djherbis/stow?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/djherbis/stow)](https://goreportcard.com/report/github.com/djherbis/stow)

Usage
------------

This package provides a persistence manager for objects backed by [bbolt (orig. boltdb)](https://github.com/etcd-io/bbolt).

```go
package main

import (
  "encoding/gob"
  "fmt"
  "log"

  bolt "go.etcd.io/bbolt"
  "github.com/djherbis/stow/v4"
)

func main() {
  // Create a boltdb (bbolt fork) database
  db, err := bolt.Open("my.db", 0600, nil)
  if err != nil {
    log.Fatal(err)
  }

  // Open/Create a Json-encoded Store, Xml and Gob are also built-in
  // We'll store a greeting and person in a boltdb bucket named "people"
  peopleStore := stow.NewJSONStore(db, []byte("people"))

  peopleStore.Put("hello", Person{Name: "Dustin"})

  peopleStore.ForEach(func(greeting string, person Person) {
    fmt.Println(greeting, person.Name)
  })

  // Open/Create a Gob-encoded Store. The Gob encoding keeps type information,
  // so you can encode/decode interfaces!
  sayerStore := stow.NewStore(db, []byte("greetings"))

  var sayer Sayer = Person{Name: "Dustin"}
  sayerStore.Put("hello", &sayer)

  var retSayer Sayer
  sayerStore.Get("hello", &retSayer)
  retSayer.Say("hello")

  sayerStore.ForEach(func(sayer Sayer) {
    sayer.Say("hey")
  })
}

type Sayer interface {
  Say(something string)
}

type Person struct {
  Name string
}

func (p Person) Say(greeting string) {
  fmt.Printf("%s says %s.\n", p.Name, greeting)
}

func init() {
  gob.Register(&Person{})
}

```

Installation
------------
```sh
go get github.com/djherbis/stow/v4
```
