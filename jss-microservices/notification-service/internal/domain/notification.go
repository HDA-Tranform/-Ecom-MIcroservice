package domain

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a single message intended for a user or role.
// It is stored in MongoDB and keyed by a BSON ObjectID.  The UserID
// field may be zero if the notification is broadcast to a role.
type Notification struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    UserID    primitive.ObjectID `bson:"userId,omitempty"`
    Role      string             `bson:"role,omitempty"`
    Title     string             `bson:"title"`
    Message   string             `bson:"message"`
    IsRead    bool               `bson:"isRead"`
    CreatedAt time.Time          `bson:"createdAt"`
}