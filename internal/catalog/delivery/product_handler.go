package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"mini-ecommerce/internal/catalog/domain"
	"mini-ecommerce/internal/catalog/usecase"
)

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ProductHandler struct {
	createUC usecase.CreateProductUseCase
	updateUC usecase.UpdateProductUseCase
	deleteUC usecase.DeleteProductUseCase
	getUC    usecase.GetProductUseCase
	listUC   usecase.ListProductsUseCase
}

func NewProductHandler(create usecase.CreateProductUseCase, update usecase.UpdateProductUseCase, del usecase.DeleteProductUseCase, get usecase.GetProductUseCase, list usecase.ListProductsUseCase) *ProductHandler {
	return &ProductHandler{createUC: create, updateUC: update, deleteUC: del, getUC: get, listUC: list}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	product, err := h.createUC.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: product})
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req usecase.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	req.ID = r.PathValue("id")
	product, err := h.updateUC.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: product})
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.deleteUC.Execute(r.Context(), r.PathValue("id")); err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "product deleted"})
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	product, err := h.getUC.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: product})
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	result, err := h.listUC.Execute(r.Context(), category, page, limit)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: result})
}

func mapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrEmptyProductName), errors.Is(err, domain.ErrCategoryRequired),
		errors.Is(err, domain.ErrNegativeStock), errors.Is(err, domain.ErrInvalidPrice):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, usecase.ErrInvalidProductID):
		return http.StatusBadRequest, "invalid product id"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func writeJSON(w http.ResponseWriter, status int, body apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
