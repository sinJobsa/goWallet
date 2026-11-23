package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/config"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config) (*Storage, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}
	poolCfg.MaxConns = int32(cfg.DB.MaxConns)
	poolCfg.MinConns = int32(cfg.DB.MinConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
