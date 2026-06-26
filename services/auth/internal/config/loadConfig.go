package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Cfg struct {
	AppEnv     string
	AppName    string
	AppVersion string

	AppPort  string
	AppDebug bool

	LogLevel string

	CORSAllowedOrigins         []string
	CORSAllowCredentials       bool
	AuthRateLimitRequests      int
	AuthRateLimitWindowSeconds int

	JWTSecret          string
	JWTIssuer          string
	AccessTokenMinutes int
	RefreshTokenHours  int

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBURL      string

	RedisHost     string
	RedisPort     string
	RedisUser     string
	RedisPassword string
	RedisURL      string

	MongoHost     string
	MongoPort     string
	MongoUser     string
	MongoPassword string
	MongoDB       string
	MongoURL      string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIORegion    string
	MinIOUseSSL    bool
}

func LoadConfig(path string) (Cfg, error) {
	err := godotenv.Load(path)
	if err != nil && !os.IsNotExist(err) {
		return Cfg{}, err
	}

	cfg := Cfg{
		AppEnv:     getEnv("APP_ENV", "development"),
		AppName:    getEnv("APP_NAME", "auth-service"),
		AppVersion: getEnv("APP_VERSION", ""),

		AppPort:  getEnv("APP_PORT", ""),
		AppDebug: getEnvAsBool("APP_DEBUG", false),

		LogLevel: getEnv("APP_LOG_LEVEL", ""),

		CORSAllowedOrigins:         getEnvAsCSV("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		CORSAllowCredentials:       getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		AuthRateLimitRequests:      getEnvAsInt("AUTH_RATE_LIMIT_REQUESTS", 20),
		AuthRateLimitWindowSeconds: getEnvAsInt("AUTH_RATE_LIMIT_WINDOW_SECONDS", 60),

		JWTSecret:          getEnv("JWT_SECRET", "local-dev-secret-change-me"),
		JWTIssuer:          getEnv("JWT_ISSUER", "auth-service"),
		AccessTokenMinutes: getEnvAsInt("ACCESS_TOKEN_MINUTES", 15),
		RefreshTokenHours:  getEnvAsInt("REFRESH_TOKEN_HOURS", 168),

		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),
		DBURL:      getEnv("DATABASE_URL", ""),

		RedisHost:     getEnv("REDIS_HOST", ""),
		RedisPort:     getEnv("REDIS_PORT", ""),
		RedisUser:     getEnv("REDIS_USER", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisURL:      getEnv("REDIS_URL", ""),

		MongoHost:     getEnv("MONGO_HOST", ""),
		MongoPort:     getEnv("MONGO_PORT", ""),
		MongoUser:     getEnv("MONGO_USER", ""),
		MongoPassword: getEnv("MONGO_PASSWORD", ""),
		MongoDB:       getEnv("MONGO_DB", ""),
		MongoURL:      getEnv("MONGO_URL", ""),

		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", getEnv("S3_ENDPOINT", "")),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", getEnv("S3_ACCESS_KEY", "")),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", getEnv("S3_SECRET_KEY", "")),
		MinIOBucket:    getEnv("MINIO_BUCKET", getEnv("S3_BUCKET", "")),
		MinIORegion:    getEnv("MINIO_REGION", getEnv("S3_REGION", "us-east-1")),
		MinIOUseSSL:    getEnvAsBool("MINIO_USE_SSL", false),
	}

	return cfg, nil
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func getEnvAsCSV(key string, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	items := strings.Split(value, ",")
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			cleaned = append(cleaned, item)
		}
	}
	return cleaned
}
