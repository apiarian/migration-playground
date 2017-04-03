package main

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/boltdb/bolt"
)

var (
	ThingsBucketKey = []byte("THINGS")
)

type BoltThings struct {
	db *bolt.DB
}

func NewBoltThings(path string) (*BoltThings, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(ThingsBucketKey)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &BoltThings{
		db: db,
	}, nil
}

func (bt *BoltThings) CreateThing(name string, foo int) (*Thing, error) {
	if name == "" {
		return nil, errors.New("name must be something")
	}

	if foo == 0 {
		return nil, errors.New("foo must not be zero")
	}

	now := time.Now()
	t := &Thing{
		ID:        0,
		Name:      name,
		Foo:       foo,
		CreatedOn: now,
		UpdatedOn: now,
		Version:   0,
	}

	err := bt.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ThingsBucketKey)
		if b == nil {
			return errors.New("no bucket for things")
		}

		i, err := b.NextSequence()
		if err != nil {
			return err
		}

		t.ID = i

		d, err := json.Marshal(t)
		if err != nil {
			return err
		}

		return b.Put(itob(i), d)
	})

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (bt *BoltThings) UpdateThing(id int, name string, foo int) (*Thing, error) {
	if name == "" {
		return nil, errors.New("name must be something")
	}

	if foo == 0 {
		return nil, errors.New("foo must not be zero")
	}

	var t *Thing

	err := bt.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ThingsBucketKey)
		if b == nil {
			return errors.New("no bucket for things")
		}

	})

}

func itob(i int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

var _ ThingService = &BoltThings{}
