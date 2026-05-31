package bankportapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/httpapi"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/store"
)

func Run(ctx context.Context, stdout, stderr io.Writer) error {
	cfg := config.Load()
	startupLogger := slog.New(slog.NewJSONHandler(stderr, nil))
	if err := cfg.Validate(); err != nil {
		startupLogger.Error("bankport_api_invalid_config", "error", err)
		return err
	}

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{
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
		logger.Info("bankport_api_starting",
			"addr", server.Addr,
			"env", cfg.Environment,
			"go_version", runtime.Version(),
			"gomaxprocs", runtime.GOMAXPROCS(0),
			"num_cpu", runtime.NumCPU(),
			"pprof_enabled", cfg.PprofEnabled,
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errs:
		logger.Error("bankport_api_failed", "error", err)
		return err
	case <-signalCtx.Done():
		logger.Info("bankport_api_stopping", "reason", signalCtx.Err())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("bankport_api_shutdown_failed", "error", err)
		return err
	}
	logger.Info("bankport_api_stopped")
	return nil
}
