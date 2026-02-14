package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/segmentio/kafka-go"
)

// BillingConsumer processes billing events from Kafka
type BillingConsumer struct {
	consumer           *Consumer
	subscriptionWriter billing_out.SubscriptionWriter
	subscriptionReader billing_out.SubscriptionReader
	planReader         billing_out.PlanReader
	eventPublisher     *EventPublisher
}

// BillingConsumerConfig holds configuration for billing consumer
type BillingConsumerConfig struct {
	GroupID            string
	SubscriptionWriter billing_out.SubscriptionWriter
	SubscriptionReader billing_out.SubscriptionReader
	PlanReader         billing_out.PlanReader
	EventPublisher     *EventPublisher
}

// NewBillingConsumer creates a new billing consumer
func NewBillingConsumer(client *Client, config *BillingConsumerConfig) *BillingConsumer {
	topics := []string{
		TopicBillingSubscriptionCreated,
		TopicBillingSubscriptionUpgraded,
		TopicBillingSubscriptionCancelled,
		TopicBillingSubscriptionExpired,
		TopicBillingPaymentProcessed,
		TopicBillingPaymentFailed,
		TopicBillingPaymentRefunded,
	}

	consumerConfig := DefaultConsumerConfig(config.GroupID, topics)
	consumer := NewConsumer(client, consumerConfig)

	bc := &BillingConsumer{
		consumer:           consumer,
		subscriptionWriter: config.SubscriptionWriter,
		subscriptionReader: config.SubscriptionReader,
		planReader:         config.PlanReader,
		eventPublisher:     config.EventPublisher,
	}

	// Register handlers for each billing topic
	consumer.RegisterHandler(TopicBillingSubscriptionCreated, bc.handleSubscriptionCreated)
	consumer.RegisterHandler(TopicBillingSubscriptionUpgraded, bc.handleSubscriptionUpgraded)
	consumer.RegisterHandler(TopicBillingSubscriptionCancelled, bc.handleSubscriptionCancelled)
	consumer.RegisterHandler(TopicBillingSubscriptionExpired, bc.handleSubscriptionExpired)
	consumer.RegisterHandler(TopicBillingPaymentProcessed, bc.handlePaymentProcessed)
	consumer.RegisterHandler(TopicBillingPaymentFailed, bc.handlePaymentFailed)
	consumer.RegisterHandler(TopicBillingPaymentRefunded, bc.handlePaymentRefunded)

	return bc
}

// setResourceOwnerContext sets the resource owner in context using the standard context pattern
func setResourceOwnerContext(ctx context.Context, userID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	return ctx
}

func (bc *BillingConsumer) handleSubscriptionCreated(ctx context.Context, msg *kafka.Message) error {
	var event SubscriptionEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal subscription created event", "error", err)
		return fmt.Errorf("failed to unmarshal subscription created event: %w", err)
	}

	slog.InfoContext(ctx, "Processing subscription created event",
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
		"plan_id", event.PlanID,
	)

	// Log audit trail for subscription creation
	slog.InfoContext(ctx, "Subscription created successfully",
		"event_id", event.EventID,
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
		"plan_id", event.PlanID,
		"is_free", event.IsFree,
		"billing_period", event.BillingPeriod,
	)

	return nil
}

func (bc *BillingConsumer) handleSubscriptionUpgraded(ctx context.Context, msg *kafka.Message) error {
	var event SubscriptionEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal subscription upgraded event", "error", err)
		return fmt.Errorf("failed to unmarshal subscription upgraded event: %w", err)
	}

	slog.InfoContext(ctx, "Processing subscription upgraded event",
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
		"previous_plan_id", event.PreviousPlanID,
		"new_plan_id", event.PlanID,
	)

	// Log audit trail for subscription upgrade
	slog.InfoContext(ctx, "Subscription upgraded successfully",
		"event_id", event.EventID,
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
		"previous_plan_id", event.PreviousPlanID,
		"new_plan_id", event.PlanID,
		"reason", event.Reason,
	)

	return nil
}

func (bc *BillingConsumer) handleSubscriptionCancelled(ctx context.Context, msg *kafka.Message) error {
	var event SubscriptionEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal subscription cancelled event", "error", err)
		return fmt.Errorf("failed to unmarshal subscription cancelled event: %w", err)
	}

	slog.InfoContext(ctx, "Processing subscription cancelled event",
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
		"reason", event.Reason,
	)

	// Update subscription status if writer is available
	if bc.subscriptionWriter != nil && bc.subscriptionReader != nil {
		// Create a context with resource owner
		resourceOwner := shared.ResourceOwner{
			UserID: event.UserID,
		}
		ctx = setResourceOwnerContext(ctx, event.UserID)

		subscription, err := bc.subscriptionReader.GetCurrentSubscription(ctx, resourceOwner)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get current subscription for cancellation", "error", err)
			return err
		}

		if subscription != nil && subscription.ID == event.SubscriptionID {
			subscription.Status = billing_entities.SubscriptionStatusCanceled
			_, err = bc.subscriptionWriter.Update(ctx, subscription)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to update subscription status to cancelled", "error", err)
				return err
			}
		}
	}

	slog.InfoContext(ctx, "Subscription cancelled successfully",
		"event_id", event.EventID,
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
		"reason", event.Reason,
	)

	return nil
}

