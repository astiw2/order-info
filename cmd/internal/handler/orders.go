package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const (
	contentType = "application/json"

	internalServerErrorMessage = "500 Internal Server Error"
	badRequestMessage          = "400 Bad Request"
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

type OrderProcessor struct {
	items   []CustomerItem
	errs    []ValidationError
	summary map[string]*CustomerSummary
}

func NewOrderProcessor() *OrderProcessor {
	return &OrderProcessor{summary: make(map[string]*CustomerSummary)}
}

func PostOrdersInfo(w http.ResponseWriter, r *http.Request) {
	var orders []Order
	if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
		http.Error(w, badRequestMessage, http.StatusBadRequest)
		return
	}

	processor := NewOrderProcessor()
	processor.Process(orders)
	response := processor.OrderResponse()

	w.Header().Set("Content-Type", contentType)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
	}
}

func (p *OrderProcessor) Process(orders []Order) {
	if len(orders) == 0 {
		p.errs = append(p.errs, ValidationError{Index: 0, OrderID: "", Error: "Request must contain at least one order to process"})
		return
	}

	for i, o := range orders {
		if vErr := ValidateOrder(o); vErr != nil {
			vErr.Index = i
			p.errs = append(p.errs, *vErr)
			continue
		}
		for _, item := range o.Items {
			p.items = append(p.items, CustomerItem{CustomerID: o.CustomerID, ItemID: item.ItemID, CostEur: item.CostEur})
			s, ok := p.summary[o.CustomerID]
			if !ok {
				s = &CustomerSummary{CustomerID: o.CustomerID}
				p.summary[o.CustomerID] = s
			}
			s.NbrOfPurchasedItems++
			s.TotalAmountEur += item.CostEur
		}
	}
}

func (p *OrderProcessor) OrderResponse() OrdersResponse {
	summaries := make([]CustomerSummary, 0, len(p.summary))
	for _, s := range p.summary {
		summaries = append(summaries, *s)
	}
	return OrdersResponse{
		Items:     p.items,
		Summaries: summaries,
		Errors:    p.errs,
	}
}

func ValidateOrder(order Order) *ValidationError {
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

		if item.CostEur <= 0 {
			return &ValidationError{
				OrderID: order.OrderID,
				Error:   fmt.Sprintf("Item %d has negative cost. %d must be non-negative", i, item.CostEur),
			}
		}
	}

	return nil
}
