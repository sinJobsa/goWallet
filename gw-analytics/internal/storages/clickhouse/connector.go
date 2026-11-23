package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sinJobsa/goWallet/gw-analytics/internal/config"
)

type Storage struct {
	conn clickhouse.Conn
}

func New(ctx context.Context, cfg config.ClickHouseConfig) (*Storage, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		DialTimeout: cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	return &Storage{conn: conn}, nil
}

func (s *Storage) Close() error {
	return s.conn.Close()
}
