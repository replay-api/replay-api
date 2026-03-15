package exchange_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
)

// ExchangeAdapter represents an external cryptocurrency exchange
// Implementations: CoinbaseAdapter, KrakenAdapter
type ExchangeAdapter interface {
	// GetProvider returns the exchange provider identifier
	GetProvider() exchange_vo.ExchangeProvider

	// PlaceMarketBuyOrder places a market buy order on the exchange
	// amountUSD is the USD amount to spend (exchange will return BTC)
	PlaceMarketBuyOrder(ctx context.Context, pair exchange_vo.ExchangePair, amountUSD float64) (*ExchangeOrderResult, error)

	// PlaceMarketSellOrder places a market sell order on the exchange
	// amountBTC is the BTC amount to sell
	PlaceMarketSellOrder(ctx context.Context, pair exchange_vo.ExchangePair, amountBTC float64) (*ExchangeOrderResult, error)

	// GetOrderStatus checks the status of an exchange order
	GetOrderStatus(ctx context.Context, orderID string) (*ExchangeOrderResult, error)

	// GetTicker returns the current ticker for a trading pair
	GetTicker(ctx context.Context, pair exchange_vo.ExchangePair) (*TickerResult, error)

	// GetAccountBalance returns balances on the exchange
	GetAccountBalance(ctx context.Context) (map[string]float64, error)

	// WithdrawBTC withdraws BTC from the exchange to an external address
	WithdrawBTC(ctx context.Context, address string, amountBTC float64) (*WithdrawResult, error)

	// HealthCheck verifies the exchange connection is healthy
	HealthCheck(ctx context.Context) error
}

// ExchangeOrderResult represents the outcome of an exchange order
type ExchangeOrderResult struct {
	OrderID      string                      `json:"order_id"`
	Status       string                      `json:"status"`
	FilledQtyBTC float64                     `json:"filled_qty_btc"`
	AvgPriceUSD  float64                     `json:"avg_price_usd"`
	FeeUSD       float64                     `json:"fee_usd"`
	FeeCurrency  string                      `json:"fee_currency"`
	Provider     exchange_vo.ExchangeProvider `json:"provider"`
	RawResponse  map[string]interface{}       `json:"raw_response,omitempty"`
}

// IsFilled returns true if the exchange order is fully filled
func (r *ExchangeOrderResult) IsFilled() bool {
	return r.Status == "filled" || r.Status == "done" || r.Status == "closed"
}

// TickerResult represents current market data for a trading pair
type TickerResult struct {
	Pair      exchange_vo.ExchangePair     `json:"pair"`
	Bid       float64                      `json:"bid"`
	Ask       float64                      `json:"ask"`
	Last      float64                      `json:"last"`
	Volume24h float64                      `json:"volume_24h"`
	High24h   float64                      `json:"high_24h"`
	Low24h    float64                      `json:"low_24h"`
	Provider  exchange_vo.ExchangeProvider `json:"provider"`
}

// MidPrice returns the midpoint between bid and ask
func (t *TickerResult) MidPrice() float64 {
	return (t.Bid + t.Ask) / 2.0
}

// WithdrawResult represents the result of a BTC withdrawal from an exchange
type WithdrawResult struct {
	WithdrawID string  `json:"withdraw_id"`
	Status     string  `json:"status"`
	AmountBTC  float64 `json:"amount_btc"`
	FeeBTC     float64 `json:"fee_btc"`
	TxHash     string  `json:"tx_hash,omitempty"`
}

// PriceFeedProvider is a source for BTC/USD price data
// Implementations: CoinGeckoAdapter, CoinbasePriceAdapter, KrakenPriceAdapter
type PriceFeedProvider interface {
	// GetBTCUSDPrice returns the current BTC/USD price
	GetBTCUSDPrice(ctx context.Context) (*exchange_entities.PricePoint, error)

	// GetProvider returns the provider name
	GetProvider() string
}

// OrderRepository persists exchange orders
type OrderRepository interface {
	Save(ctx context.Context, order *exchange_entities.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*exchange_entities.Order, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*exchange_entities.Order, int, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*exchange_entities.Order, error)
	Update(ctx context.Context, order *exchange_entities.Order) error
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*exchange_entities.Order, error)
	CountByUserIDSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)
}

// QuoteRepository persists exchange quotes
type QuoteRepository interface {
	Save(ctx context.Context, quote *exchange_entities.Quote) error
	FindByID(ctx context.Context, id uuid.UUID) (*exchange_entities.Quote, error)
	MarkConsumed(ctx context.Context, id uuid.UUID) error
}

// ExchangeRateRepository persists exchange rate history
type ExchangeRateRepository interface {
	Save(ctx context.Context, rate *exchange_entities.ExchangeRate) error
	FindLatest(ctx context.Context, pair exchange_vo.ExchangePair) (*exchange_entities.ExchangeRate, error)
	FindHistory(ctx context.Context, pair exchange_vo.ExchangePair, limit int) ([]*exchange_entities.ExchangeRate, error)
}

// RateCache provides fast access to cached exchange rates (Redis/Dragonfly)
type RateCache interface {
	SetRate(ctx context.Context, pair exchange_vo.ExchangePair, rate *exchange_entities.ExchangeRate) error
	GetRate(ctx context.Context, pair exchange_vo.ExchangePair) (*exchange_entities.ExchangeRate, error)
}

// ExchangeEventPublisher publishes exchange-related domain events
type ExchangeEventPublisher interface {
	PublishOrderCreated(ctx context.Context, order *exchange_entities.Order) error
	PublishOrderExecuting(ctx context.Context, order *exchange_entities.Order) error
	PublishOrderFilled(ctx context.Context, order *exchange_entities.Order) error
	PublishOrderFailed(ctx context.Context, order *exchange_entities.Order) error
	PublishOrderCancelled(ctx context.Context, order *exchange_entities.Order) error
	PublishQuoteCreated(ctx context.Context, quote *exchange_entities.Quote) error
	PublishPriceUpdated(ctx context.Context, rate *exchange_entities.ExchangeRate) error
}
