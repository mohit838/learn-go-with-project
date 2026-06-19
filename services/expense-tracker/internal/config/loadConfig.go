package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Cfg struct {
	AppEnv     string
	AppName    string
	AppVersion string

	AppPort  string
	AppDebug bool

	LogLevel string

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
}

func LoadConfig(path string) (Cfg, error) {
	err := godotenv.Load(path)
	if err != nil {
		return Cfg{}, err
	}

	cfg := Cfg{
		AppEnv:     getEnv("APP_ENV", "development"),
		AppName:    getEnv("APP_NAME", "expense-tracker"),
		AppVersion: getEnv("APP_VERSION", ""),

		AppPort:  getEnv("APP_PORT", ""),
		AppDebug: getEnvAsBool("APP_DEBUG", false),

		LogLevel: getEnv("APP_LOG_LEVEL", ""),

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
