package transport

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/task/application"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
	"github.com/mohit838/learn-go-with-project/internal/utils"
)

type TaskHandler struct {
	tasks *application.TaskService
}

func NewTaskHandler(tasks *application.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	req, upload, err := parseCreateRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task request", err.Error())
		return
	}
	defer closeUpload(upload)

	result, err := h.tasks.Create(r.Context(), user, req, upload)
	if err != nil {
		writeTaskError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, "task created", result)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	req, upload, err := parseUpdateRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task request", err.Error())
		return
	}
	defer closeUpload(upload)

	result, err := h.tasks.Update(r.Context(), user, chi.URLParam(r, "id"), req, upload)
	if err != nil {
		writeTaskError(w, err)
		return
	}
	response.Success(w, http.StatusOK, "task updated", result)
}

func (h *TaskHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	result, err := h.tasks.FindByID(r.Context(), user, chi.URLParam(r, "id"))
	if err != nil {
		writeTaskError(w, err)
		return
	}
	response.Success(w, http.StatusOK, "task found", result)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	pagination := utils.ParsePagination(r)
	result, err := h.tasks.List(r.Context(), user, application.TaskListQuery{
		Search:   r.URL.Query().Get("search"),
		Status:   r.URL.Query().Get("status"),
		Priority: r.URL.Query().Get("priority"),
		IsActive: parseBoolPtr(r.URL.Query().Get("is_active")),
		OnlyMine: parseBoolDefault(r.URL.Query().Get("only_mine"), true),
		Page:     pagination.Page,
		PerPage:  pagination.PerPage,
		Offset:   pagination.Offset,
	})
	if err != nil {
		writeTaskError(w, err)
		return
	}
	response.Success(w, http.StatusOK, "tasks found", result)
}

func (h *TaskHandler) MarkInactive(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	if err := h.tasks.MarkInactive(r.Context(), user, chi.URLParam(r, "id")); err != nil {
		writeTaskError(w, err)
		return
	}
	response.Success(w, http.StatusOK, "task marked inactive", nil)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	if err := h.tasks.Delete(r.Context(), user, chi.URLParam(r, "id")); err != nil {
		writeTaskError(w, err)
		return
	}
	response.Success(w, http.StatusOK, "task deleted", nil)
}

func parseCreateRequest(r *http.Request) (application.CreateTaskRequest, *domain.ImageUpload, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		req, upload, err := parseMultipartTask(r)
		return application.CreateTaskRequest(req), upload, err
	}
	var req application.CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, nil, err
}

func parseUpdateRequest(r *http.Request) (application.UpdateTaskRequest, *domain.ImageUpload, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		req, upload, err := parseMultipartTask(r)
		return application.UpdateTaskRequest(req), upload, err
	}
	var req application.UpdateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, nil, err
}

func parseMultipartTask(r *http.Request) (application.CreateTaskRequest, *domain.ImageUpload, error) {
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		return application.CreateTaskRequest{}, nil, err
	}

	req := application.CreateTaskRequest{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Status:      r.FormValue("status"),
		Priority:    r.FormValue("priority"),
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return req, nil, nil
		}
		return req, nil, err
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return req, &domain.ImageUpload{
		FileName:    header.Filename,
		ContentType: contentType,
		Size:        header.Size,
		Reader:      file,
		Closer:      file,
	}, nil
}

func closeUpload(upload *domain.ImageUpload) {
	if upload != nil && upload.Closer != nil {
		_ = upload.Closer.Close()
	}
}

func parseBoolPtr(value string) *bool {
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseBoolDefault(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func writeTaskError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrInvalidInput):
		response.Error(w, http.StatusBadRequest, "invalid task input", err.Error())
	case errors.Is(err, sql.ErrNoRows):
		response.Error(w, http.StatusNotFound, "task not found", err.Error())
	case errors.Is(err, io.EOF):
		response.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "task request failed", err.Error())
	}
}
