package billing_entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// WithdrawalStatus represents the processing state of a withdrawal
type WithdrawalStatus string

const (
	WithdrawalStatusPending    WithdrawalStatus = "pending"
	WithdrawalStatusReviewing  WithdrawalStatus = "reviewing"
	WithdrawalStatusApproved   WithdrawalStatus = "approved"
	WithdrawalStatusProcessing WithdrawalStatus = "processing"
	WithdrawalStatusCompleted  WithdrawalStatus = "completed"
	WithdrawalStatusRejected   WithdrawalStatus = "rejected"
	WithdrawalStatusFailed     WithdrawalStatus = "failed"
	WithdrawalStatusCanceled   WithdrawalStatus = "canceled"
)

// WithdrawalMethod represents the method of withdrawal
type WithdrawalMethod string

const (
	WithdrawalMethodBankTransfer WithdrawalMethod = "bank_transfer"
	WithdrawalMethodPIX          WithdrawalMethod = "pix"
	WithdrawalMethodPayPal       WithdrawalMethod = "paypal"
	WithdrawalMethodCrypto       WithdrawalMethod = "crypto"
	WithdrawalMethodWire         WithdrawalMethod = "wire"
)

// BankDetails contains bank account information for withdrawals
type BankDetails struct {
	AccountHolder   string `json:"account_holder" bson:"account_holder"`
	BankName        string `json:"bank_name" bson:"bank_name"`
	BankCode        string `json:"bank_code,omitempty" bson:"bank_code,omitempty"`
	AccountNumber   string `json:"account_number" bson:"account_number"`
	RoutingNumber   string `json:"routing_number,omitempty" bson:"routing_number,omitempty"`
	IBAN            string `json:"iban,omitempty" bson:"iban,omitempty"`
	SWIFT           string `json:"swift,omitempty" bson:"swift,omitempty"`
	PIXKey          string `json:"pix_key,omitempty" bson:"pix_key,omitempty"`
	PIXKeyType      string `json:"pix_key_type,omitempty" bson:"pix_key_type,omitempty"`
	PayPalEmail     string `json:"paypal_email,omitempty" bson:"paypal_email,omitempty"`
	CryptoAddress   string `json:"crypto_address,omitempty" bson:"crypto_address,omitempty"`
	CryptoNetwork   string `json:"crypto_network,omitempty" bson:"crypto_network,omitempty"`
}

// Withdrawal represents a withdrawal request from a user's wallet
type Withdrawal struct {
	ID                   uuid.UUID            `json:"id" bson:"_id"`
	UserID               uuid.UUID            `json:"user_id" bson:"user_id"`
	WalletID             uuid.UUID            `json:"wallet_id" bson:"wallet_id"`
	Amount               float64              `json:"amount" bson:"amount"`
	Currency             string               `json:"currency" bson:"currency"`
	Method               WithdrawalMethod     `json:"method" bson:"method"`
	Status               WithdrawalStatus     `json:"status" bson:"status"`
	BankDetails          BankDetails          `json:"bank_details" bson:"bank_details"`
	Fee                  float64              `json:"fee" bson:"fee"`
	NetAmount            float64              `json:"net_amount" bson:"net_amount"`
	ExternalReference    string               `json:"external_reference,omitempty" bson:"external_reference,omitempty"`
	ProviderReference    string               `json:"provider_reference,omitempty" bson:"provider_reference,omitempty"`
	RejectionReason      string               `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
	Notes                string               `json:"notes,omitempty" bson:"notes,omitempty"`
	ReviewedBy           *uuid.UUID           `json:"reviewed_by,omitempty" bson:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time           `json:"reviewed_at,omitempty" bson:"reviewed_at,omitempty"`
	ProcessedAt          *time.Time           `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	CompletedAt          *time.Time           `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	History              []WithdrawalHistory  `json:"history" bson:"history"`
	ResourceOwner        common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt            time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at" bson:"updated_at"`
}

