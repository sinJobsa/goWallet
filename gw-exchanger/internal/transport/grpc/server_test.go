package transport

import (
	"context"
	"errors"
	"testing"

	pb "github.com/proto-exchange/exchange"
	"github.com/sinJobsa/gw/gw-exchanger/internal/storages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockStorage — мок storages.Storage.
type mockStorage struct {
	rates map[string]float32
	err   error
}

func (m *mockStorage) GetRates(ctx context.Context) (map[string]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rates, nil
}

func (m *mockStorage) GetRate(ctx context.Context, from, to string) (float32, error) {
	if m.err != nil {
		return 0, m.err
	}
	fromRate, ok := m.rates[from]
	if !ok {
		return 0, storages.ErrRateNotFound
	}
	toRate, ok := m.rates[to]
	if !ok {
		return 0, storages.ErrRateNotFound
	}
	return fromRate / toRate, nil
}

func TestGetExchangeRates_Success(t *testing.T) {
	storage := &mockStorage{
		rates: map[string]float32{"USD": 90.5, "EUR": 98.2, "RUB": 1.0},
	}
	srv := New(storage)

	resp, err := srv.GetExchangeRates(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("GetExchangeRates: %v", err)
	}
	if len(resp.GetRates()) != 3 {
		t.Errorf("rates len = %d, want 3", len(resp.GetRates()))
	}
	if resp.GetRates()["USD"] != 90.5 {
		t.Errorf("USD = %v, want 90.5", resp.GetRates()["USD"])
	}
}

func TestGetExchangeRates_StorageError(t *testing.T) {
	storage := &mockStorage{err: errors.New("db down")}
	srv := New(storage)

	_, err := srv.GetExchangeRates(context.Background(), &pb.Empty{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Errorf("code = %v, want Internal", st.Code())
	}
}

func TestGetExchangeRateForCurrency_Success(t *testing.T) {
	storage := &mockStorage{
		rates: map[string]float32{"USD": 90.5, "EUR": 98.2},
	}
	srv := New(storage)

	resp, err := srv.GetExchangeRateForCurrency(context.Background(), &pb.CurrencyRequest{
		FromCurrency: "USD",
		ToCurrency:   "EUR",
	})
	if err != nil {
		t.Fatalf("GetExchangeRateForCurrency: %v", err)
	}
	if resp.GetFromCurrency() != "USD" {
		t.Errorf("FromCurrency = %q", resp.GetFromCurrency())
	}
	if resp.GetRate() <= 0 {
		t.Errorf("Rate = %v, want > 0", resp.GetRate())
	}
}

func TestGetExchangeRateForCurrency_NotFound(t *testing.T) {
	storage := &mockStorage{
		rates: map[string]float32{"USD": 90.5},
	}
	srv := New(storage)

	_, err := srv.GetExchangeRateForCurrency(context.Background(), &pb.CurrencyRequest{
		FromCurrency: "USD",
		ToCurrency:   "GBP",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}
