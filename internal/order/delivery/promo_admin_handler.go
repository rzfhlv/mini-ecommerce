package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"mini-ecommerce/internal/order/domain"
	"mini-ecommerce/internal/order/usecase"
)

type PromoAdminHandler struct {
	createUseCase    usecase.CreatePromoUseCase
	updateUseCase    usecase.UpdatePromoUseCase
	setActiveUseCase usecase.SetPromoActiveUseCase
	deleteUseCase    usecase.DeletePromoUseCase
	getUseCase       usecase.GetPromoUseCase
	listUseCase      usecase.ListPromoUseCase
}

func NewPromoAdminHandler(create usecase.CreatePromoUseCase, update usecase.UpdatePromoUseCase, setActive usecase.SetPromoActiveUseCase, del usecase.DeletePromoUseCase, get usecase.GetPromoUseCase, list usecase.ListPromoUseCase) *PromoAdminHandler {
	return &PromoAdminHandler{createUseCase: create, updateUseCase: update, setActiveUseCase: setActive, deleteUseCase: del, getUseCase: get, listUseCase: list}
}

func (h *PromoAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreatePromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	promo, err := h.createUseCase.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapPromoErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: promo})
}

func (h *PromoAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req usecase.UpdatePromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	req.ID = r.PathValue("id")
	promo, err := h.updateUseCase.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapPromoErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: promo})
}

func (h *PromoAdminHandler) Activate(w http.ResponseWriter, r *http.Request)   { h.setActive(w, r, true) }
func (h *PromoAdminHandler) Deactivate(w http.ResponseWriter, r *http.Request) { h.setActive(w, r, false) }
func (h *PromoAdminHandler) setActive(w http.ResponseWriter, r *http.Request, active bool) {
	promo, err := h.setActiveUseCase.Execute(r.Context(), r.PathValue("id"), active)
	if err != nil {
		status, msg := mapPromoErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: promo})
}

func (h *PromoAdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.deleteUseCase.Execute(r.Context(), r.PathValue("id")); err != nil {
		status, msg := mapPromoErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "promo deleted"})
}

func (h *PromoAdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	promo, err := h.getUseCase.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		status, msg := mapPromoErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: promo})
}

func (h *PromoAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	promos, err := h.listUseCase.Execute(r.Context(), page, limit)
	if err != nil {
		status, msg := mapPromoErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: promos})
}

func mapPromoErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, usecase.ErrPromoCodeAlreadyExists):
		return http.StatusConflict, "promo code already exists"
	case errors.Is(err, usecase.ErrInvalidPromoID):
		return http.StatusBadRequest, "invalid promo id"
	case errors.Is(err, domain.ErrEmptyPromoCode), errors.Is(err, domain.ErrInvalidDiscountType),
		errors.Is(err, domain.ErrInvalidPercentageVal), errors.Is(err, domain.ErrInvalidFixedAmount),
		errors.Is(err, domain.ErrPromoAlreadyExpired):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
