package router

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mohit838/learn-go-with-project/internal/avatar"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/user"
	"github.com/redis/go-redis/v9"
)

func registerAppAPI(r chi.Router, serviceName string, db *sql.DB, cache *redis.Client, avatarClient *avatar.Client) {
	r.Get(constants.RootPath, func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{
			"service": serviceName,
			"status":  "ok",
		})
	})

	r.Get(constants.HealthPath, func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/users", user.NewHandler(db, cache, avatarClient).Routes)
}
