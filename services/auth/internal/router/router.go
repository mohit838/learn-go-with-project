package router

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/mohit838/go-project-stater/config"
	"github.com/mohit838/learn-go-with-project/internal/auth/application"
	"github.com/mohit838/learn-go-with-project/internal/auth/infrastructure"
	"github.com/mohit838/learn-go-with-project/internal/auth/transport"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// NewRouter initializes and returns the HTTP router with all endpoints and middleware
// Parameters:
//   - db: PostgreSQL database connection for primary data
//   - mongoDB: MongoDB database for audit logging (auth_logs collection)
//   - redisClient: Redis client for caching (Database 0)
//   - minioClient: MinIO client for object storage
func NewRouter(db *sql.DB, mongoDB *mongo.Database, redisClient *redis.Client, minioClient *minio.Client, minioBucket string, cfg config.Cfg) http.Handler {
	r := chi.NewRouter()

	// ========================
	// Middleware Stack
	// ========================
	// RequestID: Add unique request ID to each request
	r.Use(middleware.RequestID)
	// ClientIPFromRemoteAddr: Extract client IP from remote address
	r.Use(middleware.ClientIPFromRemoteAddr)
	// Logger: Log all HTTP requests
	r.Use(middleware.Logger)
	// Recoverer: Recover from panics and log them
	r.Use(middleware.Recoverer)
	// Timeout: Set 60 second timeout for all requests
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

	authRepo := infrastructure.NewAuthRepository(db)
	tokenService := application.NewTokenService(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		time.Duration(cfg.AccessTokenMinutes)*time.Minute,
		time.Duration(cfg.RefreshTokenHours)*time.Hour,
	)
	authService := application.NewAuthService(authRepo, tokenService)
	authHandler := transport.NewAuthHandler(authService)

	r.Post(constants.RouteRegister, authHandler.Register)
	r.Post(constants.RouteLogin, authHandler.Login)
	r.Group(func(r chi.Router) {
		r.Use(transport.RequireRole(tokenService, constants.DefaultRoleSuperadmin))
		r.Get(constants.RouteUsers, authHandler.ListUsers)
	})

	// ========================
	// MongoDB Endpoints
	// ========================
	// POST /logs - Create audit log entry in MongoDB
	// Example: curl -X POST http://localhost:8484/logs
	r.Post(constants.RouteLogs, func(w http.ResponseWriter, r *http.Request) {
		collection := mongoDB.Collection(constants.AuditLogCollection)
		_, err := collection.InsertOne(r.Context(), map[string]any{
			"action":    constants.DefaultLogAction,
			"timestamp": time.Now(),
		})
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to create log", err.Error())
			return
		}
		response.Success(w, http.StatusCreated, "log created", map[string]string{
			"collection": constants.AuditLogCollection,
		})
	})

	// ========================
	// Redis Cache Endpoints
	// ========================
	// POST /cache - Set a cache value (1 hour TTL)
	// Example: curl -X POST http://localhost:8484/cache
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
	// Example: curl http://localhost:8484/cache/test_key
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
