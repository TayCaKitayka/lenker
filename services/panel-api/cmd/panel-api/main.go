package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lenker/lenker/services/panel-api/internal/app"
	"github.com/lenker/lenker/services/panel-api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := app.Run(context.Background(), cfg); err != nil {
		slog.Error("panel api stopped with error", "error", err)
		os.Exit(1)
	}
}
