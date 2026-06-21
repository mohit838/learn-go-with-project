package router

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(db *sql.DB, mongoDB *mongo.Database) http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Auth Service API is running!"))
	})

	// MongoDB test endpoint
	r.Post("/logs", func(w http.ResponseWriter, r *http.Request) {
		collection := mongoDB.Collection("auth_logs")
		_, err := collection.InsertOne(r.Context(), map[string]any{
			"action":    "login",
			"timestamp": time.Now(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Log created!"))
	})

	return r
}
