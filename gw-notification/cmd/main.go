package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sinJobsa/goWallet/gw-notification/internal/config"
	"github.com/sinJobsa/goWallet/gw-notification/internal/consumer"
	"github.com/sinJobsa/goWallet/gw-notification/internal/storages/mongo"
	"github.com/sinJobsa/goWallet/gw-notification/pkg/logger"
)

func main() {
	configPath := flag.String("c", "config.env", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("config loaded",
		"kafka_topic", cfg.Kafka.Topic,
		"kafka_group", cfg.Kafka.GroupID,
		"mongo_db", cfg.Mongo.Database,
	)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.Timeout)
	defer cancel()

	storage, err := mongo.New(ctx, cfg.Mongo)
	if err != nil {
		logger.Error("failed to connect to mongo", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to mongo", "db", cfg.Mongo.Database)


	cons := consumer.New(cfg.Kafka, storage, logger)
	defer cons.Close()


	runCtx, runCancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := cons.Run(runCtx); err != nil {
			logger.Error("consumer stopped with error", "error", err)
		}
	}()

	logger.Info("notification started",
		"topic", cfg.Kafka.Topic,
		"group", cfg.Kafka.GroupID,
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("shutdown signal received")

	runCancel()

	select {
	case <-done:
		logger.Info("consumer stopped gracefully")
	case <-time.After(cfg.Server.ShutdownTimeout):
		logger.Warn("consumer shutdown timeout, forcing exit")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := storage.Close(shutdownCtx); err != nil {
		logger.Error("failed to close mongo", "error", err)
	} else {
		logger.Info("mongo closed")
	}

	logger.Info("notification stopped")
}
