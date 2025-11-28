// Package payment_out defines outbound repository interfaces for payments
package payment_out

import (
	"context"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
)

// PaymentFilters defines filters for querying payments
type PaymentFilters struct {
	UserID         *uuid.UUID                       `json:"user_id,omitempty"`
	WalletID       *uuid.UUID                       `json:"wallet_id,omitempty"`
	Provider       *payment_entities.PaymentProvider `json:"provider,omitempty"`
	Status         *payment_entities.PaymentStatus   `json:"status,omitempty"`
	Type           *payment_entities.PaymentType     `json:"type,omitempty"`
	IdempotencyKey *string                          `json:"idempotency_key,omitempty"`
	Limit          int                              `json:"limit"`
	Offset         int                              `json:"offset"`
}

// PaymentRepository defines persistence operations for payments
type PaymentRepository interface {
	// Save creates a new payment record
	Save(ctx context.Context, payment *payment_entities.Payment) error

	// FindByID retrieves a payment by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*payment_entities.Payment, error)

	// FindByProviderPaymentID retrieves a payment by the provider's payment ID
	FindByProviderPaymentID(ctx context.Context, providerPaymentID string) (*payment_entities.Payment, error)

	// FindByIdempotencyKey retrieves a payment by idempotency key
	FindByIdempotencyKey(ctx context.Context, key string) (*payment_entities.Payment, error)

	// FindByUserID retrieves all payments for a user
	FindByUserID(ctx context.Context, userID uuid.UUID, filters PaymentFilters) ([]*payment_entities.Payment, error)

	// FindByWalletID retrieves all payments for a wallet
	FindByWalletID(ctx context.Context, walletID uuid.UUID, filters PaymentFilters) ([]*payment_entities.Payment, error)

	// Update updates an existing payment record
	Update(ctx context.Context, payment *payment_entities.Payment) error

	// GetPendingPayments retrieves all pending payments older than specified duration
	GetPendingPayments(ctx context.Context, olderThanSeconds int) ([]*payment_entities.Payment, error)
}
