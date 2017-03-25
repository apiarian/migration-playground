package main

import (
	"sort"
	"sync"
	"time"

	"net/http"

	"github.com/pkg/errors"
)

type MemoryThings struct {
	store  map[int]*Thing
	nextId int
	mux    *sync.Mutex
	stream chan *Thing
}

func NewMemoryThings() *MemoryThings {
	return &MemoryThings{
		store:  make(map[int]*Thing),
		nextId: 0,
		mux:    &sync.Mutex{},
		stream: make(chan *Thing),
	}
}

type codedError struct {
	error
	code int
}

func NewCodedError(err error, code int) codedError {
	return codedError{
		error: err,
		code:  code,
	}
}

func (ce codedError) Code() int {
	return ce.code
}

func (mt *MemoryThings) CreateThing(name string, foo int) (*Thing, error) {
	if name == "" {
		return nil, errors.New("name must be something")
	}

	if foo == 0 {
		return nil, errors.New("foo must not be zero")
	}

	mt.mux.Lock()
	defer mt.mux.Unlock()

	now := time.Now()
	t := &Thing{
		ID:        mt.nextId,
		Name:      name,
		Foo:       foo,
		CreatedOn: now,
		UpdatedOn: now,
		Version:   0,
	}
	mt.store[t.ID] = t

	go func(c chan<- *Thing) { c <- t }(mt.stream)

	mt.nextId = mt.nextId + 1

	return t, nil
}

func (mt *MemoryThings) UpdateThing(id int, version int, name string, foo int) (*Thing, error) {
	if name == "" {
		return nil, errors.New("name must be something")
	}

	if foo == 0 {
		return nil, errors.New("foo must not be zero")
	}

	mt.mux.Lock()
	defer mt.mux.Unlock()

	t, ok := mt.store[id]
	if !ok {
		return nil, NewCodedError(errors.New("not found"), http.StatusNotFound)
	}

	if t.Version != version {
		return nil, NewCodedError(errors.New("version conflict"), http.StatusConflict)
	}

	t.Name = name
	t.Foo = foo
	t.Version = t.Version + 1
	t.UpdatedOn = time.Now()

	go func(c chan<- *Thing) { c <- t }(mt.stream)

	return t, nil
}

func (mt *MemoryThings) GetThing(id int) (*Thing, error) {
	mt.mux.Lock()
	defer mt.mux.Unlock()

	t, ok := mt.store[id]
	if !ok {
		return nil, NewCodedError(errors.New("not found"), http.StatusNotFound)
	}

	return t, nil
}

func (mt *MemoryThings) ListThings() ([]*Thing, error) {
	mt.mux.Lock()
	defer mt.mux.Unlock()

	var ts []*Thing
	for _, t := range mt.store {
		ts = append(ts, t)
	}

	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })

	return ts, nil
}

func (mt *MemoryThings) ThingStream() <-chan *Thing {
	return mt.stream
}

func (mt *MemoryThings) Close() {
	close(mt.stream)
}

var _ ThingService = &MemoryThings{}
