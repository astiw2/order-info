package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
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

func PostOrdersInfo(w http.ResponseWriter, r *http.Request) {
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
