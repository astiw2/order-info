package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

func validateOrder(order Order) *ValidationError {
	if strings.TrimSpace(order.CustomerID) == "" {
		return &ValidationError{
			OrderID: order.OrderID,
			Error:   "Missing required field: customerId",
		}
	}

	if strings.TrimSpace(order.OrderID) == "" {
		return &ValidationError{
			OrderID: order.OrderID,
			Error:   "Missing required field: orderId",
		}
	}

	if strings.TrimSpace(order.Timestamp) == "" {
		return &ValidationError{
			OrderID: order.OrderID,
			Error:   "Missing required field: timestamp",
		}
	}

	if len(order.Items) == 0 {
		return &ValidationError{
			OrderID: order.OrderID,
			Error:   "Missing required field: items must have at least one item",
		}
	}

	for i, item := range order.Items {
		if strings.TrimSpace(item.ItemID) == "" {
			return &ValidationError{
				OrderID: order.OrderID,
				Error:   fmt.Sprintf("Item %d missing required field: itemId", i),
			}
		}

		if item.CostEur < 0 {
			return &ValidationError{
				OrderID: order.OrderID,
				Error:   fmt.Sprintf("Item %d has negative cost. %d must be non-negative", i, item.CostEur),
			}
		}
	}

	return nil
}

func validateRequest(orders []Order) []ValidationError {
	var vErrors []ValidationError

	if len(orders) == 0 {
		return []ValidationError{{
			Index:   0,
			OrderID: "",
			Error:   "Request must contain at least one order",
		}}
	}
	for index, order := range orders {
		if err := validateOrder(order); err != nil {
			err.Index = index
			vErrors = append(vErrors, *err)
		}
	}

	return vErrors
}

func PostOrdersInfo(w http.ResponseWriter, r *http.Request) {
	var orders []Order
	if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	validationErrors := validateRequest(orders)

	response := OrdersResponse{
		Items:     []CustomerItem{},
		Summaries: []CustomerSummary{},
		Errors:    validationErrors,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
