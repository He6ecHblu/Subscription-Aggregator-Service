package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAppPort  = "8080"
	defaultAppEnv   = "local"
	defaultLogLevel = "info"
)

type Config struct {
	AppPort     string
	AppEnv      string
	LogLevel    string
	DatabaseURL string
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppPort:     getEnv("APP_PORT", defaultAppPort),
		AppEnv:      getEnv("APP_ENV", defaultAppEnv),
		LogLevel:    getEnv("LOG_LEVEL", defaultLogLevel),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	port, err := strconv.Atoi(cfg.AppPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("APP_PORT must be a valid TCP port")
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}

	if strings.TrimSpace(cfg.AppEnv) == "" {
		return fmt.Errorf("APP_ENV must not be empty")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("load .env: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("load .env: line %d must be KEY=VALUE", lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if key == "" {
			return fmt.Errorf("load .env: line %d has empty key", lineNumber)
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("load .env: set %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	return nil
}
