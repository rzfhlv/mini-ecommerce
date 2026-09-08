package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"mini-ecommerce/internal/user/domain"
	"mini-ecommerce/internal/user/usecase"
	"mini-ecommerce/pkg/middleware"
)

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type UserHandler struct {
	registerUseCase usecase.RegisterUseCase
	loginUseCase    usecase.LoginUseCase
	getMeUseCase    usecase.GetMeUseCase
}

func NewUserHandler(register usecase.RegisterUseCase, login usecase.LoginUseCase, getMe usecase.GetMeUseCase) *UserHandler {
	return &UserHandler{registerUseCase: register, loginUseCase: login, getMeUseCase: getMe}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req usecase.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	user, err := h.registerUseCase.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: user})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req usecase.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "invalid request body"})
		return
	}
	authResp, err := h.loginUseCase.Execute(r.Context(), req)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: authResp})
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	user, err := h.getMeUseCase.Execute(r.Context(), userID)
	if err != nil {
		status, msg := mapErrorToHTTP(err)
		writeJSON(w, status, apiResponse{Success: false, Message: msg})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: user})
}

func mapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyRegistered):
		return http.StatusConflict, "email already registered"
	case errors.Is(err, domain.ErrPasswordTooShort):
		return http.StatusBadRequest, "password must be at least 8 characters"
	case errors.Is(err, domain.ErrInvalidEmail):
		return http.StatusBadRequest, "invalid email format"
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid email or password"
	case errors.Is(err, usecase.ErrInvalidUserID):
		return http.StatusBadRequest, "invalid user id"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func writeJSON(w http.ResponseWriter, status int, body apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
