package router

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/avatar"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/taskclient"
	"github.com/mohit838/learn-go-with-project/internal/user"
	"github.com/redis/go-redis/v9"
)

func registerAppAPI(r chi.Router, serviceName string, db *sql.DB, cache *redis.Client, avatarClient *avatar.Client, taskClient *taskclient.Client) {
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

	r.Route("/users", user.NewHandler(db, cache, avatarClient).Routes)
	r.Route("/tasks", taskclient.Routes(taskClient))
}
