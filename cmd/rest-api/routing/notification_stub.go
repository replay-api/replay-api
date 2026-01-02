package routing

import (
	"encoding/json"
	"net/http"
)

// NotificationStubHandler provides stub endpoints for notifications
// This allows the frontend to function while the full notification system is implemented
type NotificationStubHandler struct{}

// NewNotificationStubHandler creates a new notification stub handler
func NewNotificationStubHandler() *NotificationStubHandler {
	return &NotificationStubHandler{}
}

// NotificationsResponse represents the response structure for listing notifications
type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"total_count"`
	UnreadCount   int            `json:"unread_count"`
}

// Notification represents a single notification
type Notification struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Read      bool              `json:"read"`
	ActionURL string            `json:"actionUrl,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ListNotifications returns an empty list of notifications
func (h *NotificationStubHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	response := NotificationsResponse{
		Notifications: []Notification{},
		TotalCount:    0,
		UnreadCount:   0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetNotification returns a 404 for individual notification requests
func (h *NotificationStubHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "not_found",
		"message": "Notification not found",
	})
}

// MarkAsRead handles marking a single notification as read
func (h *NotificationStubHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

// MarkAllAsRead handles marking all notifications as read
func (h *NotificationStubHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
