package clickhouse

import (
	"context"
	"testing"
	"time"

	"github.com/sinJobsa/goWallet/gw-analytics/internal/config"
	"github.com/sinJobsa/goWallet/gw-analytics/internal/storages"
)

func TestSaveEvent_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := config.ClickHouseConfig{
		Addr: "localhost:9000", Database: "analytics",
		User: "default", Password: "", Timeout: 5 * time.Second,
	}
	s, err := New(ctx, cfg)
	if err != nil {
		t.Skipf("skip: %v", err)
	}
	defer s.Close()

	ev := &storages.Event{
		TransactionID: "tx-test-1",
		UserID:        "user-1",
		FromCurrency:  "USD",
		ToCurrency:    "EUR",
		Amount:        50000,
		Status:        "success",
		CreatedAt:     time.Now().UTC(),
		ReceivedAt:    time.Now().UTC(),
		LatencyMs:     123,
	}
	if err := s.SaveEvent(ctx, ev); err != nil {
		t.Fatalf("SaveEvent: %v", err)
	}
}
