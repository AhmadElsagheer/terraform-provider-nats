package nats

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Client interface {
	GetStream(streamName string) (StreamInfo, error)
	CreateStream(streamConfig StreamConfig) (StreamInfo, error)
	UpdateStream(streamConfig StreamConfig) (StreamInfo, error)
	DeleteStream(streamName string) error

	GetConsumer(streamName, consumerName string) (ConsumerInfo, error)
	CreateConsumer(streamName string, consumerConfig ConsumerConfig) (ConsumerInfo, error)
	UpdateConsumer(streamName string, consumerConfig ConsumerConfig) (ConsumerInfo, error)
	DeleteConsumer(streamName, consumerName string) error
}

type client struct {
	url string
}

// NewClient returns a new nats client.
func NewClient(url string) Client {
	return &client{url: url}
}

func (c *client) GetStream(streamName string) (StreamInfo, error) {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	info, err := js.StreamInfo(streamName)
	if err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) || errors.Is(err, nats.ErrConsumerNotFound) {
			return StreamInfo{}, ErrNotFound
		}
		return StreamInfo{}, fmt.Errorf("failed to retrieve stream info: %w", err)
	}
	return StreamInfo(*info), nil
}

func (c *client) CreateStream(streamConfig StreamConfig) (StreamInfo, error) {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	cfg := nats.StreamConfig(streamConfig)
	info, err := js.AddStream(&cfg)
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to create stream: %w", err)
	}
	return StreamInfo(*info), nil
}

func (c *client) UpdateStream(streamConfig StreamConfig) (StreamInfo, error) {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	cfg := nats.StreamConfig(streamConfig)
	info, err := js.UpdateStream(&cfg)
	if err != nil {
		return StreamInfo{}, fmt.Errorf("failed to update stream: %w", err)
	}
	return StreamInfo(*info), nil
}

func (c *client) DeleteStream(streamName string) error {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	err = js.DeleteStream(streamName)
	if err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}
	return nil
}

func (c *client) GetConsumer(streamName, consumerName string) (ConsumerInfo, error) {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	info, err := js.ConsumerInfo(streamName, consumerName)
	if err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) || errors.Is(err, nats.ErrConsumerNotFound) {
			return ConsumerInfo{}, ErrNotFound
		}
		return ConsumerInfo{}, fmt.Errorf("failed to retrieve consumer info: %w", err)
	}
	return ConsumerInfo(*info), nil
}

func (c *client) CreateConsumer(streamName string, consumerConfig ConsumerConfig) (ConsumerInfo, error) {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	cfg := nats.ConsumerConfig(consumerConfig)
	info, err := js.AddConsumer(streamName, &cfg)
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to create consumer: %w", err)
	}
	return ConsumerInfo(*info), nil
}

func (c *client) UpdateConsumer(streamName string, consumerConfig ConsumerConfig) (ConsumerInfo, error) {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	cfg := nats.ConsumerConfig(consumerConfig)
	info, err := js.UpdateConsumer(streamName, &cfg)
	if err != nil {
		return ConsumerInfo{}, fmt.Errorf("failed to update consumer: %w", err)
	}
	return ConsumerInfo(*info), nil
}

func (c *client) DeleteConsumer(streamName, consumerName string) error {
	nc, err := nats.Connect(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to nats: %w", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to create a jetstream context: %w", err)
	}
	err = js.DeleteConsumer(streamName, consumerName)
	if err != nil {
		return fmt.Errorf("failed to delete consumer: %w", err)
	}
	return nil
}
