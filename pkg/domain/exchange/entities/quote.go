package exchange_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// Quote represents a locked price quote for a Bitcoin trade
// Quotes are immutable once created and expire after a configurable TTL (default 30s)
// Invariants:
//   - A quote cannot be modified after creation
//   - A quote can only be consumed once
//   - An expired quote must be rejected
type Quote struct {
	shared.BaseEntity `bson:"baseentity"`

	UserID uuid.UUID              `json:"user_id" bson:"user_id"`
	Side   exchange_vo.OrderSide  `json:"side" bson:"side"`
	Pair   exchange_vo.ExchangePair `json:"pair" bson:"pair"`

	// Price information
	BTCPriceUSD float64 `json:"btc_price_usd" bson:"btc_price_usd"` // Locked BTC/USD price

	// For buy quotes: user specifies USD amount
	AmountUSD wallet_vo.Amount    `json:"amount_usd" bson:"amount_usd"`
	BTCAmount wallet_vo.BtcAmount `json:"btc_amount" bson:"btc_amount"` // Calculated BTC the user will receive/send

	// Fee breakdown
	FeePercent float64         `json:"fee_percent" bson:"fee_percent"`
	FeeAmount  wallet_vo.Amount `json:"fee_amount" bson:"fee_amount"`

	// Total cost (buy) or net proceeds (sell)
	TotalCostUSD    wallet_vo.Amount `json:"total_cost_usd,omitempty" bson:"total_cost_usd,omitempty"`       // Buy: amount + fee
	NetProceedsUSD  wallet_vo.Amount `json:"net_proceeds_usd,omitempty" bson:"net_proceeds_usd,omitempty"` // Sell: amount - fee

	// Expiry
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
	Consumed  bool      `json:"consumed" bson:"consumed"`

	// Source price data
	PriceSource   string  `json:"price_source" bson:"price_source"`     // "median" or specific provider
	PriceConfidence float64 `json:"price_confidence" bson:"price_confidence"` // 0-1, how confident the price is
}

// NewBuyQuote creates a new buy quote (user pays USD, receives BTC)
func NewBuyQuote(
	resourceOwner shared.ResourceOwner,
	userID uuid.UUID,
	amountUSD wallet_vo.Amount,
	btcPriceUSD float64,
	feePercent float64,
	ttl time.Duration,
	priceSource string,
	priceConfidence float64,
) (*Quote, error) {
	if amountUSD.IsZero() || amountUSD.IsNegative() {
		return nil, fmt.Errorf("amount must be positive, got: %s", amountUSD.String())
	}
	if btcPriceUSD <= 0 {
		return nil, fmt.Errorf("BTC price must be positive, got: %f", btcPriceUSD)
	}

	feeAmount := wallet_vo.NewAmount(amountUSD.Dollars() * (feePercent / 100.0))
	netForExchange := amountUSD.Subtract(feeAmount)
	btcAmount := wallet_vo.FromUSD(netForExchange, btcPriceUSD)
	totalCost := amountUSD // User pays the full amount (fee is taken from it)

	quote := &Quote{
		BaseEntity:      shared.NewPrivateEntity(resourceOwner),
		UserID:          userID,
		Side:            exchange_vo.OrderSideBuy,
		Pair:            exchange_vo.PairBTCUSD,
		BTCPriceUSD:     btcPriceUSD,
		AmountUSD:       amountUSD,
		BTCAmount:       btcAmount,
		FeePercent:      feePercent,
		FeeAmount:       feeAmount,
		TotalCostUSD:    totalCost,
		ExpiresAt:       time.Now().UTC().Add(ttl),
		Consumed:        false,
		PriceSource:     priceSource,
		PriceConfidence: priceConfidence,
	}

	return quote, nil
}

// NewSellQuote creates a new sell quote (user sends BTC, receives USD)
func NewSellQuote(
	resourceOwner shared.ResourceOwner,
	userID uuid.UUID,
	amountBTC wallet_vo.BtcAmount,
	btcPriceUSD float64,
	feePercent float64,
	ttl time.Duration,
	priceSource string,
	priceConfidence float64,
) (*Quote, error) {
	if amountBTC.IsZero() || amountBTC.IsNegative() {
		return nil, fmt.Errorf("BTC amount must be positive, got: %s", amountBTC.String())
	}
	if btcPriceUSD <= 0 {
		return nil, fmt.Errorf("BTC price must be positive, got: %f", btcPriceUSD)
	}

	grossUSD := amountBTC.ToUSD(btcPriceUSD)
	feeAmount := wallet_vo.NewAmount(grossUSD.Dollars() * (feePercent / 100.0))
	netProceeds := grossUSD.Subtract(feeAmount)

	quote := &Quote{
		BaseEntity:      shared.NewPrivateEntity(resourceOwner),
		UserID:          userID,
		Side:            exchange_vo.OrderSideSell,
		Pair:            exchange_vo.PairBTCUSD,
		BTCPriceUSD:     btcPriceUSD,
		AmountUSD:       grossUSD,
		BTCAmount:       amountBTC,
		FeePercent:      feePercent,
		FeeAmount:       feeAmount,
		NetProceedsUSD:  netProceeds,
		ExpiresAt:       time.Now().UTC().Add(ttl),
		Consumed:        false,
		PriceSource:     priceSource,
		PriceConfidence: priceConfidence,
	}

	return quote, nil
}

// IsExpired returns true if the quote has expired
func (q *Quote) IsExpired() bool {
	return time.Now().UTC().After(q.ExpiresAt)
}

// IsUsable returns true if the quote can be used for an order
func (q *Quote) IsUsable() bool {
	return !q.IsExpired() && !q.Consumed
}

// MarkConsumed marks the quote as consumed (used for an order)
func (q *Quote) MarkConsumed() error {
	if q.Consumed {
		return fmt.Errorf("quote %s has already been consumed", q.ID.String())
	}
	if q.IsExpired() {
		return fmt.Errorf("quote %s has expired at %s", q.ID.String(), q.ExpiresAt.Format(time.RFC3339))
	}
	q.Consumed = true
	q.UpdatedAt = time.Now()
	return nil
}

// RemainingSeconds returns how many seconds until the quote expires
func (q *Quote) RemainingSeconds() int {
	remaining := time.Until(q.ExpiresAt).Seconds()
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}

// Validate ensures quote invariants
func (q *Quote) Validate() error {
	if q.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if !q.Side.IsValid() {
		return fmt.Errorf("invalid side: %s", q.Side)
	}
	if q.BTCPriceUSD <= 0 {
		return fmt.Errorf("BTC price must be positive")
	}
	return nil
}
