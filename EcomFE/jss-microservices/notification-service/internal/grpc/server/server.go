package server

import (
	"context"
	// "strconv"

	"github.com/linhhuynhcoding/jss-microservices/notification-service/internal/service"
	notificationpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/notification"
	"go.uber.org/zap"
)

// Server implements the gRPC NotificationServiceServer interface.  It
// delegates requests to an underlying NotificationService.
type Server struct {
    notificationpb.UnimplementedNotificationServiceServer
    svc *service.NotificationService
    log *zap.Logger
}

// NewServer constructs a new Server with the given service and logger.
func NewServer(svc *service.NotificationService, log *zap.Logger) *Server {
    return &Server{svc: svc, log: log}
}

// CreateNotification handles the RPC for creating notifications.
func (s *Server) CreateNotification(ctx context.Context, req *notificationpb.CreateNotificationRequest) (*notificationpb.Notification, error) {
    n, err := s.svc.Create(ctx, req.GetUserId(), req.GetRole(), req.GetTitle(), req.GetMessage())
    if err != nil {
        return nil, err
    }
    return &notificationpb.Notification{
        Id:        n.ID.Hex(),
        UserId:    n.UserID.Hex(),
        Role:      n.Role,
        Title:     n.Title,
        Message:   n.Message,
        IsRead:    n.IsRead,
        CreatedAt: n.CreatedAt.Unix(),
    }, nil
}

// ListNotifications returns a paginated list of notifications for a
// specific user or role.
func (s *Server) ListNotifications(ctx context.Context, req *notificationpb.ListNotificationsRequest) (*notificationpb.ListNotificationsResponse, error) {
    page := req.GetPage()
    pageSize := req.GetPageSize()
    if page == 0 {
        page = 1
    }
    if pageSize == 0 {
        pageSize = 10
    }
    notifs, total, err := s.svc.List(ctx, req.GetUserId(), req.GetRole(), int64(page), int64(pageSize))
    if err != nil {
        return nil, err
    }
    resp := &notificationpb.ListNotificationsResponse{
        Total: int32(total),
    }
    for _, n := range notifs {
        resp.Notifications = append(resp.Notifications, &notificationpb.Notification{
            Id:        n.ID.Hex(),
            UserId:    n.UserID.Hex(),
            Role:      n.Role,
            Title:     n.Title,
            Message:   n.Message,
            IsRead:    n.IsRead,
            CreatedAt: n.CreatedAt.Unix(),
        })
    }
    return resp, nil
}

// MarkAsRead marks a notification as read and returns the updated
// document.  If the notification does not exist a gRPC NotFound
// error will be returned.
func (s *Server) MarkAsRead(ctx context.Context, req *notificationpb.MarkAsReadRequest) (*notificationpb.Notification, error) {
    n, err := s.svc.MarkAsRead(ctx, req.GetNotificationId())
    if err != nil {
        return nil, err
    }
    return &notificationpb.Notification{
        Id:        n.ID.Hex(),
        UserId:    n.UserID.Hex(),
        Role:      n.Role,
        Title:     n.Title,
        Message:   n.Message,
        IsRead:    n.IsRead,
        CreatedAt: n.CreatedAt.Unix(),
    }, nil
}

// HealthCheck can be used by orchestration tools to verify the service is
// ready.  It simply returns an empty response.  Exposing a dedicated
// health endpoint avoids having to call application methods just to
// determine if the process is up.
func (s *Server) HealthCheck(ctx context.Context, req *notificationpb.Empty) (*notificationpb.Empty, error) {
    return &notificationpb.Empty{}, nil
}