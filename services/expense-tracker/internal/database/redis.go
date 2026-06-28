package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates and returns a Redis client connection
// Parses the Redis URL and establishes connection with health check
// Parameters:
//   - redisURL: Redis connection string (format: redis://user:pass@host:port/database)
//     Database number determined by the last segment (0-15)
//
// Returns:
//   - *redis.Client: Connected Redis client
//   - error: Connection error if any
func NewRedis(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Verify the connection is successful with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return client, nil
}
