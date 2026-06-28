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

// ServeHTTP is a tiny GraphQL endpoint for one dashboard query.
//
// This is not a full GraphQL execution engine yet. It accepts the normal
// GraphQL-over-HTTP JSON shape, checks that the query asks for "dashboard", then
// delegates to the application service. That keeps the learning path simple.
func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The gateway already validated the JWT and wrote trusted identity headers.
	// RequireGatewayAuth converted those headers into UserContext.
	user, ok := UserFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "missing user context", nil)
		return
	}

	req, err := readGraphQLRequest(r)
	if err != nil {
		writeGraphQLErrors(w, http.StatusBadRequest, "invalid GraphQL request")
		return
	}
	if !isDashboardQuery(req.Query) {
		writeGraphQLErrors(w, http.StatusBadRequest, "only dashboard query is supported")
		return
	}

	// Business authorization stays in the application service. The transport
	// layer only translates HTTP/GraphQL into a service call.
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

func readGraphQLRequest(r *http.Request) (graphQLRequest, error) {
	var req graphQLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func isDashboardQuery(query string) bool {
	return strings.Contains(strings.ToLower(query), "dashboard")
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
