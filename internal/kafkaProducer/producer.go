package kafkaProducer

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"strings"
	"time"
)

const (
	flushTimeout = 5000 // ms
)

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(address []string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"message.timeout.ms": 5000,
		"acks":               "all",               // Ждём подтверждения от всех реплик
		"retries":            3,                   // Количество попыток повтора
		"partitioner":        "consistent_random", // Стратегия партиционирования
	})
	if err != nil {
		return nil, fmt.Errorf("error with new producer: %w", err)
	}

	return &Producer{producer: p}, nil
}

func (p *Producer) Produce(message []byte, topic string) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value:     message,
		Key:       nil,
		Timestamp: time.Now(),
	}

	kafkaChan := make(chan kafka.Event)
	if err := p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("error with produce: %w", err)
	}

	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return fmt.Errorf("error with produce: %w", ev)
	default:
		return errors.New("unknown event type")
	}
}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}
