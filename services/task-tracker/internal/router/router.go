package router

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(db *sql.DB, mongoDB *mongo.Database, redisClient *redis.Client) http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Task Tracker Service API is running!"))
	})

	// MongoDB test endpoint
	r.Post("/tasks-log", func(w http.ResponseWriter, r *http.Request) {
		collection := mongoDB.Collection("task_logs")
		_, err := collection.InsertOne(r.Context(), map[string]any{
			"action":    "task_created",
			"timestamp": time.Now(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Task log created!"))
	})

	// Redis test endpoint
	r.Post("/cache", func(w http.ResponseWriter, r *http.Request) {
		err := redisClient.Set(r.Context(), "task_key", "task_value", 1*time.Hour).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Cache set!"))
	})

	// Redis get endpoint
	r.Get("/cache/:key", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		val, err := redisClient.Get(r.Context(), key).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write([]byte(val))
	})

	return r
}
