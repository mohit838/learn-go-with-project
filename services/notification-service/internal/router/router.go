package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mohit838/learn-go-with-project/internal/notification/application"
	"github.com/mohit838/learn-go-with-project/internal/notification/transport"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

func NewRouter(service *application.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		response.Success(w, http.StatusOK, "notification service is running", map[string]string{
			"service": "notification-service",
			"status":  "ok",
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, req *http.Request) {
		response.Success(w, http.StatusOK, "notification service is ready", service.Stats())
	})

	handler := transport.NewHandler(service)
	r.Group(func(r chi.Router) {
		r.Use(transport.RequireGatewayAuth())
		r.Post("/notifications", handler.Submit)
		r.Get("/notifications/stats", handler.Stats)
		r.Get("/notifications/{id}", handler.Find)
	})

	return r
}
