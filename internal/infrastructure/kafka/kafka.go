package kafka

import (
	"context"
	"fmt"
	"mpc/internal/infrastructure/config"

	"github.com/segmentio/kafka-go"
)

type Writer = kafka.Writer
type Reader = kafka.Reader

type kafkaOptions struct {
	Topic string
}

type KafkaOption func(*kafkaOptions)

func defaultKafkaOptions(cfg *config.Config) kafkaOptions {
	return kafkaOptions{
		Topic: cfg.Kafka.Topic,
	}
}

func WithTopic(topic string) KafkaOption {
	return func(o *kafkaOptions) {
		o.Topic = topic
	}
}

func NewKafkaProducer(cfg *config.Config, opts ...KafkaOption) (*Writer, error) {
	options := defaultKafkaOptions(cfg)
	for _, opt := range opts {
		opt(&options)
	}
	err := createTopicIfNotExists(cfg.Kafka.Brokers, options.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}
	producer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  cfg.Kafka.Brokers,
		Topic:    options.Topic,
		Balancer: &kafka.LeastBytes{},
	})
	return producer, nil
}

func NewKafkaConsumer(cfg *config.Config, opts ...KafkaOption) (*kafka.Reader, error) {
	options := defaultKafkaOptions(cfg)
	for _, opt := range opts {
		opt(&options)
	}

	if options.Topic != cfg.Kafka.Topic {
		err := createTopicIfNotExists(cfg.Kafka.Brokers, options.Topic)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic: %w", err)
		}
	}

	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   options.Topic,
	})
	return consumer, nil
}

func CloseKafkaConnections(producer *kafka.Writer, consumers ...*kafka.Reader) error {
	var err error
	if producer != nil {
		if e := producer.Close(); e != nil {
			err = e
		}
	}
	for _, consumer := range consumers {
		if consumer != nil {
			if e := consumer.Close(); e != nil && err == nil {
				err = e
			}
		}
	}
	return err
}

func createTopicIfNotExists(brokers []string, topic string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)

	if err != nil {
		// If the topic already exists, it's not an error
		if err != kafka.TopicAlreadyExists {
			return err
		}
	}

	return nil
}

func PublishMessage(producer *kafka.Writer, key, value []byte) error {
	return producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)
}

func ConsumeMessages(consumer *kafka.Reader, handler func(kafka.Message) error) error {
	for {
		msg, err := consumer.ReadMessage(context.Background())
		if err != nil {
			return err
		}
		if err := handler(msg); err != nil {
			return err
		}
	}
}

// ReadNewMessage reads only new messages published after this function is called
func ReadNewMessage(ctx context.Context, reader *kafka.Reader) (kafka.Message, error) {
	// First, seek to the end of the topic
	err := reader.SetOffset(kafka.LastOffset)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("failed to set offset: %w", err)
	}

	// Now read the next message, which should be a new one
	return reader.ReadMessage(ctx)
}
