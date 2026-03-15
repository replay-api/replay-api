package wallet_vo

import (
	"fmt"
	"strings"
	"time"
)

// LightningInvoice represents a BOLT11 Lightning Network payment request
type LightningInvoice struct {
	PaymentRequest string    `json:"payment_request" bson:"payment_request"`
	PaymentHash    string    `json:"payment_hash" bson:"payment_hash"`
	AmountSats     int64     `json:"amount_sats" bson:"amount_sats"`
	Description    string    `json:"description,omitempty" bson:"description,omitempty"`
	ExpiresAt      time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
}

// LightningInvoiceStatus represents the status of a Lightning invoice
type LightningInvoiceStatus string

const (
	LightningInvoiceStatusOpen     LightningInvoiceStatus = "OPEN"
	LightningInvoiceStatusSettled  LightningInvoiceStatus = "SETTLED"
	LightningInvoiceStatusCanceled LightningInvoiceStatus = "CANCELED"
	LightningInvoiceStatusExpired  LightningInvoiceStatus = "EXPIRED"
)

// NewLightningInvoice creates a new Lightning invoice with validation
func NewLightningInvoice(paymentRequest string, amountSats int64, description string, expiresAt time.Time) (*LightningInvoice, error) {
	paymentRequest = strings.TrimSpace(paymentRequest)

	if err := ValidateBOLT11(paymentRequest); err != nil {
		return nil, err
	}

	if amountSats < 0 {
		return nil, fmt.Errorf("lightning invoice amount must be non-negative, got: %d sats", amountSats)
	}

	// Max single Lightning payment: 0.04294967295 BTC (~4.29M sats) per BOLT spec
	const maxLightningPaymentSats = 4_294_967_295
	if amountSats > maxLightningPaymentSats {
		return nil, fmt.Errorf("lightning invoice amount %d sats exceeds maximum (%d sats)", amountSats, maxLightningPaymentSats)
	}

	return &LightningInvoice{
		PaymentRequest: paymentRequest,
		AmountSats:     amountSats,
		Description:    description,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

// IsExpired returns true if the invoice has expired
func (li *LightningInvoice) IsExpired() bool {
	return time.Now().UTC().After(li.ExpiresAt)
}

// ToBtcAmount converts the invoice amount to a BtcAmount
func (li *LightningInvoice) ToBtcAmount() BtcAmount {
	return NewBtcAmountFromSatoshis(li.AmountSats)
}

// ValidateBOLT11 validates a BOLT11 payment request format
// BOLT11 invoices start with "lnbc" (mainnet), "lntb" (testnet), or "lnbcrt" (regtest)
func ValidateBOLT11(paymentRequest string) error {
	paymentRequest = strings.TrimSpace(paymentRequest)
	lower := strings.ToLower(paymentRequest)

	if lower == "" {
		return fmt.Errorf("lightning payment request cannot be empty")
	}

	// Check prefix
	validPrefixes := []string{"lnbc", "lntb", "lnbcrt", "lnsb"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(lower, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("invalid Lightning invoice: must start with lnbc (mainnet), lntb (testnet), lnbcrt (regtest), or lnsb (signet)")
	}

	// Minimum length check (a valid BOLT11 is at least ~90 chars)
	if len(paymentRequest) < 90 {
		return fmt.Errorf("invalid Lightning invoice: too short (minimum 90 characters)")
	}

	// Maximum length check
	if len(paymentRequest) > 7089 {
		return fmt.Errorf("invalid Lightning invoice: too long (maximum 7089 characters)")
	}

	return nil
}

// LightningPayment represents a completed or in-progress Lightning payment
type LightningPayment struct {
	PaymentHash    string                 `json:"payment_hash" bson:"payment_hash"`
	PaymentPreimage string               `json:"payment_preimage,omitempty" bson:"payment_preimage,omitempty"`
	AmountSats     int64                  `json:"amount_sats" bson:"amount_sats"`
	FeeSats        int64                  `json:"fee_sats" bson:"fee_sats"`
	Status         LightningInvoiceStatus `json:"status" bson:"status"`
	CreatedAt      time.Time              `json:"created_at" bson:"created_at"`
	SettledAt      *time.Time             `json:"settled_at,omitempty" bson:"settled_at,omitempty"`
}

// LightningBalance represents the balance of a Lightning node
type LightningBalance struct {
	LocalBalanceSats  int64 `json:"local_balance_sats" bson:"local_balance_sats"`
	RemoteBalanceSats int64 `json:"remote_balance_sats" bson:"remote_balance_sats"`
	PendingSats       int64 `json:"pending_sats" bson:"pending_sats"`
	NumActiveChannels int   `json:"num_active_channels" bson:"num_active_channels"`
}
