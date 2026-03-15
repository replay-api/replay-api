package exchange_in

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
)

// --- Commands ---

// BuyBitcoinCommand represents a request to buy Bitcoin with USD (via Stripe)
type BuyBitcoinCommand struct {
	UserID              uuid.UUID  `json:"user_id"`
	WalletID            uuid.UUID  `json:"wallet_id"`
	AmountUSD           float64    `json:"amount_usd"`
	QuoteID             *uuid.UUID `json:"quote_id,omitempty"`
	StripePaymentMethod string     `json:"stripe_payment_method"`
	IdempotencyKey      string     `json:"idempotency_key"`
}

// Validate validates the buy command
func (c *BuyBitcoinCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if c.WalletID == uuid.Nil {
		return fmt.Errorf("wallet_id is required")
	}
	if c.AmountUSD <= 0 {
		return fmt.Errorf("amount_usd must be positive")
	}
	if c.AmountUSD < 1.00 {
		return fmt.Errorf("minimum buy amount is $1.00")
	}
	if c.AmountUSD > 50000.00 {
		return fmt.Errorf("maximum buy amount is $50,000.00")
	}
	if c.StripePaymentMethod == "" {
		return fmt.Errorf("stripe_payment_method is required")
	}
	if c.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	return nil
}

// SellBitcoinCommand represents a request to sell Bitcoin for USD credits
type SellBitcoinCommand struct {
	UserID         uuid.UUID  `json:"user_id"`
	WalletID       uuid.UUID  `json:"wallet_id"`
	AmountBTC      float64    `json:"amount_btc"`
	QuoteID        *uuid.UUID `json:"quote_id,omitempty"`
	IdempotencyKey string     `json:"idempotency_key"`
}

// Validate validates the sell command
func (c *SellBitcoinCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if c.WalletID == uuid.Nil {
		return fmt.Errorf("wallet_id is required")
	}
	if c.AmountBTC <= 0 {
		return fmt.Errorf("amount_btc must be positive")
	}
	if c.AmountBTC < 0.00001 {
		return fmt.Errorf("minimum sell amount is 0.00001 BTC (1000 sats)")
	}
	if c.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	return nil
}

// CancelOrderCommand represents a request to cancel a pending order
type CancelOrderCommand struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
}

// Validate validates the cancel command
func (c *CancelOrderCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if c.OrderID == uuid.Nil {
		return fmt.Errorf("order_id is required")
	}
	return nil
}

// --- Command Handler Interface ---

// ExchangeCommandHandler defines the interface for exchange mutation operations
type ExchangeCommandHandler interface {
	BuyBitcoin(ctx context.Context, cmd BuyBitcoinCommand) (*BuyBitcoinResult, error)
	SellBitcoin(ctx context.Context, cmd SellBitcoinCommand) (*SellBitcoinResult, error)
	CancelOrder(ctx context.Context, cmd CancelOrderCommand) error
}

// BuyBitcoinResult is returned after initiating a buy order
type BuyBitcoinResult struct {
	OrderID               uuid.UUID `json:"order_id"`
	Status                string    `json:"status"`
	AmountUSD             float64   `json:"amount_usd"`
	EstimatedBTC          float64   `json:"estimated_btc"`
	FeeUSD                float64   `json:"fee_usd"`
	FeePercent            float64   `json:"fee_percent"`
	StripeClientSecret    string    `json:"stripe_client_secret,omitempty"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id,omitempty"`
}

// SellBitcoinResult is returned after initiating a sell order
type SellBitcoinResult struct {
	OrderID        uuid.UUID `json:"order_id"`
	Status         string    `json:"status"`
	AmountBTC      float64   `json:"amount_btc"`
	EstimatedUSD   float64   `json:"estimated_usd"`
	FeeUSD         float64   `json:"fee_usd"`
	FeePercent     float64   `json:"fee_percent"`
	NetProceedsUSD float64   `json:"net_proceeds_usd"`
}

// --- Query Types ---

// GetQuoteQuery requests a price quote for buying or selling BTC
type GetQuoteQuery struct {
	UserID    uuid.UUID             `json:"user_id"`
	Side      exchange_vo.OrderSide `json:"side"`
	AmountUSD float64               `json:"amount_usd,omitempty"`
	AmountBTC float64               `json:"amount_btc,omitempty"`
}

