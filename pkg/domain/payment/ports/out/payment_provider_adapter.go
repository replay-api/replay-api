// Package payment_out defines outbound port interfaces for payment provider adapters
package payment_out

import (
	"context"

	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
)

// CreateIntentRequest represents a request to create a payment intent with a provider
type CreateIntentRequest struct {
	Amount            int64          `json:"amount"` // in cents
	Currency          string         `json:"currency"`
	CustomerID        string         `json:"customer_id,omitempty"`
	PaymentMethodID   string         `json:"payment_method_id,omitempty"`
	Description       string         `json:"description,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	IdempotencyKey    string         `json:"idempotency_key"`
	ReturnURL         string         `json:"return_url,omitempty"`
	CancelURL         string         `json:"cancel_url,omitempty"`
}

// CreateIntentResponse represents the response from creating a payment intent
type CreateIntentResponse struct {
	ProviderPaymentID string `json:"provider_payment_id"`
	ClientSecret      string `json:"client_secret,omitempty"`  // Stripe
	RedirectURL       string `json:"redirect_url,omitempty"`   // PayPal
	CryptoAddress     string `json:"crypto_address,omitempty"` // Crypto
	Status            string `json:"status"`
}

// ConfirmPaymentRequest represents a request to confirm a payment
type ConfirmPaymentRequest struct {
	ProviderPaymentID string `json:"provider_payment_id"`
	PaymentMethodID   string `json:"payment_method_id,omitempty"`
}

// ConfirmPaymentResponse represents the response from confirming a payment
type ConfirmPaymentResponse struct {
	Status      string `json:"status"`
	ProviderFee int64  `json:"provider_fee"` // in cents
}

// RefundRequest represents a request to refund a payment
type RefundRequest struct {
	ProviderPaymentID string `json:"provider_payment_id"`
	Amount            int64  `json:"amount,omitempty"` // 0 = full refund
	Reason            string `json:"reason,omitempty"`
	IdempotencyKey    string `json:"idempotency_key"`
}

// RefundResponse represents the response from a refund request
type RefundResponse struct {
	RefundID    string `json:"refund_id"`
	Status      string `json:"status"`
	Amount      int64  `json:"amount"`
}

// CancelRequest represents a request to cancel a payment
type CancelRequest struct {
	ProviderPaymentID string `json:"provider_payment_id"`
	Reason            string `json:"reason,omitempty"`
}

// CancelResponse represents the response from canceling a payment
type CancelResponse struct {
	Status string `json:"status"`
}

// WebhookEvent represents a parsed webhook event from a provider
type WebhookEvent struct {
	EventType         string                       `json:"event_type"`
	ProviderPaymentID string                       `json:"provider_payment_id"`
	Status            payment_entities.PaymentStatus `json:"status"`
	ProviderFee       int64                        `json:"provider_fee,omitempty"`
	FailureReason     string                       `json:"failure_reason,omitempty"`
	Metadata          map[string]any               `json:"metadata,omitempty"`
}

// CustomerRequest represents a request to create/get a customer
type CustomerRequest struct {
	Email    string         `json:"email"`
	Name     string         `json:"name,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CustomerResponse represents the response from customer operations
type CustomerResponse struct {
	CustomerID string `json:"customer_id"`
}

// PaymentProviderAdapter defines the interface for payment provider integrations
type PaymentProviderAdapter interface {
	// GetProvider returns the provider type this adapter handles
	GetProvider() payment_entities.PaymentProvider

	// CreatePaymentIntent creates a payment intent with the provider
	CreatePaymentIntent(ctx context.Context, req CreateIntentRequest) (*CreateIntentResponse, error)

	// ConfirmPayment confirms a payment with the provider
	ConfirmPayment(ctx context.Context, req ConfirmPaymentRequest) (*ConfirmPaymentResponse, error)

	// RefundPayment refunds a payment with the provider
	RefundPayment(ctx context.Context, req RefundRequest) (*RefundResponse, error)

	// CancelPayment cancels a payment with the provider
	CancelPayment(ctx context.Context, req CancelRequest) (*CancelResponse, error)

	// ParseWebhook parses and validates a webhook from the provider
	ParseWebhook(payload []byte, signature string) (*WebhookEvent, error)

	// CreateOrGetCustomer creates or retrieves a customer record with the provider
	CreateOrGetCustomer(ctx context.Context, req CustomerRequest) (*CustomerResponse, error)
}
