package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astiw2/order-info/cmd/internal/handler"
)

func TestValidateOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       handler.Order
		expectedErr bool
		errContains string
	}{
		{
			name: "valid single item",
			order: handler.Order{
				CustomerID: "C1",
				OrderID:    "O1",
				Timestamp:  "123",
				Items:      []handler.Item{{ItemID: "I1", CostEur: 5}},
			},
			expectedErr: false,
		},
		{
			name: "missing customerId",
			order: handler.Order{
				OrderID:   "O1",
				Timestamp: "123",
				Items:     []handler.Item{{ItemID: "I1", CostEur: 5}},
			},
			expectedErr: true,
			errContains: "customerId",
		},
		{
			name: "missing orderId",
			order: handler.Order{
				CustomerID: "C1",
				Timestamp:  "123",
				Items:      []handler.Item{{ItemID: "I1", CostEur: 5}},
			},
			expectedErr: true,
			errContains: "orderId",
		},
		{
			name: "empty items",
			order: handler.Order{
				CustomerID: "C1",
				OrderID:    "O1",
				Timestamp:  "123",
				Items:      []handler.Item{},
			},
			expectedErr: true,
			errContains: "items",
		},
		{
			name: "item missing itemId",
			order: handler.Order{
				CustomerID: "C1",
				OrderID:    "O1",
				Timestamp:  "123",
				Items:      []handler.Item{{ItemID: "", CostEur: 3}},
			},
			expectedErr: true,
			errContains: "itemId",
		},
		{
			name: "item negative cost",
			order: handler.Order{
				CustomerID: "C1",
				OrderID:    "O1",
				Timestamp:  "123",
				Items:      []handler.Item{{ItemID: "I1", CostEur: -1}},
			},
			expectedErr: true,
			errContains: "negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateOrder(tt.order)
			if tt.expectedErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectedErr && tt.errContains != "" && !strings.Contains(err.Error, tt.errContains) {
				t.Fatalf("error %q does not contain %q", err.Error, tt.errContains)
			}
		})
	}
}

// httptest library to test HTTP handlers without starting a real server.
// It creates a mock HTTP request, and httptest.NewRecorder captures the response.
func TestPostOrdersInfoHandler(t *testing.T) {
	tests := []struct {
		name           string
		orders         []handler.Order
		wantStatusCode int
		wantItemCount  int
		wantSummaries  int
		expectedErrors int
	}{
		{
			name: "valid orders",
			orders: []handler.Order{
				{CustomerID: "C1", OrderID: "O1", Timestamp: "1", Items: []handler.Item{{ItemID: "I1", CostEur: 10}}},
				{CustomerID: "C2", OrderID: "O2", Timestamp: "2", Items: []handler.Item{{ItemID: "I2", CostEur: 15}}},
			},
			wantStatusCode: http.StatusOK,
			wantItemCount:  2,
			wantSummaries:  2,
			expectedErrors: 0,
		},
		{
			name:           "empty orders",
			orders:         []handler.Order{},
			wantStatusCode: http.StatusOK,
			wantItemCount:  0,
			wantSummaries:  0,
			expectedErrors: 1,
		},
		{
			name: "mixed valid and invalid with missing customer id",
			orders: []handler.Order{
				{CustomerID: "C1", OrderID: "O1", Timestamp: "1", Items: []handler.Item{{ItemID: "I1", CostEur: 5}}},
				{OrderID: "O2", Timestamp: "2", Items: []handler.Item{{ItemID: "I2", CostEur: 3}}},
			},
			wantStatusCode: http.StatusOK,
			wantItemCount:  1,
			wantSummaries:  1,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.orders)
			req := httptest.NewRequest(http.MethodPost, "/orders/info", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.PostOrdersInfo(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			var resp handler.OrdersResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if len(resp.Items) != tt.wantItemCount {
				t.Errorf("expected %d items, got %d", tt.wantItemCount, len(resp.Items))
			}

			if len(resp.Summaries) != tt.wantSummaries {
				t.Errorf("expected %d summaries, got %d", tt.wantSummaries, len(resp.Summaries))
			}

			if len(resp.Errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d", tt.expectedErrors, len(resp.Errors))
			}
		})
	}
}
