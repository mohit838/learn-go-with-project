package router

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/task/application"
	"github.com/mohit838/learn-go-with-project/internal/task/infrastructure"
	"github.com/mohit838/learn-go-with-project/internal/task/transport"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(db *sql.DB, mongoDB *mongo.Database, redisClient *redis.Client, minioClient *minio.Client, minioBucket string, cfg config.Cfg) http.Handler {
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

	registerHealthRoutes(r, healthDependencies{
		db:           db,
		mongoDB:      mongoDB,
		redisClient:  redisClient,
		minioClient:  minioClient,
		minioBucket:  minioBucket,
		serviceName:  constants.ServiceName,
		serviceTitle: constants.ServiceTitle,
	})

	taskRepo := infrastructure.NewTaskRepository(db)
	imageStorage := infrastructure.NewMinIOImageStorage(minioClient, minioBucket)
	taskService := application.NewTaskService(taskRepo, imageStorage)
	taskHandler := transport.NewTaskHandler(taskService)
	tokenService := application.NewTokenService(cfg.JWTSecret, cfg.JWTIssuer)

	r.Group(func(r chi.Router) {
		r.Use(transport.RequireAuth(tokenService))
		r.Post(constants.RouteTasks, taskHandler.Create)
		r.Get(constants.RouteTasks, taskHandler.List)
		r.Get(constants.RouteTaskByID, taskHandler.FindByID)
		r.Put(constants.RouteTaskByID, taskHandler.Update)
		r.Patch(constants.RouteTaskByID, taskHandler.Update)
		r.Patch(constants.RouteTaskInactive, taskHandler.MarkInactive)
		r.Delete(constants.RouteTaskByID, taskHandler.Delete)
	})

	// ========================
	// MongoDB Endpoints
	// ========================
	// POST /tasks-log - Create task event log in MongoDB
	// Example: curl -X POST http://localhost:8485/tasks-log
	r.Post(constants.RouteLogs, func(w http.ResponseWriter, r *http.Request) {
		collection := mongoDB.Collection(constants.AuditLogCollection)
		_, err := collection.InsertOne(r.Context(), map[string]any{
			"action":    constants.DefaultLogAction,
			"timestamp": time.Now(),
		})
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to create task log", err.Error())
			return
		}
		response.Success(w, http.StatusCreated, "task log created", map[string]string{
			"collection": constants.AuditLogCollection,
		})
	})

	// ========================
	// Redis Cache Endpoints
	// ========================
	// POST /cache - Set a cache value (1 hour TTL)
	// Example: curl -X POST http://localhost:8485/cache
	r.Post(constants.RouteCache, func(w http.ResponseWriter, r *http.Request) {
		err := redisClient.Set(r.Context(), constants.DefaultCacheKey, constants.DefaultCacheValue, 1*time.Hour).Err()
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to set cache", err.Error())
			return
		}
		response.Success(w, http.StatusCreated, "cache set", map[string]string{
			"key": constants.DefaultCacheKey,
		})
	})

	// GET /cache/:key - Retrieve a cache value
	// Example: curl http://localhost:8485/cache/task_key
	r.Get(constants.RouteCacheKey, func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		val, err := redisClient.Get(r.Context(), key).Result()
		if err != nil {
			response.Error(w, http.StatusNotFound, "cache key not found", map[string]string{
				"key": key,
			})
			return
		}
		response.Success(w, http.StatusOK, "cache value found", map[string]string{
			"key":   key,
			"value": val,
		})
	})

	return r
}