func (bc *BillingConsumer) handleSubscriptionExpired(ctx context.Context, msg *kafka.Message) error {
	var event SubscriptionEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal subscription expired event", "error", err)
		return fmt.Errorf("failed to unmarshal subscription expired event: %w", err)
	}

	slog.InfoContext(ctx, "Processing subscription expired event",
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
	)

	// Update subscription status if writer is available
	if bc.subscriptionWriter != nil && bc.subscriptionReader != nil {
		resourceOwner := shared.ResourceOwner{
			UserID: event.UserID,
		}
		ctx = setResourceOwnerContext(ctx, event.UserID)

		subscription, err := bc.subscriptionReader.GetCurrentSubscription(ctx, resourceOwner)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get current subscription for expiration", "error", err)
			return err
		}

		if subscription != nil && subscription.ID == event.SubscriptionID {
			subscription.Status = billing_entities.SubscriptionStatusExpired
			_, err = bc.subscriptionWriter.Update(ctx, subscription)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to update subscription status to expired", "error", err)
				return err
			}
		}

		// Create free subscription for user after expiration
		freePlan, err := bc.planReader.GetDefaultFreePlan(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get default free plan", "error", err)
			// Don't return error - subscription is already expired
		} else if freePlan != nil {
			// Create a new free subscription
			newSubscription := &billing_entities.Subscription{
				BaseEntity:    shared.NewRestrictedEntity(resourceOwner),
				PlanID:        freePlan.ID,
				BillingPeriod: billing_entities.BillingPeriodLifetime,
				Status:        billing_entities.SubscriptionStatusActive,
				IsFree:        true,
				History: []billing_entities.SubscriptionHistory{
					{
						Date:   subscription.UpdatedAt,
						Status: billing_entities.SubscriptionStatusActive,
						Reason: "Auto-created after subscription expiration",
					},
				},
			}
			newSubscription.StartAt = subscription.UpdatedAt

			_, err = bc.subscriptionWriter.Create(ctx, newSubscription)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to create free subscription after expiration", "error", err)
			} else {
				slog.InfoContext(ctx, "Created free subscription after expiration",
					"user_id", event.UserID,
					"new_subscription_id", newSubscription.ID,
				)
			}
		}
	}

	slog.InfoContext(ctx, "Subscription expired successfully",
		"event_id", event.EventID,
		"subscription_id", event.SubscriptionID,
		"user_id", event.UserID,
	)

	return nil
}

func (bc *BillingConsumer) handlePaymentProcessed(ctx context.Context, msg *kafka.Message) error {
	var event PaymentEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal payment processed event", "error", err)
		return fmt.Errorf("failed to unmarshal payment processed event: %w", err)
	}

	slog.InfoContext(ctx, "Processing payment processed event",
		"payment_id", event.PaymentID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"provider", event.Provider,
	)

	// Log audit trail for payment
	slog.InfoContext(ctx, "Payment processed successfully",
		"event_id", event.EventID,
		"payment_id", event.PaymentID,
		"user_id", event.UserID,
		"subscription_id", event.SubscriptionID,
		"amount", event.Amount,
		"currency", event.Currency,
		"provider", event.Provider,
		"provider_ref", event.ProviderRef,
	)

	return nil
}

func (bc *BillingConsumer) handlePaymentFailed(ctx context.Context, msg *kafka.Message) error {
	var event PaymentEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal payment failed event", "error", err)
		return fmt.Errorf("failed to unmarshal payment failed event: %w", err)
	}

	slog.WarnContext(ctx, "Processing payment failed event",
		"payment_id", event.PaymentID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"failure_reason", event.FailureReason,
	)

	// Update subscription status if applicable
	if bc.subscriptionWriter != nil && bc.subscriptionReader != nil && event.SubscriptionID != nil {
		resourceOwner := shared.ResourceOwner{
			UserID: event.UserID,
		}
		ctx = setResourceOwnerContext(ctx, event.UserID)

		subscription, err := bc.subscriptionReader.GetCurrentSubscription(ctx, resourceOwner)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get subscription for payment failure", "error", err)
		} else if subscription != nil && subscription.ID == *event.SubscriptionID {
			// Add history entry for payment rejection
			subscription.History = append(subscription.History, billing_entities.SubscriptionHistory{
				Date:   subscription.UpdatedAt,
				Status: billing_entities.SubscriptionStatusPaymentRejected,
				Reason: event.FailureReason,
			})

			_, err = bc.subscriptionWriter.Update(ctx, subscription)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to update subscription with payment failure", "error", err)
			}
		}
	}

	slog.WarnContext(ctx, "Payment failed recorded",
		"event_id", event.EventID,
		"payment_id", event.PaymentID,
		"user_id", event.UserID,
		"failure_reason", event.FailureReason,
	)

	return nil
}

func (bc *BillingConsumer) handlePaymentRefunded(ctx context.Context, msg *kafka.Message) error {
	var event PaymentEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal payment refunded event", "error", err)
		return fmt.Errorf("failed to unmarshal payment refunded event: %w", err)
	}

	slog.InfoContext(ctx, "Processing payment refunded event",
		"payment_id", event.PaymentID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
	)

	// Log audit trail for refund
	slog.InfoContext(ctx, "Payment refunded successfully",
		"event_id", event.EventID,
		"payment_id", event.PaymentID,
		"user_id", event.UserID,
		"subscription_id", event.SubscriptionID,
		"amount", event.Amount,
		"currency", event.Currency,
	)

	return nil
}

// Start begins consuming billing events
func (bc *BillingConsumer) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting billing consumer")
	return bc.consumer.Start(ctx)
}

// Close closes the billing consumer
func (bc *BillingConsumer) Close() error {
	return bc.consumer.Close()
}

// GetAllBillingTopics returns all billing-related topics
func GetAllBillingTopics() []string {
	return []string{
		TopicBillingSubscriptionCreated,
		TopicBillingSubscriptionUpgraded,
		TopicBillingSubscriptionCancelled,
		TopicBillingSubscriptionExpired,
		TopicBillingPaymentProcessed,
		TopicBillingPaymentFailed,
		TopicBillingPaymentRefunded,
	}
}
