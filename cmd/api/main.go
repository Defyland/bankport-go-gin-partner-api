package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/httpapi"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/store"
)

func main() {
	cfg := config.Load()
	startupLogger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := cfg.Validate(); err != nil {
		startupLogger.Error("bankport_api_invalid_config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	repository := store.NewSeededRepository(cfg)
	metrics := observability.NewMetrics(cfg.ServiceName)
	router := httpapi.NewRouter(httpapi.Dependencies{
		Config:     cfg,
		Logger:     logger,
		Repository: repository,
		Metrics:    metrics,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout + 2*time.Second,
		WriteTimeout:      cfg.RequestTimeout + 2*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		logger.Info("bankport_api_starting", "addr", server.Addr, "env", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errs:
		logger.Error("bankport_api_failed", "error", err)
		os.Exit(1)
	case sig := <-stop:
		logger.Info("bankport_api_stopping", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("bankport_api_shutdown_failed", "error", err)
		os.Exit(1)
	}
	logger.Info("bankport_api_stopped")
}
