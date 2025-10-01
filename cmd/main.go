package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/astiw2/order-info/cmd/internal/handler"
)

const (
	serverAddress = ":8080"
	readTimeout   = 10 * time.Second
	writeTimeout  = 10 * time.Second
	idleTimeout   = 120 * time.Second
)

func run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders/info", handler.PostOrdersInfo)

	srv := &http.Server{
		Addr:         serverAddress,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	slog.Info("Starting server", "address", serverAddress)
	srvErrChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErrChan <- err
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stopChan:
		slog.Info("Received shutdown signal...")
	case err := <-srvErrChan:
		slog.Error("Server error", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	slog.Info("Server stopped successfully")

	return nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err := run(context.Background()); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
