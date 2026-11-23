package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
	logger *slog.Logger
}

func NewProducer(brokers []string, topic string, logger *slog.Logger) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			MaxAttempts:  3,
			BatchTimeout: 10 * time.Millisecond,
			Async:        false,
		},
		topic:  topic,
		logger: logger,
	}
}

func (p *Producer) Produce(ctx context.Context, event Event) error {
	raw, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.TransactionID),
		Value: raw,
		Time:  time.Now(),
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("failed to publish event",
			"topic", p.topic,
			"transaction_id", event.TransactionID,
			"error", err,
		)
		return fmt.Errorf("could not write messages: %w", err)
	}

	p.logger.Info("event published",
		"topic", p.topic,
		"transaction_id", event.TransactionID,
		"amount", event.Amount,
		"currency", event.FromCurrency,
	)
	
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
