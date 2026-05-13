package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	HTTPAddr        string
	DatabaseURL     string
	DatabasePing    bool
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:          getenv("LENKER_APP_ENV", "development"),
		HTTPAddr:        getenv("LENKER_HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("LENKER_DATABASE_URL"),
		DatabasePing:    getenv("LENKER_DATABASE_PING", "false") == "true",
		LogLevel:        getenv("LENKER_LOG_LEVEL", "info"),
		ShutdownTimeout: 10 * time.Second,
	}

	if raw := os.Getenv("LENKER_SHUTDOWN_TIMEOUT_SECONDS"); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds <= 0 {
			return Config{}, errors.New("LENKER_SHUTDOWN_TIMEOUT_SECONDS must be a positive integer")
		}
		cfg.ShutdownTimeout = time.Duration(seconds) * time.Second
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
