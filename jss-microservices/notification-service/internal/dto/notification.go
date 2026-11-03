package dto

// CreateNotificationRequest is used by the service layer to accept
// parameters for creating a new notification.  Either UserID or Role
// should be provided.  If both are empty the notification will not
// be targeted to any recipient.
type CreateNotificationRequest struct {
    UserID  string `json:"userId,omitempty"`
    Role    string `json:"role,omitempty"`
    Title   string `json:"title"`
    Message string `json:"message"`
}

// NotificationResponse is returned by the service layer and maps to
// the gRPC/HTTP representation.  Timestamps are encoded as Unix
// seconds.
type NotificationResponse struct {
    ID        string `json:"id"`
    UserID    string `json:"userId"`
    Role      string `json:"role"`
    Title     string `json:"title"`
    Message   string `json:"message"`
    IsRead    bool   `json:"isRead"`
    CreatedAt int64  `json:"createdAt"`
}