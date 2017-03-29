package main

import (
	"net/http"
	"sort"
	"sync"

	"github.com/pkg/errors"
)

type StreamThings struct {
	kc         *KafkaClient
	u          *Updater
	thingCache map[string]*Thing
	mux        *sync.Mutex
}

func NewStreamThings(kc *KafkaClient, u *Updater) *StreamThings {
	return &StreamThings{
		kc:         kc,
		thingCache: make(map[string]*Thing),
		mux:        &sync.Mutex{},
		u:          u,
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

func (st *StreamThings) CreateThing(name string, foo float64) (*Thing, error) {
	if name == "" {
		return nil, errors.New("name must be something")
	}

	if foo == 0.0 {
		return nil, errors.New("foo must not be zero")
	}

	t, err := st.u.CreateThing(name, foo)
	if err != nil {
		return nil, err
	}

	st.mux.Lock()
	defer st.mux.Unlock()
	st.thingCache[t.ID] = t

	return t, nil
}

func (st *StreamThings) UpdateThing(id, version, name string, foo float64) (*Thing, error) {
	t, err := st.u.UpdateThing(id, version, name, foo)
	if err != nil {
		return nil, err
	}

	st.mux.Lock()
	defer st.mux.Unlock()
	st.thingCache[t.ID] = t

	return t, nil
}

func (st *StreamThings) GetThing(id string) (*Thing, error) {
	st.mux.Lock()
	defer st.mux.Unlock()

	t, exists := st.thingCache[id]
	if !exists {
		return nil, NewCodedError(errors.Errorf("no Thing with id %s", id), http.StatusNotFound)
	}

	return t, nil
}

func (st *StreamThings) ListThings() ([]*Thing, error) {
	st.mux.Lock()
	defer st.mux.Unlock()

	var ts []*Thing
	for _, t := range st.thingCache {
		ts = append(ts, t)
	}

	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })

	return ts, nil
}

func (st *StreamThings) CheckCommand(cid string) (*Thing, error) {
	return nil, errors.New("not implemented")
}

var _ ThingService = &StreamThings{}
