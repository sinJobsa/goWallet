package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sinJobsa/gw/gw-exchanger/internal/config"
	"github.com/sinJobsa/gw/gw-exchanger/internal/storages"
)

func setupStorage(t *testing.T) (*Storage, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := config.Config{
		DB: config.DBConfig{
			Host: "localhost", Port: 5432, User: "postgres",
			Password: "postgres", Name: "exchanger", SSLMode: "disable",
			MaxConns: 5, MinConns: 1, ConnectTimeout: 5 * time.Second,
		},
	}
	s, err := New(ctx, cfg)
	if err != nil {
		t.Skipf("skip: %v", err)
	}
	return s, func() { s.Close() }
}

func TestGetRates_Integration(t *testing.T) {
	s, cleanup := setupStorage(t)
	defer cleanup()

	ctx := context.Background()
	rates, err := s.GetRates(ctx)
	if err != nil {
		t.Fatalf("GetRates: %v", err)
	}
	if len(rates) == 0 {
		t.Error("rates is empty, expected seed data")
	}
	if _, ok := rates["USD"]; !ok {
		t.Error("USD not found in rates")
	}
}

func TestGetRate_SameCurrency(t *testing.T) {
	s, cleanup := setupStorage(t)
	defer cleanup()

	ctx := context.Background()
	rate, err := s.GetRate(ctx, "USD", "USD")
	if err != nil {
		t.Fatalf("GetRate: %v", err)
	}
	if rate != 1.0 {
		t.Errorf("rate = %f, want 1.0", rate)
	}
}

func TestGetRate_CrossRate(t *testing.T) {
	s, cleanup := setupStorage(t)
	defer cleanup()

	ctx := context.Background()
	rate, err := s.GetRate(ctx, "USD", "EUR")
	if err != nil {
		t.Fatalf("GetRate: %v", err)
	}
	if rate <= 0 {
		t.Errorf("rate = %f, want > 0", rate)
	}
}

func TestGetRate_NotFound(t *testing.T) {
	s, cleanup := setupStorage(t)
	defer cleanup()

	ctx := context.Background()
	_, err := s.GetRate(ctx, "USD", "GBP")
	if !errors.Is(err, storages.ErrRateNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
