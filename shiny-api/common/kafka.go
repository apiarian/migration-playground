package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

const (
	CreateCommandType string = "create"
	UpdateCommandType string = "update"
)

type RequestPayload struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Foo       float64   `json:"foo,omitempty"`
	CreatedOn time.Time `json:"created_on,omitempty"`
	UpdatedOn time.Time `json:"updated_on,omitempty"`
	Version   string    `json:"version,omitempty"`
}

type SuccessPayload struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Foo       float64   `json:"foo"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
	Version   string    `json:"version"`
}

type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ThingCommandEntry struct {
	CID     string          `json:"cid"`
	Type    string          `json:"type"`
	Request *RequestPayload `json:"request"`
	Success *SuccessPayload `json:"success"`
	Failure *ErrorPayload   `json:"failure"`

	encoded []byte
	err     error
}

func (te *ThingCommandEntry) ensureEncoded() {
	if te.encoded == nil && te.err == nil {
		te.encoded, te.err = json.Marshal(te)
	}
}

func (te *ThingCommandEntry) Length() int {
	te.ensureEncoded()
	return len(te.encoded)
}

func (te *ThingCommandEntry) Encode() ([]byte, error) {
	te.ensureEncoded()
	return te.encoded, te.err
}

func (te *ThingCommandEntry) String() string {
	if te == nil {
		return "nil"
	}

	return fmt.Sprintf(
		"cid:'%s' type:%s request:%+v success:%+v failure:%+v",
		te.CID,
		te.Type,
		te.Request,
		te.Success,
		te.Failure,
	)
}

type KafkaClient struct {
	client             sarama.Client
	producer           sarama.SyncProducer
	consumers          []sarama.Consumer
	partitionConsumers []sarama.PartitionConsumer
	command_topic      string
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
		command_topic:      topic,
	}, nil
}

func (c *KafkaClient) Close() error {
	for _, x := range c.consumers {
		x.Close()
	}
	c.producer.Close()
	return c.client.Close()
}

func (c *KafkaClient) actuallySendMessage(cid string, tce *ThingCommandEntry) error {
	partition, offset, err := c.producer.SendMessage(&sarama.ProducerMessage{
		Topic: c.command_topic,
		Key:   sarama.StringEncoder(cid),
		Value: tce,
	})
	if err == nil {
		log.Printf("published thing (%s) at %s|%d|%d:%s", tce, c.command_topic, partition, offset, cid)
	}

	return err
}

func (c *KafkaClient) PublishThingRequestCommand(t *Thing, typ string) (string, error) {
	b := make([]byte, 32)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		panic("the random number generator is busted")
	}
	cid := base64.StdEncoding.EncodeToString(b)

	tce := &ThingCommandEntry{
		CID:  cid,
		Type: typ,
		Request: &RequestPayload{
			ID:        t.ID,
			Name:      t.Name,
			Foo:       t.Foo,
			CreatedOn: t.CreatedOn,
			UpdatedOn: t.UpdatedOn,
			Version:   t.Version,
		},
		Success: nil,
		Failure: nil,
	}
	tce.ensureEncoded()
	if tce.err != nil {
		return "", tce.err
	}

	return cid, c.actuallySendMessage(cid, tce)
}

func (c *KafkaClient) PublishThingSuccessCommand(cid string, t *Thing, typ string) error {
	tce := &ThingCommandEntry{
		CID:     cid,
		Type:    typ,
		Request: nil,
		Success: &SuccessPayload{
			ID:        t.ID,
			Name:      t.Name,
			Foo:       t.Foo,
			CreatedOn: t.CreatedOn,
			UpdatedOn: t.UpdatedOn,
			Version:   t.Version,
		},
		Failure: nil,
	}
	tce.ensureEncoded()
	if tce.err != nil {
		return tce.err
	}

	return c.actuallySendMessage(cid, tce)
}

func (c *KafkaClient) PublishThingErrorCommand(cid string, code int, message, typ string) error {
	tce := &ThingCommandEntry{
		CID:     cid,
		Type:    typ,
		Request: nil,
		Success: nil,
		Failure: &ErrorPayload{
			Code:    code,
			Message: message,
		},
	}
	tce.ensureEncoded()
	if tce.err != nil {
		return tce.err
	}

	return c.actuallySendMessage(cid, tce)
}

func (c *KafkaClient) RegisterMessageProcessor(
	topic string,
	timeout time.Duration,
	done <-chan struct{},
	processor func(*sarama.ConsumerMessage),
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
					processor(msg)

				case <-done:
					return
				}
			}
		}(pcons)
	}

	return nil
}

func WrapThingCommandEntryProcessor(
	processor func(*ThingCommandEntry, error),
) func(*sarama.ConsumerMessage) {
	return func(m *sarama.ConsumerMessage) {
		log.Printf(
			"unwrapping message %s|%d|%d:%s (%s): %s\n",
			m.Topic,
			m.Partition,
			m.Offset,
			m.Key,
			m.Timestamp,
			m.Value,
		)

		var tce *ThingCommandEntry
		err := json.Unmarshal(m.Value, &tce)
		processor(tce, err)
	}
}
