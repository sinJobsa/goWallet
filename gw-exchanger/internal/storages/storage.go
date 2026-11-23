package storages

import (
	"context"
	"errors"
)

var ErrRateNotFound = errors.New("rate not found")

type Storage interface {
	GetRates(ctx context.Context) (map[string]float32, error)
	GetRate(ctx context.Context, from, to string) (float32, error)
}
