// Package stripe provides the Stripe payment provider adapter implementation
package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"

	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// StripeAdapter implements the PaymentProviderAdapter interface for Stripe
type StripeAdapter struct {
	webhookSecret string
}

// NewStripeAdapter creates a new Stripe adapter
func NewStripeAdapter() *StripeAdapter {
	// Set Stripe API key from environment
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeAdapter{
		webhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}
}

// GetProvider returns the provider type
func (s *StripeAdapter) GetProvider() payment_entities.PaymentProvider {
	return payment_entities.PaymentProviderStripe
}

// CreatePaymentIntent creates a Stripe PaymentIntent
func (s *StripeAdapter) CreatePaymentIntent(ctx context.Context, req payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
	}

	// Set customer if provided
	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}

	// Set payment method if provided
	if req.PaymentMethodID != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
	}

	// Set description
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	// Set metadata
	if len(req.Metadata) > 0 {
		params.Metadata = make(map[string]string)
		for k, v := range req.Metadata {
			params.Metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	// Set idempotency key
	params.SetIdempotencyKey(req.IdempotencyKey)

	// Automatic payment methods
	params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
		Enabled: stripe.Bool(true),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe PaymentIntent: %w", err)
	}

	return &payment_out.CreateIntentResponse{
		ProviderPaymentID: pi.ID,
		ClientSecret:      pi.ClientSecret,
		Status:            string(pi.Status),
	}, nil
}

// ConfirmPayment confirms a Stripe PaymentIntent
func (s *StripeAdapter) ConfirmPayment(ctx context.Context, req payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error) {
	params := &stripe.PaymentIntentConfirmParams{}

	if req.PaymentMethodID != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
	}

	pi, err := paymentintent.Confirm(req.ProviderPaymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm Stripe PaymentIntent: %w", err)
	}

	// Calculate provider fee (Stripe charges ~2.9% + $0.30)
	// This is an estimate; actual fees come from balance transactions
	providerFee := int64(float64(pi.Amount)*0.029) + 30

	return &payment_out.ConfirmPaymentResponse{
		Status:      string(pi.Status),
		ProviderFee: providerFee,
	}, nil
}

// RefundPayment refunds a Stripe payment
func (s *StripeAdapter) RefundPayment(ctx context.Context, req payment_out.RefundRequest) (*payment_out.RefundResponse, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.ProviderPaymentID),
	}

	// Partial refund if amount specified
	if req.Amount > 0 {
		params.Amount = stripe.Int64(req.Amount)
	}

	// Set reason if provided
	if req.Reason != "" {
		switch req.Reason {
		case "duplicate":
			params.Reason = stripe.String(string(stripe.RefundReasonDuplicate))
		case "fraudulent":
			params.Reason = stripe.String(string(stripe.RefundReasonFraudulent))
		default:
			params.Reason = stripe.String(string(stripe.RefundReasonRequestedByCustomer))
		}
	}

	// Set idempotency key
	params.SetIdempotencyKey(req.IdempotencyKey)

	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe refund: %w", err)
	}

	return &payment_out.RefundResponse{
		RefundID: r.ID,
		Status:   string(r.Status),
		Amount:   r.Amount,
	}, nil
}

// CancelPayment cancels a Stripe PaymentIntent
func (s *StripeAdapter) CancelPayment(ctx context.Context, req payment_out.CancelRequest) (*payment_out.CancelResponse, error) {
	params := &stripe.PaymentIntentCancelParams{}

	if req.Reason != "" {
		switch req.Reason {
		case "duplicate":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonDuplicate))
		case "fraudulent":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonFraudulent))
		case "abandoned":
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonAbandoned))
		default:
			params.CancellationReason = stripe.String(string(stripe.PaymentIntentCancellationReasonRequestedByCustomer))
		}
	}

	pi, err := paymentintent.Cancel(req.ProviderPaymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel Stripe PaymentIntent: %w", err)
	}

	return &payment_out.CancelResponse{
		Status: string(pi.Status),
	}, nil
}

// ParseWebhook parses and validates a Stripe webhook
func (s *StripeAdapter) ParseWebhook(payload []byte, signature string) (*payment_out.WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Stripe webhook signature: %w", err)
	}

	webhookEvent := &payment_out.WebhookEvent{
		EventType: string(event.Type),
		Metadata:  make(map[string]any),
	}

	// Parse event based on type
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := parseStripeObject(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse payment_intent: %w", err)
		}
		webhookEvent.ProviderPaymentID = pi.ID
		webhookEvent.Status = payment_entities.PaymentStatusSucceeded

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := parseStripeObject(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse payment_intent: %w", err)
		}
		webhookEvent.ProviderPaymentID = pi.ID
		webhookEvent.Status = payment_entities.PaymentStatusFailed
		if pi.LastPaymentError != nil {
			webhookEvent.FailureReason = pi.LastPaymentError.Msg
		}

	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := parseStripeObject(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse payment_intent: %w", err)
		}
		webhookEvent.ProviderPaymentID = pi.ID
		webhookEvent.Status = payment_entities.PaymentStatusCanceled

	case "payment_intent.processing":
		var pi stripe.PaymentIntent
		if err := parseStripeObject(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse payment_intent: %w", err)
		}
		webhookEvent.ProviderPaymentID = pi.ID
		webhookEvent.Status = payment_entities.PaymentStatusProcessing

	case "charge.refunded":
		var ch stripe.Charge
		if err := parseStripeObject(event.Data.Raw, &ch); err != nil {
			return nil, fmt.Errorf("failed to parse charge: %w", err)
		}
		if ch.PaymentIntent != nil {
			webhookEvent.ProviderPaymentID = ch.PaymentIntent.ID
		}
		webhookEvent.Status = payment_entities.PaymentStatusRefunded

	default:
		return nil, fmt.Errorf("unhandled Stripe event type: %s", event.Type)
	}

	return webhookEvent, nil
}

// CreateOrGetCustomer creates or retrieves a Stripe customer
func (s *StripeAdapter) CreateOrGetCustomer(ctx context.Context, req payment_out.CustomerRequest) (*payment_out.CustomerResponse, error) {
	// Search for existing customer by email
	searchParams := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("email:'%s'", req.Email),
		},
	}

	iter := customer.Search(searchParams)
	for iter.Next() {
		c := iter.Customer()
		return &payment_out.CustomerResponse{
			CustomerID: c.ID,
		}, nil
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to search Stripe customers: %w", err)
	}

	// Create new customer if not found
	params := &stripe.CustomerParams{
		Email: stripe.String(req.Email),
	}

	if req.Name != "" {
		params.Name = stripe.String(req.Name)
	}

	if len(req.Metadata) > 0 {
		params.Metadata = make(map[string]string)
		for k, v := range req.Metadata {
			params.Metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	return &payment_out.CustomerResponse{
		CustomerID: c.ID,
	}, nil
}

// parseStripeObject is a helper to parse Stripe webhook data
func parseStripeObject(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Ensure StripeAdapter implements PaymentProviderAdapter
var _ payment_out.PaymentProviderAdapter = (*StripeAdapter)(nil)
