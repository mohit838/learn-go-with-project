package audit

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct{ collection *mongo.Collection }

func Connect(ctx context.Context, mongoURL, database string) (*mongo.Client, *Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, err
	}
	return client, &Store{collection: client.Database(database).Collection("audit_logs")}, nil
}

func (s *Store) Record(ctx context.Context, service, requestID, method, path string, status int, duration time.Duration) {
	if s == nil {
		return
	}
	_, _ = s.collection.InsertOne(ctx, bson.M{"service": service, "request_id": requestID, "method": method, "path": path, "status": status, "duration_ms": duration.Milliseconds(), "created_at": time.Now().UTC()})
}
