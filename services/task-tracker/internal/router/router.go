package router

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	appLogger "github.com/mohit838/learn-go-with-project/internal/logger"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

func NewRouter(db *sql.DB, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(appLogger.RequestLogger(log))
	r.Use(appLogger.Recovery(log))
	r.Use(middleware.Timeout(60 * time.Second))

	registerAppAPI(r, "task-tracker-service")
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusNotFound, constants.ErrorNotFound, "route not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusMethodNotAllowed, constants.ErrorMethodNotAllowed, "method not allowed")
	})

	return r
}
