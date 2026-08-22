package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"pos-system/internal/domain/product"
	postgres "pos-system/internal/repository/postgres"
	productservice "pos-system/internal/service/product"
)

type ProductHandler struct {
	service *productservice.Service
}

func NewProductHandler(
	service *productservice.Service,
) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input struct {
		SKU   string `json:"sku"`
		Name  string `json:"name"`
		Price int64  `json:"price"`
		Stock int    `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, r, http.StatusBadRequest, ErrCodeInvalidBody, "Invalid request body")
		return
	}

	p := product.Product{
		ID:    uuid.NewString(),
		SKU:   strings.TrimSpace(input.SKU),
		Name:  strings.TrimSpace(input.Name),
		Price: input.Price,
		Stock: input.Stock,
	}

	if err := h.service.Create(r.Context(), p); err != nil {
		WriteError(w, r, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	products, err := h.service.List(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Failed to list products")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	p, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrProductNotFound) {
			WriteError(w, r, http.StatusNotFound, ErrCodeProductNotFound, "Product not found")
			return
		}

		WriteError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Failed to get product")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	var input struct {
		SKU   string `json:"sku"`
		Name  string `json:"name"`
		Price int64  `json:"price"`
		Stock int    `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, r, http.StatusBadRequest, ErrCodeInvalidBody, "Invalid request body")
		return
	}

	p := product.Product{
		ID:    id,
		SKU:   strings.TrimSpace(input.SKU),
		Name:  strings.TrimSpace(input.Name),
		Price: input.Price,
		Stock: input.Stock,
	}

	if err := h.service.Update(r.Context(), p); err != nil {
		if errors.Is(err, postgres.ErrProductNotFound) {
			WriteError(w, r, http.StatusNotFound, ErrCodeProductNotFound, "Product not found")
			return
		}

		WriteError(w, r, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, postgres.ErrProductNotFound) {
			WriteError(w, r, http.StatusNotFound, ErrCodeProductNotFound, "Product not found")
			return
		}

		WriteError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
