package constants

const (
	ContentTypeJSON = "application/json"
	RootPath        = "/"
	HealthPath      = "/health"
	ReadyPath       = "/ready"

	ErrorInternalServer   = "internal_server_error"
	ErrorBadRequest       = "bad_request"
	ErrorConflict         = "conflict"
	ErrorNotFound         = "not_found"
	ErrorMethodNotAllowed = "method_not_allowed"
)
