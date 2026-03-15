package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	kafka_go "github.com/segmentio/kafka-go"

	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
	prediction_out "github.com/replay-api/replay-api/pkg/domain/prediction/ports/out"
)

type PredictionEventPublisherAdapter struct {
	client *Client
}

func NewPredictionEventPublisherAdapter(client *Client) prediction_out.PredictionEventPublisher {
	return &PredictionEventPublisherAdapter{client: client}
}

type PredictionEvent struct {
	EventType string      `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

func (p *PredictionEventPublisherAdapter) publishPrediction(ctx context.Context, topic string, key string, event PredictionEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal prediction event", "error", err, "event_type", event.EventType)
		return fmt.Errorf("failed to marshal prediction event: %w", err)
	}

	msg := kafka_go.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	writer := p.client.GetWriter(topic)
	if writer == nil {
		slog.WarnContext(ctx, "Kafka writer not available for prediction topic", "topic", topic)
		return nil
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "Failed to publish prediction event", "error", err, "topic", topic, "event_type", event.EventType)
		return fmt.Errorf("failed to publish prediction event: %w", err)
	}

	slog.InfoContext(ctx, "Published prediction event", "topic", topic, "event_type", event.EventType, "key", key)
	return nil
}

func (p *PredictionEventPublisherAdapter) PublishBetPlaced(ctx context.Context, bet *prediction_entities.Bet) error {
	return p.publishPrediction(ctx, TopicPredictionPlaced, bet.MarketID.String(), PredictionEvent{
		EventType: EventTypePredictionPlaced,
		Timestamp: time.Now(),
		Payload:   bet,
	})
}

func (p *PredictionEventPublisherAdapter) PublishMarketResolved(ctx context.Context, market *prediction_entities.PredictionMarket) error {
	return p.publishPrediction(ctx, TopicPredictionResolved, market.ID.String(), PredictionEvent{
		EventType: EventTypePredictionResolved,
		Timestamp: time.Now(),
		Payload:   market,
	})
}
