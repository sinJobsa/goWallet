package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sinJobsa/gw/gw-exchanger/internal/storages"
)

func (s *Storage) GetRates(ctx context.Context) (map[string]float32, error) {
	rows, err := s.pool.Query(ctx, "SELECT currency, rate_to_rub FROM rates")
	if err != nil {
		return nil, fmt.Errorf("query rates: %w", err)
	}
	defer rows.Close()
	result := make(map[string]float32)
	for rows.Next() {
		var currency string
		var rate float32
		if err := rows.Scan(&currency, &rate); err != nil {
			return nil, fmt.Errorf("scan rates: %w", err)
		}
		result[currency] = rate
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rates: %w", err)
	}
	return result, nil
}

func (s *Storage) GetRate(ctx context.Context, from, to string) (float32, error) {
	var fromRate, toRate float32
	err := s.pool.QueryRow(ctx, "SELECT rate_to_rub FROM rates WHERE currency = $1", from).Scan(&fromRate)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, storages.ErrRateNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("query rate for %s: %w", from, err)
	}

	err = s.pool.QueryRow(ctx, "SELECT rate_to_rub FROM rates WHERE currency = $1", to).Scan(&toRate)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, storages.ErrRateNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("query rate for %s: %w", to, err)
	}

	return fromRate / toRate, nil
}
