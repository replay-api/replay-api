package routing

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ─── Notification Entity ────────────────────────────────────────────────────

// NotificationType defines the category of notification
type NotificationType string

const (
	NotificationTypeMatch       NotificationType = "match"
	NotificationTypeTeam        NotificationType = "team"
	NotificationTypeFriend      NotificationType = "friend"
	NotificationTypeSystem      NotificationType = "system"
	NotificationTypeAchievement NotificationType = "achievement"
	NotificationTypeMessage     NotificationType = "message"
)

// Notification represents a user notification stored in MongoDB
type Notification struct {
	ID        uuid.UUID              `json:"id" bson:"_id"`
	UserID    uuid.UUID              `json:"user_id" bson:"user_id"`
	Type      NotificationType       `json:"type" bson:"type"`
	Title     string                 `json:"title" bson:"title"`
	Message   string                 `json:"message" bson:"message"`
	Timestamp string                 `json:"timestamp" bson:"timestamp"`
	Read      bool                   `json:"read" bson:"read"`
	ActionURL string                 `json:"actionUrl,omitempty" bson:"action_url,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" bson:"updated_at"`
}

// NotificationsResponse is the list response envelope
type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"total_count"`
	UnreadCount   int            `json:"unread_count"`
}

// ─── Handler ────────────────────────────────────────────────────────────────

// NotificationHandler provides CRUD endpoints for notifications backed by MongoDB
type NotificationHandler struct {
	collection *mongo.Collection
}

// NewNotificationHandler creates a real notification handler with MongoDB persistence
func NewNotificationHandler(client *mongo.Client, dbName string) *NotificationHandler {
	col := client.Database(dbName).Collection("notifications")

	// Create indexes in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_user_created"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "read", Value: 1}},
			Options: options.Index().SetName("idx_user_read"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "type", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_user_type_created"),
		},
	}

	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create notification indexes", "error", err)
	}

	return &NotificationHandler{collection: col}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func getUserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}

func isAuthenticated(r *http.Request) bool {
	authenticated, ok := r.Context().Value(shared.AuthenticatedKey).(bool)
	return ok && authenticated
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"success": false,
		"error":   code,
		"message": message,
	})
}

// ─── Endpoints ──────────────────────────────────────────────────────────────

// ListNotifications returns paginated notifications for the authenticated user
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Parse query params
	filter := bson.M{"user_id": userID}

	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		filter["type"] = typeParam
	}
	if readParam := r.URL.Query().Get("read"); readParam != "" {
		if readParam == "true" {
			filter["read"] = true
		} else if readParam == "false" {
			filter["read"] = false
		}
	}

	limit := int64(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := int64(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Find notifications
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := h.collection.Find(ctx, filter, opts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list notifications", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch notifications")
		return
	}
	defer cursor.Close(ctx)

	notifications := []Notification{}
	if err := cursor.All(ctx, &notifications); err != nil {
		slog.ErrorContext(ctx, "Failed to decode notifications", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to decode notifications")
		return
	}

	// Get total count
	totalCount, _ := h.collection.CountDocuments(ctx, bson.M{"user_id": userID})

	// Get unread count
	unreadCount, _ := h.collection.CountDocuments(ctx, bson.M{"user_id": userID, "read": false})

	writeJSON(w, http.StatusOK, NotificationsResponse{
		Notifications: notifications,
		TotalCount:    int(totalCount),
		UnreadCount:   int(unreadCount),
	})
}

// GetNotification returns a single notification by ID
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	notifID, err := uuid.Parse(vars["notification_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid notification ID format")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var notification Notification
	err = h.collection.FindOne(ctx, bson.M{"_id": notifID, "user_id": userID}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "not_found", "Notification not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch notification")
		return
	}

	writeJSON(w, http.StatusOK, notification)
}

// MarkAsRead marks a single notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	notifID, err := uuid.Parse(vars["notification_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid notification ID format")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result, err := h.collection.UpdateOne(ctx,
		bson.M{"_id": notifID, "user_id": userID},
		bson.M{"$set": bson.M{"read": true, "updated_at": time.Now()}},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to mark notification as read")
		return
	}

	if result.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "not_found", "Notification not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// MarkAllAsRead marks all notifications as read for the user
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, err := h.collection.UpdateMany(ctx,
		bson.M{"user_id": userID, "read": false},
		bson.M{"$set": bson.M{"read": true, "updated_at": time.Now()}},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to mark all as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DeleteNotification deletes a single notification
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	notifID, err := uuid.Parse(vars["notification_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid notification ID format")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result, err := h.collection.DeleteOne(ctx, bson.M{"_id": notifID, "user_id": userID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete notification")
		return
	}

	if result.DeletedCount == 0 {
		writeError(w, http.StatusNotFound, "not_found", "Notification not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DeleteAllNotifications deletes all notifications for the user
func (h *NotificationHandler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, err := h.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetUnreadCount returns just the unread count for the user
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	count, err := h.collection.CountDocuments(ctx, bson.M{"user_id": userID, "read": false})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to count unread")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"count": count})
}

// ─── Internal: Create notification (used by WebSocket hub / services) ───────

// CreateNotification inserts a new notification and returns it
func (h *NotificationHandler) CreateNotification(ctx context.Context, userID uuid.UUID, notifType NotificationType, title, message string, actionURL string, metadata map[string]interface{}) (*Notification, error) {
	now := time.Now()
	notification := &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Timestamp: now.Format(time.RFC3339),
		Read:      false,
		ActionURL: actionURL,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := h.collection.InsertOne(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create notification", "error", err, "user_id", userID)
		return nil, err
	}

	return notification, nil
}

// ─── Seed: Create sample notifications for testing ──────────────────────────

// SeedNotifications creates sample notifications for a user (for development/testing)
func (h *NotificationHandler) SeedNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok || !isAuthenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	now := time.Now()
	samples := []Notification{
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeMatch, Title: "Match Found!",
			Message: "A competitive CS2 match has been found. Accept within 30 seconds.", Timestamp: now.Add(-2 * time.Minute).Format(time.RFC3339),
			Read: false, ActionURL: "/matchmaking",
			Metadata:  map[string]interface{}{"icon": "solar:gameboy-bold", "game": "cs2", "elo": "2100"},
			CreatedAt: now.Add(-2 * time.Minute), UpdatedAt: now.Add(-2 * time.Minute),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeTeam, Title: "Team Invite: Nova Esports",
			Message: "You've been invited to join Nova Esports as a rifler.", Timestamp: now.Add(-15 * time.Minute).Format(time.RFC3339),
			Read: false, ActionURL: "/teams/invite/123",
			Metadata:  map[string]interface{}{"icon": "solar:users-group-rounded-bold", "team_name": "Nova Esports"},
			CreatedAt: now.Add(-15 * time.Minute), UpdatedAt: now.Add(-15 * time.Minute),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeFriend, Title: "Friend Request",
			Message: "xAceKiller wants to add you as a friend.", Timestamp: now.Add(-1 * time.Hour).Format(time.RFC3339),
			Read: false,
			Metadata:  map[string]interface{}{"icon": "solar:user-plus-bold", "player_name": "xAceKiller"},
			CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeAchievement, Title: "Achievement Unlocked!",
			Message: "You earned 'First Blood' — Get the first kill in 10 competitive matches.", Timestamp: now.Add(-3 * time.Hour).Format(time.RFC3339),
			Read: false, ActionURL: "/profile/achievements",
			Metadata:  map[string]interface{}{"icon": "solar:cup-star-bold", "achievement_name": "First Blood", "rarity": "rare"},
			CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeSystem, Title: "Platform Maintenance",
			Message: "Scheduled maintenance on Feb 22, 2026 from 02:00-04:00 UTC. Matchmaking will be temporarily unavailable.", Timestamp: now.Add(-6 * time.Hour).Format(time.RFC3339),
			Read: true,
			Metadata:  map[string]interface{}{"icon": "solar:bell-bold"},
			CreatedAt: now.Add(-6 * time.Hour), UpdatedAt: now.Add(-5 * time.Hour),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeMessage, Title: "New Message from Coach_Mike",
			Message: "Hey, let's review the VOD from yesterday's scrim. I've timestamped the key rounds.", Timestamp: now.Add(-12 * time.Hour).Format(time.RFC3339),
			Read: true, ActionURL: "/messages/456",
			Metadata:  map[string]interface{}{"icon": "solar:chat-round-bold", "sender": "Coach_Mike"},
			CreatedAt: now.Add(-12 * time.Hour), UpdatedAt: now.Add(-11 * time.Hour),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeMatch, Title: "Ready Check — All Confirmed!",
			Message: "All 10 players confirmed. Game server is being prepared.", Timestamp: now.Add(-30 * time.Minute).Format(time.RFC3339),
			Read: true,
			Metadata:  map[string]interface{}{"icon": "solar:shield-check-bold"},
			CreatedAt: now.Add(-30 * time.Minute), UpdatedAt: now.Add(-28 * time.Minute),
		},
		{
			ID: uuid.New(), UserID: userID,
			Type: NotificationTypeAchievement, Title: "Season Reward: Gold Division",
			Message: "Congratulations! You finished Season 4 in Gold Division. Claim your exclusive rewards.", Timestamp: now.Add(-24 * time.Hour).Format(time.RFC3339),
			Read: false, ActionURL: "/rewards/season-4",
			Metadata:  map[string]interface{}{"icon": "solar:medal-ribbons-star-bold", "division": "gold", "season": "4"},
			CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour),
		},
	}

	docs := make([]interface{}, len(samples))
	for i := range samples {
		docs[i] = samples[i]
	}

	_, err := h.collection.InsertMany(ctx, docs)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to seed notifications", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to seed notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(samples),
		"message": "Seeded sample notifications",
	})
}
