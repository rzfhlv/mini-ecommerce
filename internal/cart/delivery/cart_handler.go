package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"mini-ecommerce/internal/cart/domain"
	"mini-ecommerce/internal/cart/usecase"
	"mini-ecommerce/pkg/middleware"
)

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type CartHandler struct {
	addUC    usecase.AddItemUseCase
	updateUC usecase.UpdateItemQuantityUseCase
	removeUC usecase.RemoveItemUseCase
	getUC    usecase.GetCartUseCase
	clearUC  usecase.ClearCartUseCase
}

func NewCartHandler(add usecase.AddItemUseCase, update usecase.UpdateItemQuantityUseCase, remove usecase.RemoveItemUseCase, get usecase.GetCartUseCase, clear usecase.ClearCartUseCase) *CartHandler {
	return &CartHandler{addUC: add, updateUC: update, removeUC: remove, getUC: get, clearUC: clear}
}

func userIDFromContext(r *http.Request) string {
	if v, ok := r.Context().Value(middleware.UserIDContextKey).(string); ok {
		return v
	}
	return ""
}

func (h *CartHandler) Get(w http.ResponseWriter, r *http.Request) {
	cart, err := h.getUC.Execute(r.Context(), userIDFromContext(r))
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: cart})
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req usecase.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	req.UserID = userIDFromContext(r)
	cart, err := h.addUC.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: cart})
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var req usecase.UpdateItemQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	req.UserID = userIDFromContext(r)
	req.ProductID = r.PathValue("productId")
	cart, err := h.updateUC.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: cart})
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	cart, err := h.removeUC.Execute(r.Context(), userIDFromContext(r), r.PathValue("productId"))
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: cart})
}

func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	if err := h.clearUC.Execute(r.Context(), userIDFromContext(r)); err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "cart cleared"})
}

func mapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidQuantity):
		return http.StatusBadRequest, "quantity must be greater than zero"
	case errors.Is(err, domain.ErrCartItemNotFound):
		return http.StatusNotFound, "item not found in cart"
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
