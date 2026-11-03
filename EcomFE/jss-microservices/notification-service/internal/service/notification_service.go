package service

import (
    "context"
    "errors"

    "github.com/linhhuynhcoding/jss-microservices/notification-service/internal/domain"
    "github.com/linhhuynhcoding/jss-microservices/notification-service/internal/repository"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.uber.org/zap"
)

// NotificationService encapsulates business logic for creating,
// listing and updating notifications.  It delegates persistence to a
// NotificationRepository.
type NotificationService struct {
    repo *repository.NotificationRepository
    log  *zap.Logger
}

// NewNotificationService constructs a new service with the given
// repository and logger.  The logger may be shared across services.
func NewNotificationService(repo *repository.NotificationRepository, log *zap.Logger) *NotificationService {
    return &NotificationService{repo: repo, log: log}
}

// Create validates the input and inserts a new notification.  Either
// userIdStr or role must be provided.  When userIdStr is non‑empty it
// is converted from a hexadecimal string to a MongoDB ObjectID.
func (s *NotificationService) Create(ctx context.Context, userIdStr, role, title, message string) (*domain.Notification, error) {
    var userID primitive.ObjectID
    var err error
    if userIdStr != "" {
        userID, err = primitive.ObjectIDFromHex(userIdStr)
        if err != nil {
            return nil, errors.New("invalid user id")
        }
    }
    if userIdStr == "" && role == "" {
        return nil, errors.New("either userId or role must be specified")
    }
    n := &domain.Notification{
        UserID:  userID,
        Role:    role,
        Title:   title,
        Message: message,
        IsRead:  false,
    }
    return s.repo.Create(ctx, n)
}

// List retrieves notifications for the specified user or role.  Page
// numbering starts at 1.  Passing zero or negative values will
// default to page 1 and a page size of 10.  A total count of
// matching documents is also returned.
func (s *NotificationService) List(ctx context.Context, userIdStr, role string, page, pageSize int64) ([]*domain.Notification, int64, error) {
    if page <= 0 {
        page = 1
    }
    if pageSize <= 0 {
        pageSize = 10
    }
    var userID primitive.ObjectID
    var err error
    if userIdStr != "" {
        userID, err = primitive.ObjectIDFromHex(userIdStr)
        if err != nil {
            return nil, 0, errors.New("invalid user id")
        }
    }
    offset := (page - 1) * pageSize
    return s.repo.FindByUser(ctx, userID, role, offset, pageSize)
}

// MarkAsRead marks an existing notification as read.  It returns
// the updated notification or an error if the ID is invalid or no
// document matches.
func (s *NotificationService) MarkAsRead(ctx context.Context, idStr string) (*domain.Notification, error) {
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        return nil, errors.New("invalid notification id")
    }
    return s.repo.MarkAsRead(ctx, oid)
}