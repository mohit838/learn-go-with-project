package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/task/application"
)

type GraphQLHandler struct {
	dashboard *application.DashboardService
}

func NewGraphQLHandler(dashboard *application.DashboardService) *GraphQLHandler {
	return &GraphQLHandler{dashboard: dashboard}
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	var req graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGraphQLErrors(w, http.StatusBadRequest, "invalid GraphQL request")
		return
	}
	if !strings.Contains(req.Query, "dashboard") {
		writeGraphQLErrors(w, http.StatusBadRequest, "only dashboard query is supported")
		return
	}

	result, err := h.dashboard.Summary(r.Context(), user)
	if err != nil {
		if errors.Is(err, application.ErrForbidden) {
			writeGraphQLErrors(w, http.StatusForbidden, "dashboard is only available for superadmin")
			return
		}
		writeGraphQLErrors(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeGraphQLData(w, http.StatusOK, map[string]any{
		"dashboard": result,
	})
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   any            `json:"data,omitempty"`
	Errors []graphQLError `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

func writeGraphQLData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(graphQLResponse{Data: data})
}

func writeGraphQLErrors(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(graphQLResponse{
		Errors: []graphQLError{{Message: message}},
	})
}
