package database

import (
    "context"
    "time"

    "github.com/linhhuynhcoding/jss-microservices/notification-service/pkg/config"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"
)

// Connect establishes a connection to MongoDB using the URI provided in
// the configuration.  It returns a mongo.Database instance scoped to
// the configured database name.  The function will terminate the
// application if a connection cannot be established or the ping fails.
func Connect(cfg *config.Config, log *zap.Logger) *mongo.Database {
    opts := options.Client().ApplyURI(cfg.MongoURI)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, opts)
    if err != nil {
        log.Fatal("failed to connect to MongoDB", zap.Error(err))
    }
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("failed to ping MongoDB", zap.Error(err))
    }
    db := client.Database(cfg.MongoDB)
    return db
}