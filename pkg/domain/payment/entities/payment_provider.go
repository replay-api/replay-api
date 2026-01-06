package entities

import (
	"time"

	"github.com/google/uuid"
)

// PaymentProvider represents the payment gateway used
type PaymentProvider string

const (
	PaymentProviderStripe  PaymentProvider = "stripe"
	PaymentProviderPayPal  PaymentProvider = "paypal"
	PaymentProviderCrypto  PaymentProvider = "crypto"
	PaymentProviderBank    PaymentProvider = "bank"
)

// PaymentStatus represents the current state of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCanceled  PaymentStatus = "canceled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentType represents the type of payment operation
type PaymentType string

const (
	PaymentTypeDeposit    PaymentType = "deposit"
	PaymentTypeWithdrawal PaymentType = "withdrawal"
	PaymentTypeSubscription PaymentType = "subscription"
)

// Payment represents a payment transaction
type Payment struct {
	ID                  uuid.UUID       `json:"id" bson:"_id"`
	UserID              uuid.UUID       `json:"user_id" bson:"user_id"`
	WalletID            uuid.UUID       `json:"wallet_id" bson:"wallet_id"`
	Type                PaymentType     `json:"type" bson:"type"`
	Provider            PaymentProvider `json:"provider" bson:"provider"`
	Status              PaymentStatus   `json:"status" bson:"status"`
	Amount              int64           `json:"amount" bson:"amount"` // in cents
	Currency            string          `json:"currency" bson:"currency"`
	Fee                 int64           `json:"fee" bson:"fee"` // platform fee in cents
	ProviderFee         int64           `json:"provider_fee" bson:"provider_fee"` // gateway fee
	NetAmount           int64           `json:"net_amount" bson:"net_amount"`

	// Provider-specific fields
	ProviderPaymentID   string          `json:"provider_payment_id" bson:"provider_payment_id"`
	ProviderCustomerID  string          `json:"provider_customer_id,omitempty" bson:"provider_customer_id,omitempty"`
	PaymentMethodID     string          `json:"payment_method_id,omitempty" bson:"payment_method_id,omitempty"`

	// Metadata
	Description         string          `json:"description,omitempty" bson:"description,omitempty"`
	Metadata            map[string]any  `json:"metadata,omitempty" bson:"metadata,omitempty"`
	FailureReason       string          `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"`

	// Timestamps
	CreatedAt           time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" bson:"updated_at"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty" bson:"completed_at,omitempty"`

	// Idempotency
	IdempotencyKey      string          `json:"idempotency_key" bson:"idempotency_key"`
}

// GetID returns the payment ID (implements shared.Entity interface)
func (p *Payment) GetID() uuid.UUID {
	return p.ID
}

// NewPayment creates a new payment
func NewPayment(userID, walletID uuid.UUID, paymentType PaymentType, provider PaymentProvider, amount int64, currency string) *Payment {
	now := time.Now().UTC()
	return &Payment{
		ID:             uuid.New(),
		UserID:         userID,
		WalletID:       walletID,
		Type:           paymentType,
		Provider:       provider,
		Status:         PaymentStatusPending,
		Amount:         amount,
		Currency:       currency,
		Fee:            0,
		ProviderFee:    0,
		NetAmount:      amount,
		CreatedAt:      now,
		UpdatedAt:      now,
		IdempotencyKey: uuid.New().String(),
		Metadata:       make(map[string]any),
	}
}

// SetProviderDetails sets provider-specific details
func (p *Payment) SetProviderDetails(providerPaymentID, providerCustomerID, paymentMethodID string) {
	p.ProviderPaymentID = providerPaymentID
	p.ProviderCustomerID = providerCustomerID
	p.PaymentMethodID = paymentMethodID
	p.UpdatedAt = time.Now().UTC()
}

// MarkProcessing marks the payment as processing
func (p *Payment) MarkProcessing() {
	p.Status = PaymentStatusProcessing
	p.UpdatedAt = time.Now().UTC()
}

// MarkSucceeded marks the payment as succeeded
func (p *Payment) MarkSucceeded(providerFee int64) {
	now := time.Now().UTC()
	p.Status = PaymentStatusSucceeded
	p.ProviderFee = providerFee
	p.NetAmount = p.Amount - p.Fee - providerFee
	p.CompletedAt = &now
	p.UpdatedAt = now
}

// MarkFailed marks the payment as failed
func (p *Payment) MarkFailed(reason string) {
	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	p.UpdatedAt = time.Now().UTC()
}

// MarkCanceled marks the payment as canceled
func (p *Payment) MarkCanceled() {
	p.Status = PaymentStatusCanceled
	p.UpdatedAt = time.Now().UTC()
}

// MarkRefunded marks the payment as refunded
func (p *Payment) MarkRefunded() {
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = time.Now().UTC()
}

// SetFee sets the platform fee
func (p *Payment) SetFee(fee int64) {
	p.Fee = fee
	p.NetAmount = p.Amount - fee - p.ProviderFee
	p.UpdatedAt = time.Now().UTC()
}

// IsPending returns true if the payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsCompleted returns true if the payment is in a final state
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusSucceeded ||
		p.Status == PaymentStatusFailed ||
		p.Status == PaymentStatusCanceled ||
		p.Status == PaymentStatusRefunded
}
