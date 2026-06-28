package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoDB creates and returns a MongoDB client connection
// Establishes connection to MongoDB with a 10-second timeout
// Parameters:
//   - mongoURL: MongoDB connection string (format: mongodb://user:pass@host:port/database?authSource=appdb)
//
// Returns:
//   - *mongo.Client: Connected MongoDB client
//   - error: Connection error if any
func NewMongoDB(mongoURL string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	// Verify the connection is successful
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetAuthDB returns the MongoDB database instance
// This is a generic function that can be used for any database in the MongoDB cluster
// Parameters:
//   - client: Connected MongoDB client
//   - dbName: Database name (e.g., auth_log_db, task_log_db, expense_log_db)
//
// Returns:
//   - *mongo.Database: MongoDB database instance
func GetAuthDB(client *mongo.Client, dbName string) *mongo.Database {
	return client.Database(dbName)
}
