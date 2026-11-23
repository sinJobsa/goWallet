package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/sinJobsa/goWallet/gw-currency-wallet/docs"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/auth"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/cache"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/config"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/kafka"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/storages/postgres"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/transport/exchanger"
	transporthttp "github.com/sinJobsa/goWallet/gw-currency-wallet/internal/transport/http"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/pkg/logger"
)

// @title           gw-currency-wallet API
// @version         1.0
// @description     Кошелёк-обменник с JWT-аутентификацией
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	configPath := flag.String("c", "config.env", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	//logger
	logger := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("config loaded",
		"http_port", cfg.HTTP.Port,
		"log_level", cfg.Log.Level,
	)

	//storages
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DB.ConnectTimeout)
	defer cancel()

	storage, err := postgres.New(ctx, *cfg)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer storage.Close()
	logger.Info("connected to postgres", "db", cfg.DB.Name)

	//auth
	authSvc := auth.New(cfg.JWT.Secret, cfg.JWT.TTL)

	//exchanger client
	exchangerClient, err := exchanger.New(cfg.Exchanger.Addr, cfg.Exchanger.Timeout)
	if err != nil {
		logger.Error("failed to create exchanger client", "error", err)
		os.Exit(1)
	}
	defer exchangerClient.Close()
	logger.Info("exchanger client created", "addr", cfg.Exchanger.Addr)

	//redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	ratesCache, err := cache.NewRatesCache(ctx, redisAddr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.RatesTTL)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer ratesCache.Close()
	logger.Info("connected to redis", "addr", redisAddr)

	//kafka
	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, logger)
	defer producer.Close()
	logger.Info("kafka producer created", "topic", cfg.Kafka.Topic)

	//HTTP
	handler := transporthttp.NewHandler(storage, authSvc, exchangerClient, ratesCache, producer, logger)
	router := transporthttp.NewRouter(handler, authSvc)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http serve failed", "error", err)
		}
	}()

	//Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown failed", "error", err)
	} else {
		logger.Info("http server stopped gracefully")
	}

	logger.Info("wallet stopped")

}
