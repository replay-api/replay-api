// Package payment_in defines inbound query interfaces for payment operations
package payment_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
)

// GetPaymentQuery request to get a single payment by ID
type GetPaymentQuery struct {
	PaymentID uuid.UUID
	UserID    uuid.UUID // For authorization
}

// Validate validates the query parameters
func (q *GetPaymentQuery) Validate() error {
	if q.PaymentID == uuid.Nil {
		return &ValidationError{Field: "payment_id", Message: "payment_id is required"}
	}
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// GetUserPaymentsQuery request to get payments for a user
type GetUserPaymentsQuery struct {
	UserID  uuid.UUID
	Filters PaymentQueryFilters
}

// Validate validates the query parameters
func (q *GetUserPaymentsQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	if q.Filters.Limit <= 0 {
		q.Filters.Limit = 20
	}
	if q.Filters.Limit > 100 {
		q.Filters.Limit = 100
	}
	return nil
}

// PaymentQueryFilters defines filters for payment queries
type PaymentQueryFilters struct {
	Provider  *payment_entities.PaymentProvider `json:"provider,omitempty"`
	Status    *payment_entities.PaymentStatus   `json:"status,omitempty"`
	Type      *payment_entities.PaymentType     `json:"type,omitempty"`
	Currency  *string                           `json:"currency,omitempty"`
	FromDate  *time.Time                        `json:"from_date,omitempty"`
	ToDate    *time.Time                        `json:"to_date,omitempty"`
	Limit     int                               `json:"limit"`
	Offset    int                               `json:"offset"`
	SortBy    string                            `json:"sort_by"`    // "created_at", "amount", "updated_at"
	SortOrder string                            `json:"sort_order"` // "asc", "desc"
}

// PaymentDTO represents a payment in API responses
type PaymentDTO struct {
	ID                 uuid.UUID                       `json:"id"`
	UserID             uuid.UUID                       `json:"user_id"`
	WalletID           uuid.UUID                       `json:"wallet_id"`
	Type               payment_entities.PaymentType    `json:"type"`
	Provider           payment_entities.PaymentProvider `json:"provider"`
	Status             payment_entities.PaymentStatus  `json:"status"`
	Amount             int64                           `json:"amount"`
	Currency           string                          `json:"currency"`
	Fee                int64                           `json:"fee"`
	ProviderFee        int64                           `json:"provider_fee"`
	NetAmount          int64                           `json:"net_amount"`
	Description        string                          `json:"description,omitempty"`
	FailureReason      string                          `json:"failure_reason,omitempty"`
	ProviderPaymentID  string                          `json:"provider_payment_id,omitempty"`
	CreatedAt          time.Time                       `json:"created_at"`
	UpdatedAt          time.Time                       `json:"updated_at"`
	CompletedAt        *time.Time                      `json:"completed_at,omitempty"`
}

// PaymentsResult represents the result of a payments list query
type PaymentsResult struct {
	Payments   []PaymentDTO `json:"payments"`
	TotalCount int64        `json:"total_count"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
}

// PaymentQuery defines query operations for payments
type PaymentQuery interface {
	GetPayment(ctx context.Context, query GetPaymentQuery) (*PaymentDTO, error)
	GetUserPayments(ctx context.Context, query GetUserPaymentsQuery) (*PaymentsResult, error)
}

// PaymentToDTO converts a Payment entity to a PaymentDTO
func PaymentToDTO(p *payment_entities.Payment) PaymentDTO {
	return PaymentDTO{
		ID:                p.ID,
		UserID:            p.UserID,
		WalletID:          p.WalletID,
		Type:              p.Type,
		Provider:          p.Provider,
		Status:            p.Status,
		Amount:            p.Amount,
		Currency:          p.Currency,
		Fee:               p.Fee,
		ProviderFee:       p.ProviderFee,
		NetAmount:         p.NetAmount,
		Description:       p.Description,
		FailureReason:     p.FailureReason,
		ProviderPaymentID: p.ProviderPaymentID,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
		CompletedAt:       p.CompletedAt,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
