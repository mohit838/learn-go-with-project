package router

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootAndNotFoundResponses(t *testing.T) {
	handler := NewRouter(nil, slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), nil, nil)

	rootRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	rootResponse := httptest.NewRecorder()
	handler.ServeHTTP(rootResponse, rootRequest)
	if rootResponse.Code != http.StatusOK || rootResponse.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("root response = status %d, content-type %q", rootResponse.Code, rootResponse.Header().Get("Content-Type"))
	}

	notFoundRequest := httptest.NewRequest(http.MethodGet, "/missing", nil)
	notFoundResponse := httptest.NewRecorder()
	handler.ServeHTTP(notFoundResponse, notFoundRequest)
	if notFoundResponse.Code != http.StatusNotFound || notFoundResponse.Body.String() != "{\"error\":{\"code\":\"not_found\",\"message\":\"route not found\"}}\n" {
		t.Fatalf("not-found response = status %d, body %s", notFoundResponse.Code, notFoundResponse.Body.String())
	}
}
