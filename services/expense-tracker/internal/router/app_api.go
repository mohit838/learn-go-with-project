package router

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/expense"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

func registerAppAPI(r chi.Router, serviceName string, db *sql.DB) {
	r.Get(constants.RootPath, func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{
			"service": serviceName,
			"status":  "ok",
		})
	})

	r.Get(constants.HealthPath, func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get(constants.ReadyPath, func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if db == nil || db.PingContext(ctx) != nil {
			response.Error(w, http.StatusServiceUnavailable, constants.ErrorInternalServer, "service dependencies are unavailable")
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	r.Route("/expenses", expense.NewHandler(db).Routes)
}
