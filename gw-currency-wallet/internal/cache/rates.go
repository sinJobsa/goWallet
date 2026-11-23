package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

const ratesKey = "rates:all"

type RatesCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRatesCache(ctx context.Context, addr, password string, db int, ttl time.Duration) (*RatesCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RatesCache{client: client, ttl: ttl}, nil
}

func (c *RatesCache) GetRates(ctx context.Context) (map[string]float64, error) {
	raw, err := c.client.Get(ctx, ratesKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}
	var rates map[string]float64
	if err := json.Unmarshal([]byte(raw), &rates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rates: %w", err)
	}
	return rates, nil
}

func (c *RatesCache) SetRates(ctx context.Context, rates map[string]float64) error {
	raw, err := json.Marshal(rates)
	if err != nil {
		return fmt.Errorf("failed to marshal rates: %w", err)
	}
	if err := c.client.Set(ctx, ratesKey, raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set rates: %w", err)
	}
	return nil
}

func (c *RatesCache) Close() error {
	return c.client.Close()
}
