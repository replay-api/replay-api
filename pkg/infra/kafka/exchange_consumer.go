package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
	exchange_services "github.com/replay-api/replay-api/pkg/domain/exchange/services"
)

// ExchangeConsumer handles exchange-related Kafka events
type ExchangeConsumer struct {
	orderService *exchange_services.OrderService
	consumer     *Consumer
}

// NewExchangeConsumer creates a new exchange consumer
func NewExchangeConsumer(client *Client, groupID string, orderService *exchange_services.OrderService) *ExchangeConsumer {
	topics := []string{
		TopicExchangeOrderCreated,
		TopicExchangeOrderFailed,
	}
	config := DefaultConsumerConfig(groupID, topics)
	consumer := NewConsumer(client, config)

	ec := &ExchangeConsumer{
		orderService: orderService,
		consumer:     consumer,
	}

	consumer.RegisterHandler(TopicExchangeOrderCreated, ec.handleOrderCreated)
	consumer.RegisterHandler(TopicExchangeOrderFailed, ec.handleOrderFailed)

	return ec
}

// Start starts the exchange consumer
func (c *ExchangeConsumer) Start(ctx context.Context) error {
	return c.consumer.Start(ctx)
}

// Close closes the exchange consumer
func (c *ExchangeConsumer) Close() error {
	return c.consumer.Close()
}

// handleOrderCreated processes a newly created order by executing it on the exchange
func (c *ExchangeConsumer) handleOrderCreated(ctx context.Context, msg *kafka.Message) error {
	var event struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	orderID, err := uuid.Parse(event.ID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	log.Printf("[ExchangeConsumer] Processing order %s", orderID)

	if err := c.orderService.ExecuteOrderOnExchange(ctx, orderID); err != nil {
		log.Printf("[ExchangeConsumer] Failed to execute order %s: %v", orderID, err)
		return err
	}

	log.Printf("[ExchangeConsumer] Successfully executed order %s", orderID)
	return nil
}

// handleOrderFailed handles failed orders with retry logic
func (c *ExchangeConsumer) handleOrderFailed(ctx context.Context, msg *kafka.Message) error {
	var event struct {
		ID         string `json:"id"`
		RetryCount int    `json:"retry_count"`
		MaxRetries int    `json:"max_retries"`
	}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order failed event: %w", err)
	}

	log.Printf("[ExchangeConsumer] Order %s failed (retry %d/%d)", event.ID, event.RetryCount, event.MaxRetries)

	if event.RetryCount < event.MaxRetries {
		orderID, err := uuid.Parse(event.ID)
		if err != nil {
			return err
		}
		log.Printf("[ExchangeConsumer] Retrying order %s", event.ID)
		return c.orderService.ExecuteOrderOnExchange(ctx, orderID)
	}

	log.Printf("[ExchangeConsumer] Order %s exhausted retries, sending to DLQ", event.ID)
	return nil
}
