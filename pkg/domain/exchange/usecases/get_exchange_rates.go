package exchange_usecases

import (
	"context"
	"fmt"
	"time"

	exchange_in "github.com/replay-api/replay-api/pkg/domain/exchange/ports/in"
	exchange_services "github.com/replay-api/replay-api/pkg/domain/exchange/services"
)

// GetExchangeRatesUseCase returns current BTC/USD exchange rates
type GetExchangeRatesUseCase struct {
	pricingService *exchange_services.PricingService
}

// NewGetExchangeRatesUseCase creates a new use case
func NewGetExchangeRatesUseCase(pricingService *exchange_services.PricingService) *GetExchangeRatesUseCase {
	return &GetExchangeRatesUseCase{pricingService: pricingService}
}

// Execute returns current exchange rates
func (uc *GetExchangeRatesUseCase) Execute(ctx context.Context) (*exchange_in.RatesResult, error) {
	pricing, err := uc.pricingService.GetCurrentPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	var sources []exchange_in.PriceSourceInfo
	for _, src := range pricing.Sources {
		sources = append(sources, exchange_in.PriceSourceInfo{
			Provider:  src.Provider,
			Price:     src.Price,
			Timestamp: src.Timestamp.Format(time.RFC3339),
		})
	}

	return &exchange_in.RatesResult{
		BTCUSD:      pricing.MedianPrice,
		LastUpdated: pricing.Timestamp.Format(time.RFC3339),
		Sources:     sources,
	}, nil
}
