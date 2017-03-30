package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type ThingEntry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Foo       float64   `json:"foo"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
	Version   string    `json"version"`

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
	client             sarama.Client
	producer           sarama.SyncProducer
	consumers          []sarama.Consumer
	partitionConsumers []sarama.PartitionConsumer
	new_topic          string
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
		client:             client,
		producer:           producer,
		consumers:          make([]sarama.Consumer, 0),
		partitionConsumers: make([]sarama.PartitionConsumer, 0),
		new_topic:          topic,
	}, nil
}

func (c *KafkaClient) Close() error {
	for _, x := range c.consumers {
		x.Close()
	}
	c.producer.Close()
	return c.client.Close()
}

func (c *KafkaClient) RegisterMessageProcessor(
	ctx context.Context,
	topic string,
	timeout time.Duration,
	processor chan<- *sarama.ConsumerMessage,
) error {
	limit := time.Now().Add(timeout)

SearchLoop:
	for {
		err := c.client.RefreshMetadata()
		if err != nil {
			return err
		}

		ts, err := c.client.Topics()
		if err != nil {
			return err
		}

		for _, t := range ts {
			if t == topic {
				break SearchLoop
			}
		}

		if time.Now().After(limit) {
			return errors.New("topic not found before timeout")
		}

		time.Sleep(500 * time.Millisecond)
	}

	cons, err := sarama.NewConsumerFromClient(c.client)
	if err != nil {
		return err
	}

	c.consumers = append(c.consumers, cons)

	ps, err := cons.Partitions(topic)
	if err != nil {
		return err
	}

	for _, part := range ps {
		pcons, err := cons.ConsumePartition(topic, part, sarama.OffsetOldest)
		if err != nil {
			return err
		}

		c.partitionConsumers = append(c.partitionConsumers, pcons)

		go func(p sarama.PartitionConsumer) {
			for {
				select {
				case msg := <-p.Messages():
					processor <- msg

				case <-ctx.Done():
					return
				}
			}
		}(pcons)
	}

	return nil
}

func (c *KafkaClient) PublishThing(t *Thing) error {
	te, err := EntryFromThing(t)
	if err != nil {
		return err
	}

	partition, offset, err := c.producer.SendMessage(&sarama.ProducerMessage{
		Topic: c.new_topic,
		Key:   sarama.StringEncoder(te.ID),
		Value: te,
	})
	if err != nil {
		return err
	}

	log.Printf("published thing %+v at %s|%d|%d", t, c.new_topic, partition, offset)

	return nil
}

func ExtractThingFromMessage(m *sarama.ConsumerMessage) (*Thing, error) {
	var te *ThingEntry
	err := json.Unmarshal(m.Value, &te)
	if err != nil {
		return nil, err
	}

	return &Thing{
		ID:        te.ID,
		Name:      te.Name,
		Foo:       te.Foo,
		CreatedOn: te.CreatedOn,
		UpdatedOn: te.UpdatedOn,
		Version:   te.Version,
	}, nil
}
