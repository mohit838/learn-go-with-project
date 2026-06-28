package constants

const (
	ServiceName  = "auth-service"
	ServiceTitle = "Auth Service"
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
	RouteRegister       = "/register"
	RouteLogin          = "/login"
	RouteUsers          = "/users"
	RouteLogs           = "/logs"
	RouteCache          = "/cache"
	RouteCacheKey       = "/cache/{key}"
)

const (
	AuditLogCollection    = "auth_logs"
	DefaultLogAction      = "login"
	DefaultCacheKey       = "test_key"
	DefaultCacheValue     = "test_value"
	DefaultRoleSuperadmin = "superadmin"
	DefaultRoleAdmin      = "admin"
	DefaultRoleOwner      = "owner"
	DefaultRoleStaff      = "staff"
	DefaultRoleGuest      = "guest"
)
