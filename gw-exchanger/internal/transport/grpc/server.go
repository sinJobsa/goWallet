package transport

import (
	"context"
	"errors"

	pb "github.com/proto-exchange/exchange"
	"github.com/sinJobsa/gw/gw-exchanger/internal/storages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExchangeServer struct {
	pb.UnimplementedExchangeServiceServer
	storage storages.Storage
}

func New(storage storages.Storage) *ExchangeServer {
	return &ExchangeServer{storage: storage}
}

func (s *ExchangeServer) GetExchangeRates(ctx context.Context, _ *pb.Empty) (*pb.ExchangeRatesResponse, error) {
	rates, err := s.storage.GetRates(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rates: %v", err)
	}
	return &pb.ExchangeRatesResponse{Rates: rates}, nil
}

func (s *ExchangeServer) GetExchangeRateForCurrency(ctx context.Context, req *pb.CurrencyRequest) (*pb.ExchangeRateResponse, error) {
	from := req.GetFromCurrency()
	to := req.GetToCurrency()
	if from == "" || to == "" {
		return nil, status.Error(codes.InvalidArgument, "from and to must be specified")
	}

	rate, err := s.storage.GetRate(ctx, from, to)
	if errors.Is(err, storages.ErrRateNotFound) {
		return nil, status.Errorf(codes.NotFound, "rate %s-%s not found", from, to)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rate: %v", err)
	}
	return &pb.ExchangeRateResponse{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
	}, nil
}
