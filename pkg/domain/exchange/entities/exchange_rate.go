package exchange_entities

import (
	"time"

	"github.com/google/uuid"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ExchangeRate represents a cached BTC/USD exchange rate from multiple sources
type ExchangeRate struct {
	shared.BaseEntity `bson:"baseentity"`

	Pair        exchange_vo.ExchangePair `json:"pair" bson:"pair"`
	MedianPrice float64                  `json:"median_price" bson:"median_price"`
	Sources     []PricePoint             `json:"sources" bson:"sources"`
	Timestamp   time.Time                `json:"timestamp" bson:"timestamp"`
	Confidence  float64                  `json:"confidence" bson:"confidence"` // 0-1, based on source agreement
	Spread      float64                  `json:"spread" bson:"spread"`         // Max deviation between sources as %
}

// PricePoint represents a single price data point from one provider
type PricePoint struct {
	Provider  string    `json:"provider" bson:"provider"`
	Price     float64   `json:"price" bson:"price"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Volume24h float64   `json:"volume_24h,omitempty" bson:"volume_24h,omitempty"`
	Bid       float64   `json:"bid,omitempty" bson:"bid,omitempty"`
	Ask       float64   `json:"ask,omitempty" bson:"ask,omitempty"`
}

// NewExchangeRate creates a new exchange rate from multiple price sources
func NewExchangeRate(
	resourceOwner shared.ResourceOwner,
	pair exchange_vo.ExchangePair,
	medianPrice float64,
	sources []PricePoint,
	confidence float64,
	spread float64,
) *ExchangeRate {
	return &ExchangeRate{
		BaseEntity:  shared.NewUnrestrictedEntity(resourceOwner),
		Pair:        pair,
		MedianPrice: medianPrice,
		Sources:     sources,
		Timestamp:   time.Now().UTC(),
		Confidence:  confidence,
		Spread:      spread,
	}
}

// IsStale returns true if the rate is older than the given duration
func (r *ExchangeRate) IsStale(maxAge time.Duration) bool {
	return time.Since(r.Timestamp) > maxAge
}

// IsReliable returns true if the rate has sufficient confidence
func (r *ExchangeRate) IsReliable(minConfidence float64) bool {
	return r.Confidence >= minConfidence
}

// PricingResult holds the computed pricing from multiple sources
type PricingResult struct {
	MedianPrice float64      `json:"median_price"`
	Sources     []PricePoint `json:"sources"`
	Timestamp   time.Time    `json:"timestamp"`
	Confidence  float64      `json:"confidence"`
	Spread      float64      `json:"spread"`
	RateID      uuid.UUID    `json:"rate_id"`
}
