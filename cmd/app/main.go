package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "subscription-aggregator-service/docs"
	"subscription-aggregator-service/internal/config"
	"subscription-aggregator-service/internal/handler"
	"subscription-aggregator-service/internal/repository"
	"subscription-aggregator-service/internal/service"
	"subscription-aggregator-service/internal/transport"

	"github.com/jackc/pgx/v5/pgxpool"
)

// @title Subscription Aggregator Service API
// @version 1.0
// @description REST service for aggregating user online subscription data.
// @BasePath /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))

	if cfg.DatabaseURL == "" {
		logger.Error("config_validation_failed", "error", "DATABASE_URL must not be empty")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_pool_create_failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("database_ping_failed", "error", err)
		os.Exit(1)
	}

	logger.Info("database_connected")

	subscriptionRepository := repository.NewSubscriptionRepository(dbPool)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, logger)
	router := transport.NewRouter(subscriptionHandler, logger)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info(
		"server_started",
		"addr", server.Addr,
		"env", cfg.AppEnv,
		"database_configured", cfg.DatabaseURL != "",
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server_failed", "error", err)
		os.Exit(1)
	}
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
