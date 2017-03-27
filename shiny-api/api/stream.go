package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/apiarian/migration-playground/shiny-api/common"
	"github.com/pkg/errors"
)

type StreamThings struct {
	kc         *common.KafkaClient
	thingCache map[string]*common.Thing
	mux        *sync.Mutex
}

func NewStreamThings(kc *common.KafkaClient) *StreamThings {
	return &StreamThings{
		kc:         kc,
		thingCache: make(map[string]*common.Thing),
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

func (st *StreamThings) CreateThing(name string, foo float64) (*common.Thing, error) {
	if name == "" {
		return nil, errors.New("name must be something")
	}

	if foo == 0.0 {
		return nil, errors.New("foo must not be zero")
	}

	now := time.Now()
	cid, err := st.kc.PublishThingRequestCommand(&common.Thing{
		ID:        "",
		Name:      name,
		Foo:       foo,
		CreatedOn: now,
		UpdatedOn: now,
		Version:   "0",
	}, common.CreateCommandType)
	if err != nil {
		return nil, NewCodedError(err, http.StatusInternalServerError)
	}

	return nil, NewCodedError(errors.Errorf("sent command '%s'", cid), http.StatusAccepted)
}

func (st *StreamThings) UpdateThing(id, version, name string, foo float64) (*common.Thing, error) {
	return nil, errors.New("not implemented")
}

func (st *StreamThings) GetThing(id string) (*common.Thing, error) {
	return nil, errors.New("not implemented")
}

func (st *StreamThings) ListThings() ([]*common.Thing, error) {
	return nil, errors.New("not implemented")
}

func (st *StreamThings) CheckCommand(cid string) (*common.Thing, error) {
	return nil, errors.New("not implemented")
}

var _ common.ThingService = &StreamThings{}
