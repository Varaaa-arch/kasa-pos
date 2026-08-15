package api

import "net/http"

func NewRouter(
	productHandler *ProductHandler,
	transactionHandler *TransactionHandler,
	checkoutHandler *CheckoutHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

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

	w.Write([]byte(`{"status":"ok"}`))
}
