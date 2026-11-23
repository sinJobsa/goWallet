package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sinJobsa/goWallet/gw-analytics/internal/config"
	"github.com/sinJobsa/goWallet/gw-analytics/internal/storages"
)

type Consumer struct {
	reader  *kafka.Reader
	storage storages.Storage
	logger  *slog.Logger
}

func New(cfg config.KafkaConfig, storage storages.Storage, logger *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 0, // manual commit
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{reader: reader, storage: storage, logger: logger}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.logger.Info("consumer stopped by context")
				return nil
			}
			c.logger.Error("failed to read message", "error", err)
			continue
		}

		if err := c.processMessage(ctx, msg.Value); err != nil {
			c.logger.Error("failed to process message",
				"error", err,
				"partition", msg.Partition,
				"offset", msg.Offset,
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("failed to commit message",
				"error", err,
				"partition", msg.Partition,
				"offset", msg.Offset,
			)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, payload []byte) error {
	var event eventPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		c.logger.Error("invalid message payload", "error", err)
		return nil
	}

	//вычисляем Latency
	receivedAt := time.Now().UTC()
	latencyMs := receivedAt.Sub(event.Timestamp).Milliseconds()
	if latencyMs < 0 {
		latencyMs = 0
	}

	ev := &storages.Event{
		TransactionID: event.TransactionID,
		UserID:        event.UserID,
		FromCurrency:  event.FromCurrency,
		ToCurrency:    event.ToCurrency,
		Amount:        event.Amount,
		Status:        event.Status,
		CreatedAt:     event.Timestamp,
		ReceivedAt:    receivedAt,
		LatencyMs:     latencyMs,
	}

	if err := c.storage.SaveEvent(ctx, ev); err != nil {
		return fmt.Errorf("save event: %w", err)
	}

	c.logger.Info("event saved",
		"transaction_id", ev.TransactionID,
		"status", ev.Status,
		"latency_ms", ev.LatencyMs,
	)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

type eventPayload struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	FromCurrency  string    `json:"from_currency"`
	ToCurrency    string    `json:"to_currency"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
}
