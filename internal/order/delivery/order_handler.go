package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"mini-ecommerce/internal/order/domain"
	"mini-ecommerce/internal/order/usecase"
	"mini-ecommerce/pkg/middleware"
)

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type OrderHandler struct {
	checkoutUseCase       usecase.CheckoutUseCase
	confirmPaymentUseCase usecase.ConfirmPaymentUseCase
	webhookParser         domain.WebhookParser
}

func NewOrderHandler(checkoutUseCase usecase.CheckoutUseCase, confirmPaymentUseCase usecase.ConfirmPaymentUseCase, webhookParser domain.WebhookParser) *OrderHandler {
	return &OrderHandler{checkoutUseCase: checkoutUseCase, confirmPaymentUseCase: confirmPaymentUseCase, webhookParser: webhookParser}
}

func getUserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(middleware.UserIDContextKey).(string); ok {
		return v
	}
	return ""
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req usecase.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	req.UserID = getUserIDFromContext(r.Context())

	order, err := h.checkoutUseCase.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: order})
}

// POST /api/v1/webhooks/payment -- dipanggil gateway, bukan user.
// Produksi nyata WAJIB verifikasi signature request di sini.
func (h *OrderHandler) PaymentWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "cannot read request body"})
		return
	}
	notification, err := h.webhookParser.ParseWebhook(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid webhook payload"})
		return
	}
	if err := h.confirmPaymentUseCase.Execute(r.Context(), notification); err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "payment confirmed"})
}

func mapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrOrderEmpty):
		return http.StatusBadRequest, "cart cannot be empty"
	case errors.Is(err, domain.ErrInsufficientStock):
		return http.StatusConflict, "one or more products are out of stock"
	case errors.Is(err, domain.ErrPromoCodeNotFound):
		return http.StatusBadRequest, "promo code is invalid or expired"
	case errors.Is(err, domain.ErrDiscountExceedsSubtotal):
		return http.StatusBadRequest, "discount cannot exceed order subtotal"
	case errors.Is(err, domain.ErrInvalidPaymentMethod):
		return http.StatusBadRequest, "invalid payment method"
	case errors.Is(err, usecase.ErrOrderNotFound):
		return http.StatusNotFound, "order not found"
	case errors.Is(err, usecase.ErrInvalidUserID), errors.Is(err, usecase.ErrInvalidProductID):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func writeJSON(w http.ResponseWriter, status int, body apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
