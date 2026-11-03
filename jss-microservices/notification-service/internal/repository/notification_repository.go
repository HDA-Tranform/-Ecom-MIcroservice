package repository

import (
    "context"
    "errors"
    "time"

    "github.com/linhhuynhcoding/jss-microservices/notification-service/internal/domain"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"
)

// NotificationRepository provides CRUD operations on notifications.
type NotificationRepository struct {
    coll *mongo.Collection
    log  *zap.Logger
}

// NewNotificationRepository creates a repository bound to the
// notifications collection.  The collection will be created
// automatically on first insert.
func NewNotificationRepository(db *mongo.Database, log *zap.Logger) *NotificationRepository {
    return &NotificationRepository{
        coll: db.Collection("notifications"),
        log:  log,
    }
}

// Create inserts a new notification document and populates the ID and
// CreatedAt fields.  It returns a copy of the inserted document.
func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
    n.CreatedAt = time.Now()
    res, err := r.coll.InsertOne(ctx, n)
    if err != nil {
        return nil, err
    }
    if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
        n.ID = oid
    }
    return n, nil
}

// FindByUser returns notifications belonging to a specific user ID or
// role.  Results are sorted by creation time descending.  The total
// number of documents matching the filter is also returned.  If
// userID is zero then only the role filter is applied.  If both are
// empty all notifications are returned.
func (r *NotificationRepository) FindByUser(ctx context.Context, userID primitive.ObjectID, role string, offset, limit int64) ([]*domain.Notification, int64, error) {
    filter := bson.M{}
    if userID != primitive.NilObjectID {
        filter["userId"] = userID
    }
    if role != "" {
        filter["role"] = role
    }
    count, err := r.coll.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }
    opts := options.Find().
        SetSort(bson.D{{Key: "createdAt", Value: -1}}).
        SetSkip(offset).
        SetLimit(limit)
    cursor, err := r.coll.Find(ctx, filter, opts)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)
    var notifications []*domain.Notification
    for cursor.Next(ctx) {
        var n domain.Notification
        if err := cursor.Decode(&n); err != nil {
            return nil, 0, err
        }
        notifications = append(notifications, &n)
    }
    if err := cursor.Err(); err != nil {
        return nil, 0, err
    }
    return notifications, count, nil
}

// MarkAsRead sets the isRead flag on a notification by ID and
// returns the updated document.  If no document matches the ID an
// error is returned.
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id primitive.ObjectID) (*domain.Notification, error) {
    filter := bson.M{"_id": id}
    update := bson.M{"$set": bson.M{"isRead": true}}
    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    var n domain.Notification
    if err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&n); err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, mongo.ErrNoDocuments
        }
        return nil, err
    }
    return &n, nil
}