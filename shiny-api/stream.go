package main

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type StreamThings struct {
	kc         *KafkaClient
	u          *Updater
	thingCache map[string]*Thing
	mux        *sync.Mutex
	topic      string
}

func NewStreamThings(kc *KafkaClient, u *Updater, topic string) *StreamThings {
	return &StreamThings{
		kc:         kc,
		thingCache: make(map[string]*Thing),
		mux:        &sync.Mutex{},
		u:          u,
		topic:      topic,
	}
}

func (st *StreamThings) Start() <-chan error {
	errs := make(chan error)

	messages := make(chan *sarama.ConsumerMessage)

	go func(c <-chan *sarama.ConsumerMessage) {
		for cm := range c {
			t, err := ExtractThingFromMessage(cm)
			if err != nil {
				log.Printf(
					"trouble with message %s|%d|%d:%s (%s): %s: %v",
					cm.Topic,
					cm.Partition,
					cm.Offset,
					cm.Key,
					cm.Timestamp,
					cm.Value,
					err,
				)
				continue
			}

			err = st.HandleThingFromMessage(t)
			if err != nil {
				log.Printf("error handling thing %+v: %s", t, err)
			}
		}
	}(messages)

	go func() {
		err := st.kc.RegisterMessageProcessor(
			context.Background(),
			st.topic,
			5*time.Minute,
			messages,
		)

		if err != nil {
			errs <- err
		} else {
			log.Printf("stream message processor registered")
		}
	}()

	return errs
}

func (st *StreamThings) HandleThingFromMessage(t *Thing) error {
	st.mux.Lock()
	defer st.mux.Unlock()

	st.thingCache[t.ID] = t.Clone()

	return nil
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
	st.thingCache[t.ID] = t.Clone()

	return t, nil
}

func (st *StreamThings) UpdateThing(id, version, name string, foo float64) (*Thing, error) {
	t, err := st.u.UpdateThing(id, version, name, foo)
	if err != nil {
		return nil, err
	}

	st.mux.Lock()
	defer st.mux.Unlock()
	st.thingCache[t.ID] = t.Clone()

	return t, nil
}

func (st *StreamThings) GetThing(id string) (*Thing, error) {
	st.mux.Lock()
	defer st.mux.Unlock()

	t, exists := st.thingCache[id]
	if !exists {
		return nil, NewCodedError(errors.Errorf("no Thing with id %s", id), http.StatusNotFound)
	}

	return t.Clone(), nil
}

func (st *StreamThings) ListThings() ([]*Thing, error) {
	st.mux.Lock()
	defer st.mux.Unlock()

	var ts []*Thing
	for _, t := range st.thingCache {
		ts = append(ts, t.Clone())
	}

	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })

	return ts, nil
}

func (st *StreamThings) CheckCommand(cid string) (*Thing, error) {
	return nil, errors.New("not implemented")
}

var _ ThingService = &StreamThings{}
