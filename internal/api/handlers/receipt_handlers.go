package handlers

import (
	"encoding/json"
	"net/http"

	"receipt-processor/internal/models"
	"receipt-processor/internal/service"

	"github.com/gorilla/mux"
)

type ReceiptHandler struct {
	service *service.ReceiptService
}

func NewReceiptHandler(service *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		service: service,
	}
}

func (h *ReceiptHandler) ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt models.Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, "The receipt is invalid", http.StatusBadRequest)
		return
	}

	id, err := h.service.ProcessReceipt(receipt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := models.ReceiptID{ID: id}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ReceiptHandler) GetPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	points, err := h.service.GetPoints(id)
	if err != nil {
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}

	response := models.Points{Points: points}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
