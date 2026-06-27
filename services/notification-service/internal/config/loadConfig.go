package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Cfg struct {
	AppName    string
	AppEnv     string
	AppVersion string
	AppPort    string
	GRPCPort   string
	AppDebug   bool
	LogLevel   string
	Workers    int
	QueueSize  int
}

func LoadConfig(path string) (Cfg, error) {
	_ = godotenv.Load(path)

	return Cfg{
		AppName:    getEnv("APP_NAME", "notification-service"),
		AppEnv:     getEnv("APP_ENV", "development"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		AppPort:    getEnv("APP_PORT", "8487"),
		GRPCPort:   getEnv("GRPC_PORT", "8587"),
		AppDebug:   getBool("APP_DEBUG", true),
		LogLevel:   getEnv("APP_LOG_LEVEL", "debug"),
		Workers:    getInt("NOTIFICATION_WORKERS", 3),
		QueueSize:  getInt("NOTIFICATION_QUEUE_SIZE", 50),
	}, nil
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func getBool(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
