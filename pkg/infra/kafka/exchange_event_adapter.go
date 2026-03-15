package kafka

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
)

// Compile-time interface compliance check
var _ exchange_out.ExchangeEventPublisher = (*ExchangeEventPublisherAdapter)(nil)

// ExchangeEventPublisherAdapter bridges the domain ExchangeEventPublisher port to Kafka
type ExchangeEventPublisherAdapter struct {
	client *Client
}

// NewExchangeEventPublisherAdapter creates a new exchange event publisher adapter
func NewExchangeEventPublisherAdapter(client *Client) *ExchangeEventPublisherAdapter {
	return &ExchangeEventPublisherAdapter{client: client}
}

func (a *ExchangeEventPublisherAdapter) publish(ctx context.Context, topic, eventType string, key uuid.UUID, data interface{}) error {
	if a.client == nil {
		return nil
	}

	msg := &Message{
		Key:   key.String(),
		Value: data,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": eventType,
		},
	}

	if err := a.client.Publish(ctx, topic, msg); err != nil {
		log.Printf("[ExchangeEventPublisher] Failed to publish %s to %s: %v", eventType, topic, err)
		return err
	}

	return nil
}

func (a *ExchangeEventPublisherAdapter) PublishOrderCreated(ctx context.Context, order *exchange_entities.Order) error {
	return a.publish(ctx, TopicExchangeOrderCreated, EventTypeOrderCreated, order.ID, order)
}

func (a *ExchangeEventPublisherAdapter) PublishOrderExecuting(ctx context.Context, order *exchange_entities.Order) error {
	return a.publish(ctx, TopicExchangeOrderExecuting, EventTypeOrderExecuting, order.ID, order)
}

func (a *ExchangeEventPublisherAdapter) PublishOrderFilled(ctx context.Context, order *exchange_entities.Order) error {
	return a.publish(ctx, TopicExchangeOrderFilled, EventTypeOrderFilled, order.ID, order)
}

func (a *ExchangeEventPublisherAdapter) PublishOrderFailed(ctx context.Context, order *exchange_entities.Order) error {
	return a.publish(ctx, TopicExchangeOrderFailed, EventTypeOrderFailed, order.ID, order)
}

func (a *ExchangeEventPublisherAdapter) PublishOrderCancelled(ctx context.Context, order *exchange_entities.Order) error {
	return a.publish(ctx, TopicExchangeOrderCancelled, EventTypeOrderCancelled, order.ID, order)
}

func (a *ExchangeEventPublisherAdapter) PublishQuoteCreated(ctx context.Context, quote *exchange_entities.Quote) error {
	return a.publish(ctx, TopicExchangeQuoteCreated, EventTypeQuoteCreated, quote.ID, quote)
}

func (a *ExchangeEventPublisherAdapter) PublishPriceUpdated(ctx context.Context, rate *exchange_entities.ExchangeRate) error {
	return a.publish(ctx, TopicExchangePriceUpdated, EventTypePriceUpdated, rate.ID, rate)
}
