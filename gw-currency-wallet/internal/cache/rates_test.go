package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestSetAndGetRates(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	ctx := context.Background()
	c, err := NewRatesCache(ctx, mr.Addr(), "", 0, time.Minute)
	if err != nil {
		t.Fatalf("NewRatesCache: %v", err)
	}
	defer c.Close()

	rates := map[string]float64{"USD": 90.5, "EUR": 98.2, "RUB": 1.0}
	if err := c.SetRates(ctx, rates); err != nil {
		t.Fatalf("SetRates: %v", err)
	}

	got, err := c.GetRates(ctx)
	if err != nil {
		t.Fatalf("GetRates: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("got %d rates, want 3", len(got))
	}
	if got["USD"] != 90.5 {
		t.Errorf("USD = %v, want 90.5", got["USD"])
	}
}

func TestGetRates_CacheMiss(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	ctx := context.Background()
	c, err := NewRatesCache(ctx, mr.Addr(), "", 0, time.Minute)
	if err != nil {
		t.Fatalf("NewRatesCache: %v", err)
	}
	defer c.Close()

	_, err = c.GetRates(ctx)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("err = %v, want ErrCacheMiss", err)
	}
}

func TestGetRates_Expired(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	ctx := context.Background()
	c, err := NewRatesCache(ctx, mr.Addr(), "", 0, time.Second)
	if err != nil {
		t.Fatalf("NewRatesCache: %v", err)
	}
	defer c.Close()

	_ = c.SetRates(ctx, map[string]float64{"USD": 90.5})

	mr.FastForward(2 * time.Second) // TTL истёк

	_, err = c.GetRates(ctx)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("err = %v, want ErrCacheMiss", err)
	}
}
