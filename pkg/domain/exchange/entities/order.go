package exchange_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// Order represents a Bitcoin buy or sell order aggregate root
// Invariants:
//   - An order cannot transition to a non-valid next state
//   - A completed order must have ExecutedAmountBTC > 0 and ExecutedPriceUSD > 0
//   - IdempotencyKey must be unique per user
type Order struct {
	shared.BaseEntity `bson:"baseentity"`

	// Core identification
	UserID   uuid.UUID `json:"user_id" bson:"user_id"`
	WalletID uuid.UUID `json:"wallet_id" bson:"wallet_id"`

	// Order parameters
	Side OrderSide            `json:"side" bson:"side"`
	Pair exchange_vo.ExchangePair `json:"pair" bson:"pair"`

	// Requested amounts (input)
	RequestedAmountUSD wallet_vo.Amount    `json:"requested_amount_usd" bson:"requested_amount_usd"`
	RequestedAmountBTC wallet_vo.BtcAmount `json:"requested_amount_btc,omitempty" bson:"requested_amount_btc,omitempty"`

	// Executed amounts (output — filled by exchange)
	ExecutedAmountBTC wallet_vo.BtcAmount `json:"executed_amount_btc" bson:"executed_amount_btc"`
	ExecutedPriceUSD  float64             `json:"executed_price_usd" bson:"executed_price_usd"`

	// Exchange routing
	ExchangeProvider exchange_vo.ExchangeProvider `json:"exchange_provider" bson:"exchange_provider"`
	ExchangeOrderID  string                       `json:"exchange_order_id,omitempty" bson:"exchange_order_id,omitempty"`

	// Status
	Status OrderStatus `json:"status" bson:"status"`

	// Fees
	FeePercent    float64         `json:"fee_percent" bson:"fee_percent"`
	FeeAmountUSD  wallet_vo.Amount `json:"fee_amount_usd" bson:"fee_amount_usd"`
	NetAmountUSD  wallet_vo.Amount `json:"net_amount_usd" bson:"net_amount_usd"` // For buy: amount sent to exchange. For sell: amount credited to user.

	// Payment references
	StripePaymentIntentID string     `json:"stripe_payment_intent_id,omitempty" bson:"stripe_payment_intent_id,omitempty"` // For buy orders
	QuoteID               *uuid.UUID `json:"quote_id,omitempty" bson:"quote_id,omitempty"`

	// Idempotency
	IdempotencyKey string `json:"idempotency_key" bson:"idempotency_key"`

	// Timestamps
	ExecutedAt    *time.Time `json:"executed_at,omitempty" bson:"executed_at,omitempty"`
	SettledAt     *time.Time `json:"settled_at,omitempty" bson:"settled_at,omitempty"`
	FailedAt      *time.Time `json:"failed_at,omitempty" bson:"failed_at,omitempty"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`
	FailureReason string     `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"`

	// Retry tracking
	RetryCount int `json:"retry_count" bson:"retry_count"`
	MaxRetries int `json:"max_retries" bson:"max_retries"`
}

// OrderSide wraps exchange_vo.OrderSide for local use
type OrderSide = exchange_vo.OrderSide

// OrderStatus wraps exchange_vo.OrderStatus for local use
type OrderStatus = exchange_vo.OrderStatus

// NewBuyOrder creates a new buy order
func NewBuyOrder(
	resourceOwner shared.ResourceOwner,
	userID, walletID uuid.UUID,
	amountUSD wallet_vo.Amount,
	feePercent float64,
	feeAmountUSD wallet_vo.Amount,
	idempotencyKey string,
	quoteID *uuid.UUID,
) (*Order, error) {
	if amountUSD.IsZero() || amountUSD.IsNegative() {
		return nil, fmt.Errorf("buy amount must be positive, got: %s", amountUSD.String())
	}
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key is required")
	}

	netAmount := amountUSD.Subtract(feeAmountUSD)
	if netAmount.IsNegative() || netAmount.IsZero() {
		return nil, fmt.Errorf("net amount after fees must be positive (amount: %s, fee: %s)", amountUSD.String(), feeAmountUSD.String())
	}

	order := &Order{
		BaseEntity:         shared.NewPrivateEntity(resourceOwner),
		UserID:             userID,
		WalletID:           walletID,
		Side:               exchange_vo.OrderSideBuy,
		Pair:               exchange_vo.PairBTCUSD,
		RequestedAmountUSD: amountUSD,
		ExecutedAmountBTC:  wallet_vo.NewBtcAmountFromSatoshis(0),
		Status:             exchange_vo.OrderStatusPending,
		FeePercent:         feePercent,
		FeeAmountUSD:       feeAmountUSD,
		NetAmountUSD:       netAmount,
		QuoteID:            quoteID,
		IdempotencyKey:     idempotencyKey,
		RetryCount:         0,
		MaxRetries:         3,
	}

	return order, nil
}

// NewSellOrder creates a new sell order
func NewSellOrder(
	resourceOwner shared.ResourceOwner,
	userID, walletID uuid.UUID,
	amountBTC wallet_vo.BtcAmount,
	estimatedUSD wallet_vo.Amount,
	feePercent float64,
	feeAmountUSD wallet_vo.Amount,
	idempotencyKey string,
	quoteID *uuid.UUID,
) (*Order, error) {
	if amountBTC.IsZero() || amountBTC.IsNegative() {
		return nil, fmt.Errorf("sell amount must be positive, got: %s", amountBTC.String())
	}
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key is required")
	}

	netAmount := estimatedUSD.Subtract(feeAmountUSD)

	order := &Order{
		BaseEntity:         shared.NewPrivateEntity(resourceOwner),
		UserID:             userID,
		WalletID:           walletID,
		Side:               exchange_vo.OrderSideSell,
		Pair:               exchange_vo.PairBTCUSD,
		RequestedAmountBTC: amountBTC,
		RequestedAmountUSD: estimatedUSD,
		ExecutedAmountBTC:  wallet_vo.NewBtcAmountFromSatoshis(0),
		Status:             exchange_vo.OrderStatusPending,
		FeePercent:         feePercent,
		FeeAmountUSD:       feeAmountUSD,
		NetAmountUSD:       netAmount,
		QuoteID:            quoteID,
		IdempotencyKey:     idempotencyKey,
		RetryCount:         0,
		MaxRetries:         3,
	}

	return order, nil
}

