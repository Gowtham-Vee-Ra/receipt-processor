package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"receipt-processor/internal/models"
	"receipt-processor/internal/service"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() (*mux.Router, *service.ReceiptService) {
	svc := service.NewReceiptService()
	handler := NewReceiptHandler(svc)

	router := mux.NewRouter()
	router.HandleFunc("/receipts/process", handler.ProcessReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", handler.GetPoints).Methods("GET")

	return router, svc
}

func TestProcessReceipt(t *testing.T) {
	router, _ := setupTestServer()

	tests := []struct {
		name           string
		receipt        models.Receipt
		expectedStatus int
		wantError      bool
	}{
		{
			name: "valid receipt",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
				},
				Total: "6.49",
			},
			expectedStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name: "invalid receipt",
			receipt: models.Receipt{
				Retailer: "Target@123",
				// Missing required fields
			},
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.receipt)
			req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if !tt.wantError {
				var response models.ReceiptID
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
			}
		})
	}
}

func TestGetPoints(t *testing.T) {
	router, svc := setupTestServer()

	// First create a receipt to test with
	receipt := models.Receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Items: []models.Item{
			{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
		},
		Total: "6.49",
	}
	id, err := svc.ProcessReceipt(receipt)
	if err != nil {
		t.Fatalf("Failed to create test receipt: %v", err)
	}

	tests := []struct {
		name           string
		receiptID      string
		expectedStatus int
		wantError      bool
	}{
		{
			name:           "existing receipt",
			receiptID:      id,
			expectedStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "non-existent receipt",
			receiptID:      "invalid-id",
			expectedStatus: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/receipts/"+tt.receiptID+"/points", nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if !tt.wantError {
				var response models.Points
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, response.Points, int64(0))
			}
		})
	}
}
