package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type Updater struct {
	kc          *KafkaClient
	new_topic   string
	done        chan struct{}
	ownsThings  bool
	nextID      int
	mux         *sync.Mutex
	entityCache map[string]*Thing
}

func NewUpdater(
	kc *KafkaClient,
	new_topic string,
	ownsThings bool,
) *Updater {
	return &Updater{
		kc:          kc,
		new_topic:   new_topic,
		ownsThings:  ownsThings,
		nextID:      0,
		mux:         &sync.Mutex{},
		entityCache: make(map[string]*Thing),
	}
}

func (u *Updater) Start() <-chan error {
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

			err = u.HandleThingFromMessage(t)
			if err != nil {
				log.Printf("error handling thing %+v: %s", t, err)
			}
		}
	}(messages)

	go func() {
		err := u.kc.RegisterMessageProcessor(
			context.Background(),
			u.new_topic,
			5*time.Minute,
			messages,
		)

		if err != nil {
			errs <- err
		} else {
			log.Printf("updater message processor registered")
		}
	}()

	return errs
}

func (u *Updater) HandleThingFromMessage(t *Thing) error {
	return errors.New("updater message handler not implmeneted")
}

func (u *Updater) CreateThing(name string, foo float64) (*Thing, error) {
	u.mux.Lock()
	defer u.mux.Unlock()

	var t *Thing

	if u.ownsThings {
		now := time.Now()
		t = &Thing{
			ID:        strconv.Itoa(u.nextID),
			Name:      name,
			Foo:       foo,
			CreatedOn: now,
			UpdatedOn: now,
			Version:   "0",
		}

		_, exists := u.entityCache[t.ID]
		if exists {
			return nil, errors.Errorf("a thing with id %s already exists", t.ID)
		}

		u.nextID = u.nextID + 1

	} else {
		return nil, errors.New("not owning Things isn't supported yet")
	}

	err := u.kc.PublishThing(t)
	if err != nil {
		return nil, err
	}

	u.entityCache[t.ID] = t.Clone()

	return t, nil
}

func (u *Updater) UpdateThing(id, version, name string, foo float64) (*Thing, error) {
	u.mux.Lock()
	defer u.mux.Unlock()

	var t *Thing

	if u.ownsThings {
		var exists bool
		t, exists = u.entityCache[id]
		if !exists {
			return nil, NewCodedError(errors.Errorf("no Thing with id %s", id), http.StatusNotFound)
		}

		if t.Version != version {
			return nil, NewCodedError(errors.New("version conflict"), http.StatusConflict)
		}

		var changed bool

		if name != "" {
			t.Name = name
			changed = true
		}

		if foo != 0 {
			t.Foo = foo
			changed = true
		}

		if changed {
			v, _ := strconv.Atoi(t.Version)
			t.Version = strconv.Itoa(v + 1)
		}

	} else {
		return nil, errors.New("not owning Things isn't supported yet")
	}

	err := u.kc.PublishThing(t)
	if err != nil {
		return nil, err
	}

	u.entityCache[t.ID] = t.Clone()

	return t, nil
}
