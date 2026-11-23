package consumer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sinJobsa/goWallet/gw-analytics/internal/storages"
)

type mockStorage struct {
	saved []*storages.Event
	err   error
}

func (m *mockStorage) SaveEvent(ctx context.Context, e *storages.Event) error {
	if m.err != nil {
		return m.err
	}
	m.saved = append(m.saved, e)
	return nil
}

func (m *mockStorage) Close() error { return nil }

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestProcessMessage_LatencyComputed(t *testing.T) {
	storage := &mockStorage{}
	c := &Consumer{storage: storage, logger: newTestLogger()}

	ts := time.Now().UTC().Add(-150 * time.Millisecond)
	payload := []byte(`{
		"transaction_id":"tx-1",
		"user_id":"u1",
		"from_currency":"USD",
		"to_currency":"EUR",
		"amount":50000,
		"timestamp":"` + ts.Format(time.RFC3339Nano) + `",
		"status":"success"
	}`)

	if err := c.processMessage(context.Background(), payload); err != nil {
		t.Fatalf("processMessage: %v", err)
	}
	if len(storage.saved) != 1 {
		t.Fatalf("saved = %d, want 1", len(storage.saved))
	}
	if storage.saved[0].LatencyMs < 100 {
		t.Errorf("LatencyMs = %d, want >= 100", storage.saved[0].LatencyMs)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	storage := &mockStorage{}
	c := &Consumer{storage: storage, logger: newTestLogger()}

	if err := c.processMessage(context.Background(), []byte(`not json`)); err != nil {
		t.Errorf("processMessage = %v, want nil (битое пропускаем)", err)
	}
	if len(storage.saved) != 0 {
		t.Errorf("saved = %d, want 0", len(storage.saved))
	}
}

func TestProcessMessage_SaveError(t *testing.T) {
	storage := &mockStorage{err: errors.New("clickhouse down")}
	c := &Consumer{storage: storage, logger: newTestLogger()}

	payload := []byte(`{"transaction_id":"tx-1","timestamp":"2026-11-22T00:00:00Z"}`)
	if err := c.processMessage(context.Background(), payload); err == nil {
		t.Error("processMessage: expected error, got nil")
	}
}
