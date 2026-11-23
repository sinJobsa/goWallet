package consumer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sinJobsa/goWallet/gw-notification/internal/storages"
)

// mockStorage — мок storages.Storage.
type mockStorage struct {
	saved []*storages.Transfer
	err   error
}

func (m *mockStorage) SaveTransfer(ctx context.Context, t *storages.Transfer) error {
	if m.err != nil {
		return m.err
	}
	m.saved = append(m.saved, t)
	return nil
}

func (m *mockStorage) Close(ctx context.Context) error { return nil }

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestProcessMessage_Success(t *testing.T) {
	storage := &mockStorage{}
	c := &Consumer{storage: storage, logger: newTestLogger()}

	payload := []byte(`{
		"transaction_id":"tx-123",
		"user_id":"user-1",
		"from_currency":"USD",
		"to_currency":"EUR",
		"amount":50000,
		"timestamp":"2026-11-20T12:00:00Z",
		"status":"success"
	}`)

	ctx := context.Background()
	if err := c.processMessage(ctx, payload); err != nil {
		t.Fatalf("processMessage: %v", err)
	}

	if len(storage.saved) != 1 {
		t.Fatalf("saved = %d, want 1", len(storage.saved))
	}
	got := storage.saved[0]
	if got.TransactionID != "tx-123" {
		t.Errorf("TransactionID = %q", got.TransactionID)
	}
	if got.Amount != 50000 {
		t.Errorf("Amount = %v", got.Amount)
	}
	if got.ReceivedAt.IsZero() {
		t.Error("ReceivedAt is zero")
	}
	if got.ReceivedAt.Location() != time.UTC {
		t.Errorf("ReceivedAt not UTC: %v", got.ReceivedAt.Location())
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	storage := &mockStorage{}
	c := &Consumer{storage: storage, logger: newTestLogger()}

	ctx := context.Background()
	err := c.processMessage(ctx, []byte(`not a json`))
	if err != nil {
		t.Errorf("processMessage = %v, want nil (битое пропускаем)", err)
	}
	if len(storage.saved) != 0 {
		t.Errorf("saved = %d, want 0", len(storage.saved))
	}
}

func TestProcessMessage_SaveError(t *testing.T) {
	storage := &mockStorage{err: errors.New("mongo down")}
	c := &Consumer{storage: storage, logger: newTestLogger()}

	payload := []byte(`{"transaction_id":"tx-123"}`)
	ctx := context.Background()
	err := c.processMessage(ctx, payload)
	if err == nil {
		t.Error("processMessage: expected error, got nil")
	}
}
