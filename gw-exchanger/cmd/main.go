package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/proto-exchange/exchange"
	"github.com/sinJobsa/gw/gw-exchanger/internal/config"
	"github.com/sinJobsa/gw/gw-exchanger/internal/storages/postgres"
	transportGRPC "github.com/sinJobsa/gw/gw-exchanger/internal/transport/grpc"
	lg "github.com/sinJobsa/gw/gw-exchanger/pkg/logger"
	"google.golang.org/grpc"
)

func main() {
	configPath := flag.String("c", "config.env", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("Error loading config", "path", *configPath)
		os.Exit(1)
	}

	logger := lg.NewLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("config loaded", "grpc_host", cfg.GRPC.Host, "grpc_port", cfg.GRPC.Port, "log_level", cfg.Log.Level)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DB.ConnectTimeout)
	defer cancel()

	storage, err := postgres.New(ctx, *cfg)
	if err != nil {
		logger.Error("Error creating storages", "error", err)
		os.Exit(1)
	}

	defer storage.Close()
	logger.Info("Connected to database", "host", cfg.DB.Host, "port", cfg.DB.Port)

	grpcServer := grpc.NewServer()
	pb.RegisterExchangeServiceServer(grpcServer, transportGRPC.New(storage))

	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("Error creating listener", "error", err)
		os.Exit(1)
	}

	logger.Info("Listening on", "host", addr, "port", cfg.GRPC.Port)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Error serving gRPC", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("Shutdown signal received, exiting...")

	done := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Shutdown complete")
	case <-time.After(cfg.Server.ShutdownTimeout):
		logger.Warn("Server shutdown timed out, forcing stop")
	}

	log.Println("Exchanger stopped")

}
