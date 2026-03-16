package kafka

import (
	"context"

	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
	prediction_out "github.com/replay-api/replay-api/pkg/domain/prediction/ports/out"
)

// NoopPredictionEventPublisher is a no-op prediction publisher for local and test environments without Kafka.
type NoopPredictionEventPublisher struct{}

// NewNoopPredictionEventPublisher creates a no-op prediction event publisher.
func NewNoopPredictionEventPublisher() prediction_out.PredictionEventPublisher {
	return &NoopPredictionEventPublisher{}
}

func (p *NoopPredictionEventPublisher) PublishBetPlaced(context.Context, *prediction_entities.Bet) error {
	return nil
}

func (p *NoopPredictionEventPublisher) PublishMarketResolved(context.Context, *prediction_entities.PredictionMarket) error {
	return nil
}
