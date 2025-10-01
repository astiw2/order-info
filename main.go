package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serverAddress = ":8080"
	readTimeout   = 10 * time.Second
	writeTimeout  = 10 * time.Second
	idleTimeout   = 120 * time.Second
)

type Order struct {
	CustomerID string `json:"customerId"`
	OrderID    string `json:"orderId"`
	Timestamp  string `json:"timestamp"`
	Items      []Item `json:"items"`
}

type Item struct {
	ItemID  string `json:"itemId"`
	CostEur int    `json:"costEur"`
}

type CustomerItem struct {
	CustomerID string `json:"customerId"`
	ItemID     string `json:"itemId"`
	CostEur    int    `json:"costEur"`
}

type CustomerSummary struct {
	CustomerID          string `json:"customerId"`
	NbrOfPurchasedItems int    `json:"nbrOfPurchasedItems"`
	TotalAmountEur      int    `json:"totalAmountEur"`
}

type ValidationError struct {
	Index   int    `json:"index"`
	OrderID string `json:"orderId"`
	Error   string `json:"error"`
}

type OrdersResponse struct {
	Items     []CustomerItem    `json:"items"`
	Summaries []CustomerSummary `json:"summaries"`
	Errors    []ValidationError `json:"errors,omitempty"`
}

func handlePostOrdersInfo(w http.ResponseWriter, r *http.Request) {
	var orders []Order
	if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := OrdersResponse{
		Items:     []CustomerItem{},
		Summaries: []CustomerSummary{},
		Errors:    []ValidationError{},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders/info", handlePostOrdersInfo)

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
