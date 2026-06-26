package transport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mohit838/learn-go-with-project/internal/auth/application"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

type AuthHandler struct {
	auth *application.AuthService
}

func NewAuthHandler(auth *application.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req application.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	result, err := h.auth.Register(r.Context(), req)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "user registered", result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req application.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	result, err := h.auth.Login(r.Context(), req)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "login successful", result)
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrInvalidInput):
		response.Error(w, http.StatusBadRequest, "invalid auth input", err.Error())
	case errors.Is(err, application.ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "invalid credentials", err.Error())
	case errors.Is(err, application.ErrInactiveUser):
		response.Error(w, http.StatusForbidden, "user is inactive", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "auth request failed", err.Error())
	}
}