// Validate validates the quote query
func (q *GetQuoteQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if !q.Side.IsValid() {
		return fmt.Errorf("invalid side: %s", q.Side)
	}
	if q.Side == exchange_vo.OrderSideBuy && q.AmountUSD <= 0 {
		return fmt.Errorf("amount_usd is required for buy quotes")
	}
	if q.Side == exchange_vo.OrderSideSell && q.AmountBTC <= 0 {
		return fmt.Errorf("amount_btc is required for sell quotes")
	}
	return nil
}

// --- Query Handler Interface ---

// ExchangeQueryHandler defines the interface for exchange read operations
type ExchangeQueryHandler interface {
	GetQuote(ctx context.Context, query GetQuoteQuery) (*QuoteResult, error)
	GetExchangeRates(ctx context.Context) (*RatesResult, error)
	GetOrderHistory(ctx context.Context, userID uuid.UUID, limit, offset int) (*OrderHistoryResult, error)
	GetOrderByID(ctx context.Context, userID, orderID uuid.UUID) (*OrderDetailResult, error)
	GetFeeSchedule(ctx context.Context, userID uuid.UUID) (*FeeScheduleResult, error)
}

// --- Result Types ---

// QuoteResult is returned for a price quote
type QuoteResult struct {
	QuoteID          uuid.UUID `json:"quote_id"`
	Side             string    `json:"side"`
	BTCPriceUSD      float64   `json:"btc_price_usd"`
	AmountUSD        float64   `json:"amount_usd"`
	BTCAmount        float64   `json:"btc_amount"`
	FeePercent       float64   `json:"fee_percent"`
	FeeAmountUSD     float64   `json:"fee_amount_usd"`
	TotalCostUSD     float64   `json:"total_cost_usd,omitempty"`
	NetProceedsUSD   float64   `json:"net_proceeds_usd,omitempty"`
	ExpiresAt        string    `json:"expires_at"`
	RemainingSeconds int       `json:"remaining_seconds"`
	PriceSource      string    `json:"price_source"`
	Confidence       float64   `json:"confidence"`
}

// RatesResult holds current exchange rates
type RatesResult struct {
	BTCUSD      float64           `json:"btc_usd"`
	Change24h   float64           `json:"change_24h_percent,omitempty"`
	High24h     float64           `json:"high_24h,omitempty"`
	Low24h      float64           `json:"low_24h,omitempty"`
	Volume24h   float64           `json:"volume_24h,omitempty"`
	LastUpdated string            `json:"last_updated"`
	Sources     []PriceSourceInfo `json:"sources"`
}

// PriceSourceInfo holds info about a price source
type PriceSourceInfo struct {
	Provider  string  `json:"provider"`
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
}

// OrderHistoryResult holds paginated order history
type OrderHistoryResult struct {
	Orders     []OrderSummary `json:"orders"`
	TotalCount int            `json:"total_count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

// OrderSummary is a brief view of an order
type OrderSummary struct {
	OrderID     uuid.UUID `json:"order_id"`
	Side        string    `json:"side"`
	Status      string    `json:"status"`
	AmountUSD   float64   `json:"amount_usd"`
	AmountBTC   float64   `json:"amount_btc"`
	FeeUSD      float64   `json:"fee_usd"`
	BTCPriceUSD float64   `json:"btc_price_usd"`
	CreatedAt   string    `json:"created_at"`
	CompletedAt string    `json:"completed_at,omitempty"`
}

// OrderDetailResult is a detailed view of a single order
type OrderDetailResult struct {
	OrderSummary
	ExchangeProvider      string  `json:"exchange_provider,omitempty"`
	ExchangeOrderID       string  `json:"exchange_order_id,omitempty"`
	StripePaymentIntentID string  `json:"stripe_payment_intent_id,omitempty"`
	FeePercent            float64 `json:"fee_percent"`
	NetAmountUSD          float64 `json:"net_amount_usd"`
	FailureReason         string  `json:"failure_reason,omitempty"`
	RetryCount            int     `json:"retry_count"`
}

// FeeScheduleResult holds the fee schedule for a user
type FeeScheduleResult struct {
	PlanTier       string                           `json:"plan_tier"`
	BuyFeePercent  float64                          `json:"buy_fee_percent"`
	SellFeePercent float64                          `json:"sell_fee_percent"`
	MinFeeUSD      float64                          `json:"min_fee_usd"`
	MaxFeeUSD      float64                          `json:"max_fee_usd"`
	AllTiers       map[string]exchange_vo.FeeConfig `json:"all_tiers"`
}
