package main

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

const (
	CreateCommandType string = "create"
	UpdateCommandType string = "update"
)

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
