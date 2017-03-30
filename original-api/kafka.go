package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

type ThingEntry struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Foo       int       `json:"foo"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
	Version   int       `json:"version"`

	encoded []byte
	err     error
}

func (te *ThingEntry) ensureEncoded() {
	if te.encoded == nil && te.err == nil {
		te.encoded, te.err = json.Marshal(te)
	}
}

func (te *ThingEntry) Length() int {
	te.ensureEncoded()
	return len(te.encoded)
}

func (te *ThingEntry) Encode() ([]byte, error) {
	te.ensureEncoded()
	return te.encoded, te.err
}

func EntryFromThing(t *Thing) (*ThingEntry, error) {
	te := &ThingEntry{
		ID:        t.ID,
		Name:      t.Name,
		Foo:       t.Foo,
		CreatedOn: t.CreatedOn,
		UpdatedOn: t.UpdatedOn,
		Version:   t.Version,
	}
	te.ensureEncoded()

	return te, te.err
}

type KafkaClient struct {
	client        sarama.Client
	producer      sarama.SyncProducer
	publish_topic string
}

func NewKafkaClient(brokers []string, topic string) (*KafkaClient, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}

	return &KafkaClient{
		client:        client,
		producer:      producer,
		publish_topic: topic,
	}, nil
}

func (c *KafkaClient) Close() error {
	c.producer.Close()
	return c.client.Close()
}

func (c *KafkaClient) PublishStream(s <-chan *Thing) {
	for {
		select {
		case t := <-s:
			e, err := EntryFromThing(t)
			if err != nil {
				log.Printf("failed to convert thing to thing entry: %s", err)
				continue
			}

			partition, offset, err := c.producer.SendMessage(&sarama.ProducerMessage{
				Topic: c.publish_topic,
				Key:   sarama.StringEncoder(strconv.Itoa(t.ID)),
				Value: e,
			})
			if err != nil {
				log.Printf("failed to publish thing (%+v): %v", t, err)
			} else {
				log.Printf("published thing (%+v) at things-%d-%d", t, partition, offset)
			}
		}
	}
}
