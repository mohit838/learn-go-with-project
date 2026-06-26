package utils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Event struct {
	Action    string         `bson:"action" json:"action"`
	Metadata  map[string]any `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp time.Time      `bson:"timestamp" json:"timestamp"`
}

type EventLogger interface {
	Log(ctx context.Context, event Event) error
}

type MongoEventLogger struct {
	collection *mongo.Collection
}

func NewMongoEventLogger(db *mongo.Database, collectionName string) *MongoEventLogger {
	return &MongoEventLogger{
		collection: db.Collection(collectionName),
	}
}

func (l *MongoEventLogger) Log(ctx context.Context, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	_, err := l.collection.InsertOne(ctx, event)
	return err
}
