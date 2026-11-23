package exchanger

import (
	"context"
	"fmt"
	"time"

	pb "github.com/proto-exchange/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	client  pb.ExchangeServiceClient
	timeout time.Duration
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("NewClient GRPC error: %w", err)
	}

	return &Client{conn: conn, client: pb.NewExchangeServiceClient(conn), timeout: timeout}, nil
}

func (c *Client) GetRates(ctx context.Context) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetExchangeRates(ctx, &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("GetExchangeRates GRPC error: %w", err)
	}
	rates := make(map[string]float64, len(resp.GetRates()))
	for key, rate := range resp.GetRates() {
		rates[key] = float64(rate)
	}
	return rates, nil
}

func (c *Client) GetRate(ctx context.Context, from, to string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetExchangeRateForCurrency(ctx, &pb.CurrencyRequest{FromCurrency: from, ToCurrency: to})
	if err != nil {
		return 0, fmt.Errorf("GetExchangeRateForCurrency error: %w", err)
	}
	return float64(resp.GetRate()), nil

}

func (c *Client) Close() error {
	return c.conn.Close()
}
