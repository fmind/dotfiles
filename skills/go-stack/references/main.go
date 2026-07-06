package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"<import_path>"
	"<import_path>/config"
)

func main() {
	// Load configuration and initialize structured logging (fail fast on error).
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	// Wrap the environment-appropriate handler for OTEL trace correlation.
	handler := &<slug>.OtelHandler{Handler: cfg.NewHandler(os.Stderr)}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Graceful shutdown lifecycle setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Install OpenTelemetry tracing (no-op unless an OTLP endpoint is configured).
	shutdownOTel, err := <slug>.SetupOTel(ctx, "<slug>")
	if err != nil {
		logger.Error("setup opentelemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOTel(flushCtx); err != nil {
			logger.Error("shutdown opentelemetry", "error", err)
		}
	}()

	// Initialize the application handler
	appHandler := <slug>.NewAppHandler(logger, cfg.Environment)

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           appHandler,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			_ = server.Close()
		}
	}()

	logger.Info("server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
	logger.Info("graceful shutdown complete")
}
