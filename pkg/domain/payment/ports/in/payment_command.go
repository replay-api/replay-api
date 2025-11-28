// Package payment_in defines inbound port interfaces for payment operations
package payment_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
)

// CreatePaymentIntentCommand represents a request to create a payment intent
type CreatePaymentIntentCommand struct {
	UserID      uuid.UUID                        `json:"user_id"`
	WalletID    uuid.UUID                        `json:"wallet_id"`
	Amount      int64                            `json:"amount"` // in cents
	Currency    string                           `json:"currency"`
	PaymentType payment_entities.PaymentType     `json:"payment_type"`
	Provider    payment_entities.PaymentProvider `json:"provider"`
	Metadata    map[string]any                   `json:"metadata,omitempty"`
}

// Validate validates the command parameters
func (c *CreatePaymentIntentCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if c.WalletID == uuid.Nil {
		return errors.New("wallet_id is required")
	}
	if c.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if c.Currency == "" {
		return errors.New("currency is required")
	}
	if c.PaymentType == "" {
		return errors.New("payment_type is required")
	}
	if c.Provider == "" {
		return errors.New("provider is required")
	}
	return nil
}

// ConfirmPaymentCommand represents a request to confirm a payment
type ConfirmPaymentCommand struct {
	PaymentID       uuid.UUID `json:"payment_id"`
	UserID          uuid.UUID `json:"user_id"` // For authorization
	PaymentMethodID string    `json:"payment_method_id"`
}

// Validate validates the command parameters
func (c *ConfirmPaymentCommand) Validate() error {
	if c.PaymentID == uuid.Nil {
		return errors.New("payment_id is required")
	}
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if c.PaymentMethodID == "" {
		return errors.New("payment_method_id is required")
	}
	return nil
}

// RefundPaymentCommand represents a request to refund a payment
type RefundPaymentCommand struct {
	PaymentID uuid.UUID `json:"payment_id"`
	UserID    uuid.UUID `json:"user_id"` // For authorization
	Amount    int64     `json:"amount,omitempty"` // partial refund, 0 = full refund
	Reason    string    `json:"reason"`
}

// Validate validates the command parameters
func (c *RefundPaymentCommand) Validate() error {
	if c.PaymentID == uuid.Nil {
		return errors.New("payment_id is required")
	}
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if c.Amount < 0 {
		return errors.New("amount cannot be negative")
	}
	if c.Reason == "" {
		return errors.New("reason is required for refunds")
	}
	return nil
}

// CancelPaymentCommand represents a request to cancel a payment
type CancelPaymentCommand struct {
	PaymentID uuid.UUID `json:"payment_id"`
	UserID    uuid.UUID `json:"user_id"` // For authorization
	Reason    string    `json:"reason,omitempty"`
}

// Validate validates the command parameters
func (c *CancelPaymentCommand) Validate() error {
	if c.PaymentID == uuid.Nil {
		return errors.New("payment_id is required")
	}
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	return nil
}

// ProcessWebhookCommand represents an incoming webhook from a payment provider
type ProcessWebhookCommand struct {
	Provider  payment_entities.PaymentProvider `json:"provider"`
	EventType string                           `json:"event_type"`
	Payload   []byte                           `json:"payload"`
	Signature string                           `json:"signature"`
}

// Validate validates the command parameters
func (c *ProcessWebhookCommand) Validate() error {
	if c.Provider == "" {
		return errors.New("provider is required")
	}
	if c.EventType == "" {
		return errors.New("event_type is required")
	}
	if len(c.Payload) == 0 {
		return errors.New("payload is required")
	}
	if c.Signature == "" {
		return errors.New("signature is required for webhook verification")
	}
	return nil
}

// PaymentIntentResult represents the result of creating a payment intent
type PaymentIntentResult struct {
	Payment      *payment_entities.Payment `json:"payment"`
	ClientSecret string                    `json:"client_secret,omitempty"` // For Stripe
	RedirectURL  string                    `json:"redirect_url,omitempty"`  // For PayPal
	CryptoAddress string                   `json:"crypto_address,omitempty"` // For Crypto
}

// PaymentCommand defines the inbound port for payment operations
type PaymentCommand interface {
	// CreatePaymentIntent creates a new payment intent with the specified provider
	CreatePaymentIntent(ctx context.Context, cmd CreatePaymentIntentCommand) (*PaymentIntentResult, error)

	// ConfirmPayment confirms a payment with the payment method
	ConfirmPayment(ctx context.Context, cmd ConfirmPaymentCommand) (*payment_entities.Payment, error)

	// RefundPayment refunds a payment (full or partial)
	RefundPayment(ctx context.Context, cmd RefundPaymentCommand) (*payment_entities.Payment, error)

	// CancelPayment cancels a pending payment
	CancelPayment(ctx context.Context, cmd CancelPaymentCommand) (*payment_entities.Payment, error)

	// ProcessWebhook processes a webhook event from a payment provider
	ProcessWebhook(ctx context.Context, cmd ProcessWebhookCommand) error
}
