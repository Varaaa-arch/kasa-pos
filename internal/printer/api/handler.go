package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	domainreceipt "pos-system/internal/domain/receipt"
	"pos-system/internal/printer/logging"
	printerreceipt "pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

type Handler struct {
	Printer    transport.Printer
	Renderer   *printerreceipt.Renderer
	DevicePath string
	Logger     *logging.Logger
}

func NewHandler(
	printer transport.Printer,
	renderer *printerreceipt.Renderer,
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
	Store       domainreceipt.Store   `json:"store"`
	Transaction PrintTransaction      `json:"transaction"`
	Items       []domainreceipt.Item  `json:"items"`
	Summary     domainreceipt.Summary `json:"summary"`
	Payment     domainreceipt.Payment `json:"payment"`
	Footer      domainreceipt.Footer  `json:"footer"`
}

type PrintTransaction struct {
	ID            string `json:"id"`
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

func (h *Handler) Print(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if h.Logger != nil {
		h.Logger.Printf("print request received")
	}

	var request PrintRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		if h.Logger != nil {
			h.Logger.Printf(
				"invalid print request: %v",
				err,
			)
		}

		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	jobID, err := generateJobID()
	if err != nil {
		if h.Logger != nil {
			h.Logger.Printf(
				"failed to generate job ID: %v",
				err,
			)
		}

		http.Error(
			w,
			"failed to generate print job id",
			http.StatusInternalServerError,
		)
		return
	}

	if h.Logger != nil {
		h.Logger.Printf(
			"print job created: job_id=%s",
			jobID,
		)
	}

	input := domainreceipt.Receipt{
		Store: request.Store,

		Transaction: domainreceipt.Transaction{
			ID:            request.Transaction.ID,
			InvoiceNumber: request.Transaction.InvoiceNumber,
			Cashier:       request.Transaction.Cashier,
		},

		Items:   request.Items,
		Summary: request.Summary,
		Payment: request.Payment,
		Footer:  request.Footer,
	}

	if h.Logger != nil {
		h.Logger.Printf(
			"print started: job_id=%s invoice=%s",
			jobID,
			input.Transaction.InvoiceNumber,
		)
	}

	if err := printerreceipt.Print(
		h.Printer,
		h.Renderer,
		input,
	); err != nil {
		if h.Logger != nil {
			h.Logger.Printf(
				"print failed: job_id=%s error=%v",
				jobID,
				err,
			)
		}

		http.Error(
			w,
			"failed to print receipt",
			http.StatusInternalServerError,
		)
		return
	}

	if h.Logger != nil {
		h.Logger.Printf(
			"print completed: job_id=%s invoice=%s",
			jobID,
			input.Transaction.InvoiceNumber,
		)
	}

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

func (h *Handler) Status(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	_ = json.NewEncoder(w).Encode(response)
}

func generateJobID() (string, error) {
	data := make([]byte, 8)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return "PJ-" + hex.EncodeToString(data), nil
}
