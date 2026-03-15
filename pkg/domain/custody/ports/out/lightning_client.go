package custody_out

import (
	"context"
	"time"
)

// LNInvoice represents a Lightning Network invoice
type LNInvoice struct {
	PaymentRequest string    `json:"payment_request"` // BOLT11 encoded invoice
	PaymentHash    string    `json:"payment_hash"`
	AmountSats     int64     `json:"amount_sats"`
	Description    string    `json:"description,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	Settled        bool      `json:"settled"`
}

// LNPayment represents a completed Lightning payment
type LNPayment struct {
	PaymentHash     string     `json:"payment_hash"`
	PaymentPreimage string     `json:"payment_preimage"`
	AmountSats      int64      `json:"amount_sats"`
	FeeSats         int64      `json:"fee_sats"`
	Status          string     `json:"status"` // "pending", "succeeded", "failed"
	CreatedAt       time.Time  `json:"created_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

// LNBalance represents the Lightning node's balance
type LNBalance struct {
	LocalBalanceSats   int64 `json:"local_balance_sats"`
	RemoteBalanceSats  int64 `json:"remote_balance_sats"`
	PendingSendSats    int64 `json:"pending_send_sats"`
	PendingRecvSats    int64 `json:"pending_recv_sats"`
	NumActiveChannels  int   `json:"num_active_channels"`
	NumPendingChannels int   `json:"num_pending_channels"`
}

// DecodedInvoice is a decoded BOLT11 invoice
type DecodedInvoice struct {
	PaymentHash string    `json:"payment_hash"`
	AmountSats  int64     `json:"amount_sats"`
	Description string    `json:"description"`
	Destination string    `json:"destination"` // Node public key
	ExpiresAt   time.Time `json:"expires_at"`
	IsExpired   bool      `json:"is_expired"`
}

// LightningClient provides Lightning Network operations
type LightningClient interface {
	// CreateInvoice creates a new Lightning invoice (for receiving)
	CreateInvoice(ctx context.Context, amountSats int64, memo string, expiry time.Duration) (*LNInvoice, error)

	// PayInvoice pays a Lightning invoice (for sending)
	PayInvoice(ctx context.Context, paymentRequest string, maxFeeSats int64) (*LNPayment, error)

	// GetInvoiceStatus checks the status of an invoice
	GetInvoiceStatus(ctx context.Context, paymentHash string) (*LNInvoice, error)

	// GetPaymentStatus checks the status of an outgoing payment
	GetPaymentStatus(ctx context.Context, paymentHash string) (*LNPayment, error)

	// GetBalance returns the Lightning node balance
	GetBalance(ctx context.Context) (*LNBalance, error)

	// DecodeInvoice decodes a BOLT11 invoice without paying it
	DecodeInvoice(ctx context.Context, paymentRequest string) (*DecodedInvoice, error)

	// HealthCheck verifies the Lightning node connection
	HealthCheck(ctx context.Context) error
}
