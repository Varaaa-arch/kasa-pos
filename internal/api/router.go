package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type ReadinessResponse struct {
	API  bool `json:"api"`
	DB   bool `json:"db"`
	Ready bool `json:"ready"`
}

func NewRouter(
	productHandler *ProductHandler,
	transactionHandler *TransactionHandler,
	checkoutHandler *CheckoutHandler,
	db *sql.DB,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ready", readyHandler(db))

	if productHandler != nil {
		mux.HandleFunc("GET /products", productHandler.List)
		mux.HandleFunc("POST /products", productHandler.Create)
		mux.HandleFunc("GET /products/{id}", productHandler.GetByID)
		mux.HandleFunc("PUT /products/{id}", productHandler.Update)
		mux.HandleFunc("DELETE /products/{id}", productHandler.Delete)
	}

	if checkoutHandler != nil {
		mux.HandleFunc("POST /checkout", checkoutHandler.Checkout)
	}

	if transactionHandler != nil {
		mux.HandleFunc("GET /transactions", transactionHandler.List)
		mux.HandleFunc("GET /transactions/{id}", transactionHandler.GetByID)
	}

	return Chain(
		mux,
		CORS,
		RequestID,
		RequestLogger,
		Recover,
	)
}

func healthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status: "ok",
	}

	json.NewEncoder(w).Encode(response)
}

func readyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := ReadinessResponse{
			API:  true,
			DB:   false,
			Ready: false,
		}

		// Check database connection
		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err == nil {
				response.DB = true
			}
		}

		response.Ready = response.API && response.DB

		w.Header().Set("Content-Type", "application/json")

		if response.Ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(response)
	}
}
