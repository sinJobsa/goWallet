package clickhouse

import (
	"context"
	"fmt"

	"github.com/sinJobsa/goWallet/gw-analytics/internal/storages"
)

func (s *Storage) SaveEvent(ctx context.Context, event *storages.Event) error {
	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}
	defer batch.Close()

	if err := batch.AppendStruct(event); err != nil {
		return fmt.Errorf("append event: %w", err)
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}
	return nil
}
