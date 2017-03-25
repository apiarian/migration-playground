package main

import (
	"sync"

	"github.com/pkg/errors"
)

type StreamThings struct {
	kc         *KafkaClient
	thingCache map[string]*Thing
	mux        *sync.Mutex
}

func NewStreamThings(kc *KafkaClient) *StreamThings {
	return &StreamThings{
		kc:         kc,
		thingCache: make(map[string]*Thing),
		mux:        &sync.Mutex{},
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
	return nil, errors.New("not implemented")
}

func (st *StreamThings) UpdateThing(id, version, name string, foo float64) (*Thing, error) {
	return nil, errors.New("not implemented")
}

func (st *StreamThings) GetThing(id string) (*Thing, error) {
	return nil, errors.New("not implemented")
}

func (st *StreamThings) ListThings() ([]*Thing, error) {
	return nil, errors.New("not implemented")
}

func (st *StreamThings) CheckCommand(cid string) (*Thing, error) {
	return nil, errors.New("not implemented")
}

var _ ThingService = &StreamThings{}
