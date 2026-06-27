package transport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/notification/application"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	var req application.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	notification, err := h.service.Submit(r.Context(), user, req)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Success(w, http.StatusAccepted, "notification queued", notification)
}

func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	notification, err := h.service.Find(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, "notification found", notification)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, "notification worker stats", h.service.Stats())
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrInvalidInput):
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, application.ErrMissingUser):
		response.Error(w, http.StatusUnauthorized, err.Error(), nil)
	case errors.Is(err, application.ErrQueueFull):
		response.Error(w, http.StatusTooManyRequests, err.Error(), nil)
	case errors.Is(err, application.ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error(), nil)
	default:
		response.Error(w, http.StatusInternalServerError, "notification request failed", err.Error())
	}
}
