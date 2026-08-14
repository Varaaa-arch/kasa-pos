package api

import "net/http"

func NewRouter(
	productHandler *ProductHandler,
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

	return Chain(
		mux,
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
