package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lenker/lenker/services/node-agent/internal/agent"
	"github.com/lenker/lenker/services/node-agent/internal/config"
	httpapi "github.com/lenker/lenker/services/node-agent/internal/http"
)

func Run(ctx context.Context, cfg config.Config) error {
	logger := newLogger(cfg)
	logger.Info("starting node agent", "addr", cfg.HTTPAddr, "node_id", cfg.NodeID)

	agentService := agent.NewService(agent.Identity{
		NodeID:         cfg.NodeID,
		BootstrapToken: cfg.BootstrapToken,
		NodeToken:      cfg.NodeToken,
		PanelURL:       cfg.PanelURL,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(httpapi.RouterDeps{Agent: agentService}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("node agent http server listening", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-runCtx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("node agent stopped")
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	if cfg.LogLevel == "warn" {
		level = slog.LevelWarn
	}
	if cfg.LogLevel == "error" {
		level = slog.LevelError
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
