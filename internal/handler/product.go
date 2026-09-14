package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"food-order-api/internal/service"
)

type ProductHandler struct {
	service *service.ProductService
}

// NewProductHandler - constructs a ProductHandler with the given product service.
func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// List - handles GET /product and returns every product in the catalog.
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "InternalError", "failed to retrieve products")
		return
	}
	writeJSON(w, http.StatusOK, products)
}

// Get - handles GET /product/{productId} and returns one product, or 400/404.
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productId")
	/*
		OpenAPI defines productId as int64.

		Validate that contract at the HTTP boundary.
	*/
	id, err := strconv.ParseInt(productID, 10, 64)
	if err != nil || id < 0 {
		writeError(w, http.StatusBadRequest, "InvalidID", "invalid product ID")
		return
	}
	product, found, err := h.service.Get(r.Context(), strconv.FormatInt(id, 10))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "InternalError", "failed to retrieve product")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "NotFound", "product not found")
		return
	}
	writeJSON(w, http.StatusOK, product)
}

// writeJSON - writes value as application/json with the given HTTP status.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