// WithdrawalHistory tracks status changes for a withdrawal
type WithdrawalHistory struct {
	Status    WithdrawalStatus `json:"status" bson:"status"`
	Reason    string           `json:"reason,omitempty" bson:"reason,omitempty"`
	UpdatedBy *uuid.UUID       `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	Timestamp time.Time        `json:"timestamp" bson:"timestamp"`
}

// NewWithdrawal creates a new withdrawal request
func NewWithdrawal(
	userID uuid.UUID,
	walletID uuid.UUID,
	amount float64,
	currency string,
	method WithdrawalMethod,
	bankDetails BankDetails,
	fee float64,
	rxn common.ResourceOwner,
) *Withdrawal {
	now := time.Now()
	return &Withdrawal{
		ID:            uuid.New(),
		UserID:        userID,
		WalletID:      walletID,
		Amount:        amount,
		Currency:      currency,
		Method:        method,
		Status:        WithdrawalStatusPending,
		BankDetails:   bankDetails,
		Fee:           fee,
		NetAmount:     amount - fee,
		History: []WithdrawalHistory{
			{
				Status:    WithdrawalStatusPending,
				Reason:    "Withdrawal request created",
				Timestamp: now,
			},
		},
		ResourceOwner: rxn,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GetID returns the withdrawal ID
func (w Withdrawal) GetID() uuid.UUID {
	return w.ID
}

// Validate validates the withdrawal request
func (w *Withdrawal) Validate() error {
	if w.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if w.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if w.Currency == "" {
		return errors.New("currency is required")
	}
	if w.Method == "" {
		return errors.New("withdrawal method is required")
	}
	return w.validateBankDetails()
}

func (w *Withdrawal) validateBankDetails() error {
	switch w.Method {
	case WithdrawalMethodPIX:
		if w.BankDetails.PIXKey == "" {
			return errors.New("PIX key is required for PIX withdrawals")
		}
	case WithdrawalMethodPayPal:
		if w.BankDetails.PayPalEmail == "" {
			return errors.New("PayPal email is required for PayPal withdrawals")
		}
	case WithdrawalMethodCrypto:
		if w.BankDetails.CryptoAddress == "" {
			return errors.New("crypto address is required for crypto withdrawals")
		}
	case WithdrawalMethodBankTransfer, WithdrawalMethodWire:
		if w.BankDetails.AccountNumber == "" {
			return errors.New("account number is required for bank transfers")
		}
	}
	return nil
}

// Approve marks the withdrawal as approved for processing
func (w *Withdrawal) Approve(reviewerID uuid.UUID) {
	now := time.Now()
	w.Status = WithdrawalStatusApproved
	w.ReviewedBy = &reviewerID
	w.ReviewedAt = &now
	w.UpdatedAt = now
	w.addHistory(WithdrawalStatusApproved, "Withdrawal approved", &reviewerID)
}

// Reject marks the withdrawal as rejected
func (w *Withdrawal) Reject(reviewerID uuid.UUID, reason string) {
	now := time.Now()
	w.Status = WithdrawalStatusRejected
	w.ReviewedBy = &reviewerID
	w.ReviewedAt = &now
	w.RejectionReason = reason
	w.UpdatedAt = now
	w.addHistory(WithdrawalStatusRejected, reason, &reviewerID)
}

// MarkProcessing marks the withdrawal as being processed
func (w *Withdrawal) MarkProcessing() {
	now := time.Now()
	w.Status = WithdrawalStatusProcessing
	w.ProcessedAt = &now
	w.UpdatedAt = now
	w.addHistory(WithdrawalStatusProcessing, "Payment processing started", nil)
}

// Complete marks the withdrawal as completed
func (w *Withdrawal) Complete(providerRef string) {
	now := time.Now()
	w.Status = WithdrawalStatusCompleted
	w.ProviderReference = providerRef
	w.CompletedAt = &now
	w.UpdatedAt = now
	w.addHistory(WithdrawalStatusCompleted, "Withdrawal completed", nil)
}

// Fail marks the withdrawal as failed
func (w *Withdrawal) Fail(reason string) {
	w.Status = WithdrawalStatusFailed
	w.RejectionReason = reason
	w.UpdatedAt = time.Now()
	w.addHistory(WithdrawalStatusFailed, reason, nil)
}

// Cancel marks the withdrawal as canceled by the user
func (w *Withdrawal) Cancel() {
	w.Status = WithdrawalStatusCanceled
	w.UpdatedAt = time.Now()
	w.addHistory(WithdrawalStatusCanceled, "Canceled by user", nil)
}

// IsCancelable returns true if the withdrawal can be canceled
func (w *Withdrawal) IsCancelable() bool {
	return w.Status == WithdrawalStatusPending || w.Status == WithdrawalStatusReviewing
}

// IsPending returns true if the withdrawal is pending
func (w *Withdrawal) IsPending() bool {
	return w.Status == WithdrawalStatusPending
}

func (w *Withdrawal) addHistory(status WithdrawalStatus, reason string, updatedBy *uuid.UUID) {
	w.History = append(w.History, WithdrawalHistory{
		Status:    status,
		Reason:    reason,
		UpdatedBy: updatedBy,
		Timestamp: time.Now(),
	})
}

