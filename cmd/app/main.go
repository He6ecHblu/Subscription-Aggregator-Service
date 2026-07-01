package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"subscription-aggregator-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           mux,
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
