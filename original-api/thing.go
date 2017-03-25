package main

import (
	"time"
)

type Thing struct {
	ID        int
	Name      string
	Foo       int
	CreatedOn time.Time
	UpdatedOn time.Time
	Version   int
}

type ThingService interface {
	CreateThing(name string, foo int) (*Thing, error)
	UpdateThing(id int, version int, name string, foo int) (*Thing, error)
	GetThing(id int) (*Thing, error)
	ListThings() ([]*Thing, error)
	ThingStream() <-chan *Thing
}
