package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
)

// Ensure interface compliance
var _ analytics_out.ViewEventPublisher = (*ViewEventPublisherAdapter)(nil)

// ViewEventPublisherAdapter implements analytics_out.ViewEventPublisher using Kafka
type ViewEventPublisherAdapter struct {
	publisher *EventPublisher
}

// NewViewEventPublisherAdapter creates a new adapter for view event publishing
func NewViewEventPublisherAdapter(publisher *EventPublisher) analytics_out.ViewEventPublisher {
	return &ViewEventPublisherAdapter{publisher: publisher}
}

// EntityViewedEvent is the Kafka event struct for entity views
type EntityViewedEvent struct {
	EventID      uuid.UUID         `json:"event_id"`
	EntityID     uuid.UUID         `json:"entity_id"`
	EntityType   string            `json:"entity_type"`
	ViewerID     *uuid.UUID        `json:"viewer_id,omitempty"`
	ViewerType   string            `json:"viewer_type"`
	SessionID    string            `json:"session_id"`
	DeviceType   string            `json:"device_type"`
	GeoRegion    string            `json:"geo_region"`
	ReferrerType string            `json:"referrer_type"`
	ViewedAt     time.Time         `json:"viewed_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PublishEntityViewed publishes an entity viewed event to Kafka
func (a *ViewEventPublisherAdapter) PublishEntityViewed(ctx context.Context, view *analytics_entities.EntityView) error {
	event := &EntityViewedEvent{
		EventID:      uuid.New(),
		EntityID:     view.EntityID,
		EntityType:   string(view.EntityType),
		ViewerID:     view.ViewerID,
		ViewerType:   string(view.ViewerType),
		SessionID:    view.SessionID,
		DeviceType:   string(view.DeviceType),
		GeoRegion:    view.GeoRegion,
		ReferrerType: string(view.ReferrerType),
		ViewedAt:     view.ViewedAt,
	}

	msg := &Message{
		Key:       view.EntityID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":  EventTypeEntityViewed,
			"entity_type": string(view.EntityType),
		},
	}

	err := a.publisher.client.Publish(ctx, TopicAnalyticsEntityViewed, msg)
	if err != nil {
		slog.Error("failed to publish entity viewed event",
			"err", err,
			"entity_id", view.EntityID,
			"entity_type", view.EntityType)
	}
	return err
}
