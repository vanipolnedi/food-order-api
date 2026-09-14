package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"food-order-api/internal/domain"
	"food-order-api/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler - constructs an OrderHandler with the given order service.
func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

// Create - handles POST /order: decodes JSON, creates the order, and writes the response.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.OrderRequest
	decoder := json.NewDecoder(r.Body)
	/*
		Reject unknown JSON fields.

		Example:
		{
			"items": [...],
			"foo": "bar"
		}

		will be rejected.
	*/
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "InvalidInput", "invalid JSON request")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "InvalidInput", "request body must contain a single JSON object")
		return
	}
	order, err := h.service.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrder) {
			writeError(w, http.StatusUnprocessableEntity, "ValidationError", err.Error())
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			writeError(w, http.StatusUnprocessableEntity, "ValidationError", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalError", "failed to create order")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// writeError - writes a JSON error payload with the given HTTP status.
func writeError(w http.ResponseWriter, status int, errorType string, message string) {
	response := domain.APIResponse{
		Code:    status,
		Type:    errorType,
		Message: message,
	}
	writeJSON(w, status, response)
}
