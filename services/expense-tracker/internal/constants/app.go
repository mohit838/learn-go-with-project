package constants

const (
	ServiceName  = "expense-tracker-service"
	ServiceTitle = "Expense Tracker Service"
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
	RouteLogs           = "/expense-log"
	RouteCache          = "/cache"
	RouteCacheKey       = "/cache/{key}"
)

const (
	AuditLogCollection = "expense_logs"
	DefaultLogAction   = "expense_recorded"
	DefaultCacheKey    = "expense_key"
	DefaultCacheValue  = "expense_value"
)