// MarkPaymentHeld transitions to PAYMENT_HELD (Stripe captured)
func (o *Order) MarkPaymentHeld(stripePaymentIntentID string) error {
	if o.Status != exchange_vo.OrderStatusPending {
		return fmt.Errorf("cannot mark payment held: current status is %s, expected PENDING", o.Status)
	}
	o.StripePaymentIntentID = stripePaymentIntentID
	o.Status = exchange_vo.OrderStatusPaymentHeld
	o.UpdatedAt = time.Now()
	return nil
}

// MarkExecuting transitions to EXECUTING (sent to exchange)
func (o *Order) MarkExecuting(provider exchange_vo.ExchangeProvider, exchangeOrderID string) error {
	if o.Status != exchange_vo.OrderStatusPaymentHeld && o.Status != exchange_vo.OrderStatusPending {
		return fmt.Errorf("cannot mark executing: current status is %s", o.Status)
	}
	o.ExchangeProvider = provider
	o.ExchangeOrderID = exchangeOrderID
	o.Status = exchange_vo.OrderStatusExecuting
	o.UpdatedAt = time.Now()
	return nil
}

// MarkFilled transitions to FILLED (exchange order completed)
func (o *Order) MarkFilled(executedBTC wallet_vo.BtcAmount, executedPriceUSD float64) error {
	if o.Status != exchange_vo.OrderStatusExecuting && o.Status != exchange_vo.OrderStatusPartialFill {
		return fmt.Errorf("cannot mark filled: current status is %s", o.Status)
	}
	now := time.Now()
	o.ExecutedAmountBTC = executedBTC
	o.ExecutedPriceUSD = executedPriceUSD
	o.Status = exchange_vo.OrderStatusFilled
	o.ExecutedAt = &now
	o.UpdatedAt = now
	return nil
}

// MarkCompleted transitions to COMPLETED (wallet credited)
func (o *Order) MarkCompleted() error {
	if o.Status != exchange_vo.OrderStatusFilled && o.Status != exchange_vo.OrderStatusSettling {
		return fmt.Errorf("cannot mark completed: current status is %s", o.Status)
	}
	now := time.Now()
	o.Status = exchange_vo.OrderStatusCompleted
	o.SettledAt = &now
	o.UpdatedAt = now
	return nil
}

// MarkFailed transitions to FAILED
func (o *Order) MarkFailed(reason string) error {
	if o.Status.IsTerminal() {
		return fmt.Errorf("cannot mark failed: order is already in terminal status %s", o.Status)
	}
	now := time.Now()
	o.Status = exchange_vo.OrderStatusFailed
	o.FailureReason = reason
	o.FailedAt = &now
	o.UpdatedAt = now
	return nil
}

// MarkCancelled transitions to CANCELLED
func (o *Order) MarkCancelled() error {
	if o.Status.IsTerminal() {
		return fmt.Errorf("cannot cancel: order is already in terminal status %s", o.Status)
	}
	if o.Status == exchange_vo.OrderStatusFilled || o.Status == exchange_vo.OrderStatusSettling {
		return fmt.Errorf("cannot cancel: order is already filled/settling")
	}
	now := time.Now()
	o.Status = exchange_vo.OrderStatusCancelled
	o.CancelledAt = &now
	o.UpdatedAt = now
	return nil
}

// CanRetry checks if the order can be retried
func (o *Order) CanRetry() bool {
	return o.Status == exchange_vo.OrderStatusFailed && o.RetryCount < o.MaxRetries
}

// IncrementRetry increments the retry counter and resets to PENDING
func (o *Order) IncrementRetry() error {
	if !o.CanRetry() {
		return fmt.Errorf("cannot retry: max retries (%d) reached or status is %s", o.MaxRetries, o.Status)
	}
	o.RetryCount++
	o.Status = exchange_vo.OrderStatusPending
	o.FailureReason = ""
	o.FailedAt = nil
	o.UpdatedAt = time.Now()
	return nil
}

// NeedsRefund returns true if a buy order failed after Stripe payment was captured
func (o *Order) NeedsRefund() bool {
	return o.Side == exchange_vo.OrderSideBuy &&
		o.StripePaymentIntentID != "" &&
		(o.Status == exchange_vo.OrderStatusFailed || o.Status == exchange_vo.OrderStatusCancelled)
}

// Validate ensures order invariants
func (o *Order) Validate() error {
	if o.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if o.WalletID == uuid.Nil {
		return fmt.Errorf("wallet_id is required")
	}
	if !o.Side.IsValid() {
		return fmt.Errorf("invalid order side: %s", o.Side)
	}
	if !o.Pair.IsValid() {
		return fmt.Errorf("invalid trading pair: %s", o.Pair)
	}
	if o.IdempotencyKey == "" {
		return fmt.Errorf("idempotency key is required")
	}
	if o.FeePercent < 0 {
		return fmt.Errorf("fee percent must be non-negative")
	}
	return nil
}
