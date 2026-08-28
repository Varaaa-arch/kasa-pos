package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	postgres "pos-system/internal/repository/postgres"
	checkoutsvc "pos-system/internal/service/checkout"
	transactionservice "pos-system/internal/service/transaction"
)

type TransactionHandler struct {
	service      *transactionservice.Service
	orchestrator *checkoutsvc.OrchestratorService
}

func NewTransactionHandler(
	service *transactionservice.Service,
	orchestrator *checkoutsvc.OrchestratorService,
) *TransactionHandler {
	return &TransactionHandler{
		service:      service,
		orchestrator: orchestrator,
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

func (h *TransactionHandler) Reprint(
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

	if h.orchestrator == nil {
		WriteError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Reprint is not configured")
		return
	}

	result := h.orchestrator.Reprint(
		r.Context(),
		item,
		"reprint-"+uuid.NewString(),
	)

	printJob := result.PrintJob

	response := map[string]any{
		"transaction_id": result.Transaction.ID,
		"invoice_number": result.Transaction.InvoiceNumber,
		"print_job": map[string]any{
			"id":     printJob.ID,
			"status": string(printJob.Status),
		},
	}

	if printJob.Error != "" {
		response["print_job"].(map[string]any)["error"] = printJob.Error
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
