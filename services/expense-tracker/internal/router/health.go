package router

import (
	"database/sql"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type healthDependencies struct {
	db           *sql.DB
	mongoDB      *mongo.Database
	redisClient  *redis.Client
	minioClient  *minio.Client
	minioBucket  string
	serviceName  string
	serviceTitle string
}

func registerHealthRoutes(r interface {
	Get(pattern string, handlerFn http.HandlerFunc)
}, deps healthDependencies) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, deps.serviceTitle+" API is running", map[string]string{
			"service": deps.serviceName,
			"status":  "ok",
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "service is healthy", map[string]string{
			"service": deps.serviceName,
			"status":  "ok",
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := runReadinessChecks(r, deps)
		if !checks["ready"].(bool) {
			response.JSON(w, http.StatusServiceUnavailable, response.Body{
				Success: false,
				Message: "service is not ready",
				Data:    checks,
			})
			return
		}
		response.Success(w, http.StatusOK, "service is ready", checks)
	})

	r.Get("/health/postgres", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.db.PingContext(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "PostgreSQL health check failed", err.Error())
			return
		}
		response.Success(w, http.StatusOK, "PostgreSQL connected", map[string]string{"status": "ok"})
	})

	r.Get("/health/redis", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.redisClient.Ping(r.Context()).Err(); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "Redis health check failed", err.Error())
			return
		}
		response.Success(w, http.StatusOK, "Redis connected", map[string]string{"status": "ok"})
	})

	r.Get("/health/mongo", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.mongoDB.Client().Ping(r.Context(), nil); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "MongoDB health check failed", err.Error())
			return
		}
		response.Success(w, http.StatusOK, "MongoDB connected", map[string]string{"status": "ok"})
	})

	r.Get("/health/minio", func(w http.ResponseWriter, r *http.Request) {
		if err := checkMinIO(r, deps); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "MinIO health check failed", err.Error())
			return
		}
		response.Success(w, http.StatusOK, "MinIO connected", map[string]string{
			"bucket": deps.minioBucket,
			"status": "ok",
		})
	})
}

func runReadinessChecks(r *http.Request, deps healthDependencies) map[string]any {
	checks := map[string]any{
		"service":  deps.serviceName,
		"ready":    true,
		"postgres": "ok",
		"redis":    "ok",
		"mongo":    "ok",
		"minio":    "ok",
	}

	if err := deps.db.PingContext(r.Context()); err != nil {
		checks["ready"] = false
		checks["postgres"] = err.Error()
	}
	if err := deps.redisClient.Ping(r.Context()).Err(); err != nil {
		checks["ready"] = false
		checks["redis"] = err.Error()
	}
	if err := deps.mongoDB.Client().Ping(r.Context(), nil); err != nil {
		checks["ready"] = false
		checks["mongo"] = err.Error()
	}
	if err := checkMinIO(r, deps); err != nil {
		checks["ready"] = false
		checks["minio"] = err.Error()
	}

	return checks
}

func checkMinIO(r *http.Request, deps healthDependencies) error {
	exists, err := deps.minioClient.BucketExists(r.Context(), deps.minioBucket)
	if err != nil {
		return err
	}
	if !exists {
		return minioBucketNotFoundError{bucket: deps.minioBucket}
	}
	return nil
}

type minioBucketNotFoundError struct {
	bucket string
}

func (e minioBucketNotFoundError) Error() string {
	return "bucket not found: " + e.bucket
}
