package api

import (
	"encoding/json"
	"errors"
	"net/http"

	postgres "pos-system/internal/repository/postgres"
	transactionservice "pos-system/internal/service/transaction"
)

type TransactionHandler struct {
	service *transactionservice.Service
}

func NewTransactionHandler(
	service *transactionservice.Service,
) *TransactionHandler {
	return &TransactionHandler{
		service: service,
	}
}

func (h *TransactionHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	items, err := h.service.List(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Failed to list transactions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h *TransactionHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrTransactionNotFound) {
			WriteError(w, r, http.StatusNotFound, ErrCodeTransactionNotFound, "Transaction not found")
			return
		}

		WriteError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Failed to get transaction")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}
