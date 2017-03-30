package main

import (
	"time"
)

type Thing struct {
	ID        string
	Name      string
	Foo       float64
	CreatedOn time.Time
	UpdatedOn time.Time
	Version   string
}

type ThingService interface {
	CreateThing(name string, foo float64) (*Thing, error)
	UpdateThing(id, version, name string, foo float64) (*Thing, error)
	GetThing(id string) (*Thing, error)
	ListThings() ([]*Thing, error)
	CheckCommand(cid string) (*Thing, error)
}

func (t *Thing) Clone() *Thing {
	return &Thing{
		ID:        t.ID,
		Name:      t.Name,
		Foo:       t.Foo,
		CreatedOn: t.CreatedOn,
		UpdatedOn: t.UpdatedOn,
		Version:   t.Version,
	}
}
