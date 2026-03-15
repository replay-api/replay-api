package exchange_vo

// OrderSide represents the direction of a trade
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// IsValid checks if the order side is valid
func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}

// String returns string representation
func (s OrderSide) String() string {
	return string(s)
}

// OrderStatus represents the lifecycle state of an exchange order
type OrderStatus string

const (
	OrderStatusPending     OrderStatus = "PENDING"       // Created, awaiting payment/execution
	OrderStatusPaymentHeld OrderStatus = "PAYMENT_HELD"  // Stripe payment captured, awaiting exchange execution
	OrderStatusExecuting   OrderStatus = "EXECUTING"     // Sent to exchange, awaiting fill
	OrderStatusPartialFill OrderStatus = "PARTIAL_FILL"  // Partially filled on exchange
	OrderStatusFilled      OrderStatus = "FILLED"        // Fully filled on exchange
	OrderStatusSettling    OrderStatus = "SETTLING"       // Crediting user wallet
	OrderStatusCompleted   OrderStatus = "COMPLETED"     // Fully settled in user wallet
	OrderStatusFailed      OrderStatus = "FAILED"        // Failed at any stage
	OrderStatusCancelled   OrderStatus = "CANCELLED"     // Cancelled by user or system
	OrderStatusRefunding   OrderStatus = "REFUNDING"     // Refund in progress (for buys)
	OrderStatusRefunded    OrderStatus = "REFUNDED"      // Stripe refund completed
)

// IsTerminal returns true if the order is in a final state
func (s OrderStatus) IsTerminal() bool {
	return s == OrderStatusCompleted || s == OrderStatusFailed ||
		s == OrderStatusCancelled || s == OrderStatusRefunded
}

// IsActive returns true if the order is still in progress
func (s OrderStatus) IsActive() bool {
	return !s.IsTerminal()
}

// IsValid checks if the status is a recognized value
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusPaymentHeld, OrderStatusExecuting,
		OrderStatusPartialFill, OrderStatusFilled, OrderStatusSettling,
		OrderStatusCompleted, OrderStatusFailed, OrderStatusCancelled,
		OrderStatusRefunding, OrderStatusRefunded:
		return true
	default:
		return false
	}
}

// ExchangePair represents a trading pair
type ExchangePair string

const (
	PairBTCUSD ExchangePair = "BTC/USD"
)

// IsValid checks if the pair is supported
func (p ExchangePair) IsValid() bool {
	return p == PairBTCUSD
}

// Base returns the base currency (e.g., "BTC")
func (p ExchangePair) Base() string {
	return "BTC"
}

// Quote returns the quote currency (e.g., "USD")
func (p ExchangePair) Quote() string {
	return "USD"
}

// ExchangeProvider identifies the exchange used to fill an order
type ExchangeProvider string

const (
	ExchangeProviderCoinbase ExchangeProvider = "coinbase"
	ExchangeProviderKraken   ExchangeProvider = "kraken"
)

// IsValid checks if the provider is supported
func (p ExchangeProvider) IsValid() bool {
	return p == ExchangeProviderCoinbase || p == ExchangeProviderKraken
}
