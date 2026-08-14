package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

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
	Store       PrintStore       `json:"store"`
	Transaction PrintTransaction `json:"transaction"`
	Items       []PrintItem      `json:"items"`
	Summary     PrintSummary     `json:"summary"`
	Payment     PrintPayment     `json:"payment"`
	Footer      PrintFooter      `json:"footer"`
}

type PrintStore struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type PrintTransaction struct {
	ID            string    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	Timestamp     time.Time `json:"timestamp"`
	Cashier       string    `json:"cashier"`
}

type PrintItem struct {
	ProductID string `json:"product_id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

type PrintSummary struct {
	Subtotal      int64 `json:"subtotal"`
	Discount      int64 `json:"discount"`
	Tax           int64 `json:"tax"`
	ServiceCharge int64 `json:"service_charge"`
	Total         int64 `json:"total"`
}

type PrintPayment struct {
	Method string `json:"method"`
	Paid   int64  `json:"paid"`
	Change int64  `json:"change"`
}

type PrintFooter struct {
	Message string `json:"message"`
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

	items := make([]domainreceipt.Item, len(request.Items))
	for i, it := range request.Items {
		items[i] = domainreceipt.Item{
			ProductID: it.ProductID,
			SKU:       it.SKU,
			Name:      it.Name,
			Quantity:  it.Quantity,
			UnitPrice: domainreceipt.NewMoney(it.UnitPrice, domainreceipt.IDR),
		}
	}

	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    request.Store.Name,
			Address: request.Store.Address,
			Phone:   request.Store.Phone,
		},

		Transaction: domainreceipt.Transaction{
			ID:            request.Transaction.ID,
			InvoiceNumber: request.Transaction.InvoiceNumber,
			Timestamp:     request.Transaction.Timestamp,
			Cashier:       request.Transaction.Cashier,
		},

		Items: items,

		Summary: domainreceipt.Summary{
			Subtotal:      domainreceipt.NewMoney(request.Summary.Subtotal, domainreceipt.IDR),
			Discount:      domainreceipt.NewMoney(request.Summary.Discount, domainreceipt.IDR),
			Tax:           domainreceipt.NewMoney(request.Summary.Tax, domainreceipt.IDR),
			ServiceCharge: domainreceipt.NewMoney(request.Summary.ServiceCharge, domainreceipt.IDR),
			Total:         domainreceipt.NewMoney(request.Summary.Total, domainreceipt.IDR),
		},

		Payment: domainreceipt.Payment{
			Method: request.Payment.Method,
			Paid:   domainreceipt.NewMoney(request.Payment.Paid, domainreceipt.IDR),
			Change: domainreceipt.NewMoney(request.Payment.Change, domainreceipt.IDR),
		},

		Footer: domainreceipt.Footer{
			Message: request.Footer.Message,
		},
	}

	if h.Logger != nil {
		h.Logger.Printf(
			"print job created: job_id=%s",
			jobID,
		)

		h.Logger.Printf(
			"print started: job_id=%s invoice=%s",
			jobID,
			input.Transaction.InvoiceNumber,
		)
	}

	job := printerreceipt.NewPrintJob(
		jobID,
		input,
	)

	err = job.Run(
		h.Printer,
		h.Renderer,
	)

	if err != nil {
		var validationErr printerreceipt.ValidationError

		if errors.As(err, &validationErr) {
			if h.Logger != nil {
				h.Logger.Printf(
					"invalid receipt: job_id=%s error=%v",
					jobID,
					err,
				)
			}

			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)
			return
		}

		if h.Logger != nil {
			h.Logger.Printf(
				"print failed: job_id=%s status=%s error=%v",
				jobID,
				job.Status,
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
