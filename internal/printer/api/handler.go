package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	"pos-system/internal/printer/logging"
	"pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

type Handler struct {
	Printer    transport.Printer
	Renderer   *receipt.Renderer
	DevicePath string
	Logger     *logging.Logger
}

func NewHandler(
	printer transport.Printer,
	renderer *receipt.Renderer,
	devicePath string,
	logger *logging.Logger,
) *Handler {
	return &Handler{
		Printer:    printer,
		Renderer:   renderer,
		DevicePath: devicePath,
		Logger:     logger,
	}
}

type PrintRequest struct {
	Store       receipt.Store    `json:"store"`
	Transaction PrintTransaction `json:"transaction"`
	Items       []receipt.Item   `json:"items"`
	Summary     receipt.Summary  `json:"summary"`
	Payment     receipt.Payment  `json:"payment"`
	Footer      receipt.Footer   `json:"footer"`
}

type PrintTransaction struct {
	InvoiceNumber string `json:"invoice_number"`
	Cashier       string `json:"cashier"`
}

type PrintResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

type StatusResponse struct {
	Printer   string `json:"printer"`
	Device    string `json:"device"`
	Connected bool   `json:"connected"`
}

func (h *Handler) Print(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	h.Logger.Printf("print request received")

	var request PrintRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		h.Logger.Printf("invalid request body: %v", err)
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	jobID, err := generateJobID()
	if err != nil {
		h.Logger.Printf("failed to generate job ID: %v", err)
		http.Error(
			w,
			"failed to generate print job id",
			http.StatusInternalServerError,
		)
		return
	}

	h.Logger.Printf("print job created: job_id=%s", jobID)

	transaction := receipt.Receipt{
		Store: request.Store,

		Transaction: receipt.Transaction{
			InvoiceNumber: request.Transaction.InvoiceNumber,
			Cashier:       request.Transaction.Cashier,
		},

		Items:   request.Items,
		Summary: request.Summary,
		Payment: request.Payment,
		Footer:  request.Footer,
	}

	h.Logger.Printf(
		"print started: job_id=%s invoice=%s",
		jobID,
		transaction.Transaction.InvoiceNumber,
	)

	if err := receipt.Print(
		h.Printer,
		h.Renderer,
		transaction,
	); err != nil {
		h.Logger.Printf(
			"print failed: job_id=%s error=%v",
			jobID,
			err,
		)
		http.Error(
			w,
			"failed to print receipt",
			http.StatusInternalServerError,
		)
		return
	}

	h.Logger.Printf(
		"print completed: job_id=%s invoice=%s",
		jobID,
		transaction.Transaction.InvoiceNumber,
	)

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(
		PrintResponse{
			JobID:   jobID,
			Message: "receipt printed successfully",
		},
	)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	_, err := os.Stat(h.DevicePath)

	response := StatusResponse{
		Printer:   "BP-LITE58",
		Device:    h.DevicePath,
		Connected: err == nil,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func generateJobID() (string, error) {
	data := make([]byte, 8)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return "PJ-" + hex.EncodeToString(data), nil
}
