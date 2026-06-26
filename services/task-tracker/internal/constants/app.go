package constants

const (
	ServiceName  = "task-tracker-service"
	ServiceTitle = "Task Tracker Service"
)

const (
	DefaultStatusOK        = "ok"
	DefaultPage            = 1
	DefaultPerPage         = 20
	DefaultMaxPerPage      = 100
	DefaultSuccessMessage  = "request completed successfully"
	DefaultNotReadyMessage = "service is not ready"
	DefaultReadyMessage    = "service is ready"
	DefaultHealthyMessage  = "service is healthy"
)

const (
	RouteRoot           = "/"
	RouteHealth         = "/health"
	RouteReady          = "/ready"
	RouteHealthPostgres = "/health/postgres"
	RouteHealthRedis    = "/health/redis"
	RouteHealthMongo    = "/health/mongo"
	RouteHealthMinIO    = "/health/minio"
	RouteLogs           = "/tasks-log"
	RouteCache          = "/cache"
	RouteCacheKey       = "/cache/{key}"
)

const (
	AuditLogCollection = "task_logs"
	DefaultLogAction   = "task_created"
	DefaultCacheKey    = "task_key"
	DefaultCacheValue  = "task_value"
)
