package api

import (
	"encoding/json"
	"log"
	"net/http"

	"pos-system/internal/domain/cart"
	printerreceipt "pos-system/internal/printer/receipt"
	"pos-system/internal/repository/postgres"
	"pos-system/internal/service/checkout"
)

type CheckoutRequest struct {
	Items []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`

	PaidAmount    int64  `json:"paid_amount"`
	PaymentMethod string `json:"payment_method"`
	Discount      int64  `json:"discount"`
	Tax           int64  `json:"tax"`
	ServiceCharge int64  `json:"service_charge"`
	InvoiceNumber string `json:"invoice_number"`
}

type CheckoutHandler struct {
	CheckoutService *checkout.OrchestratorService
	ProductRepo     *postgres.ProductRepository
}

func NewCheckoutHandler(
	service *checkout.OrchestratorService,
	productRepo *postgres.ProductRepository,
) *CheckoutHandler {
	return &CheckoutHandler{
		CheckoutService: service,
		ProductRepo:     productRepo,
	}
}

func (h *CheckoutHandler) Checkout(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req CheckoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if len(req.Items) == 0 {
		http.Error(
			w,
			"cart is empty",
			http.StatusBadRequest,
		)
		return
	}

	c := cart.New()

	for _, item := range req.Items {
		p, err := h.ProductRepo.GetByID(
			r.Context(),
			item.ProductID,
		)
		if err != nil {
			http.Error(
				w,
				"product not found",
				http.StatusBadRequest,
			)
			return
		}

		if err := c.AddItem(
			p,
			item.Quantity,
		); err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)
			return
		}
	}

	result, err := h.CheckoutService.Execute(
		r.Context(),
		checkout.AtomicRequest{
			Cart:          c,
			PaidAmount:    req.PaidAmount,
			Discount:      req.Discount,
			Tax:           req.Tax,
			ServiceCharge: req.ServiceCharge,
			PaymentMethod: req.PaymentMethod,
			InvoiceNumber: req.InvoiceNumber,
		},
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	tx := result.Transaction
	printJob := result.PrintJob

	if printJob.Status == printerreceipt.PrintJobFailed {
		log.Printf(
			"checkout completed with print failure: transaction_id=%s print_job_id=%s error=%s",
			tx.ID,
			printJob.ID,
			printJob.Error,
		)
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	response := map[string]any{
		"id":             tx.ID,
		"transaction_id": tx.ID,
		"invoice_number": tx.InvoiceNumber,
		"subtotal":       tx.Subtotal,
		"discount":       tx.Discount,
		"tax":            tx.Tax,
		"service_charge": tx.ServiceCharge,
		"total":          tx.Total,
		"paid_amount":    tx.PaidAmount,
		"change":         tx.Change,
		"payment_method": tx.PaymentMethod,
		"status":         tx.Status,
		"created_at":     tx.CreatedAt,
		"print_job": map[string]any{
			"id":     printJob.ID,
			"status": string(printJob.Status),
		},
	}

	if printJob.Error != "" {
		response["print_job"].(map[string]any)["error"] = printJob.Error
	}

	_ = json.NewEncoder(w).Encode(response)
}
