package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lenker/lenker/services/node-agent/internal/app"
	"github.com/lenker/lenker/services/node-agent/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := app.Run(context.Background(), cfg); err != nil {
		slog.Error("node agent stopped with error", "error", err)
		os.Exit(1)
	}
}
