package exchange_services

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// PricingService aggregates prices from multiple providers and computes a reliable median
type PricingService struct {
	providers      []exchange_out.PriceFeedProvider
	rateCache      exchange_out.RateCache
	rateRepo       exchange_out.ExchangeRateRepository
	resourceOwner  shared.ResourceOwner
	cacheTTL       time.Duration
	minSources     int
	maxSpreadPct   float64
	mu             sync.RWMutex
	lastResult     *exchange_entities.PricingResult
	lastResultTime time.Time
}

// NewPricingService creates a new pricing service
func NewPricingService(
	providers []exchange_out.PriceFeedProvider,
	rateCache exchange_out.RateCache,
	rateRepo exchange_out.ExchangeRateRepository,
	resourceOwner shared.ResourceOwner,
) *PricingService {
	return &PricingService{
		providers:     providers,
		rateCache:     rateCache,
		rateRepo:      rateRepo,
		resourceOwner: resourceOwner,
		cacheTTL:      10 * time.Second,
		minSources:    2,
		maxSpreadPct:  2.0,
	}
}

// GetCurrentPrice returns the current BTC/USD price with confidence metrics
func (s *PricingService) GetCurrentPrice(ctx context.Context) (*exchange_entities.PricingResult, error) {
	s.mu.RLock()
	if s.lastResult != nil && time.Since(s.lastResultTime) < s.cacheTTL {
		result := s.lastResult
		s.mu.RUnlock()
		return result, nil
	}
	s.mu.RUnlock()

	if s.rateCache != nil {
		cached, err := s.rateCache.GetRate(ctx, exchange_vo.PairBTCUSD)
		if err == nil && cached != nil && !cached.IsStale(s.cacheTTL) {
			result := &exchange_entities.PricingResult{
				MedianPrice: cached.MedianPrice,
				Sources:     cached.Sources,
				Timestamp:   cached.Timestamp,
				Confidence:  cached.Confidence,
				Spread:      cached.Spread,
				RateID:      cached.ID,
			}
			s.updateInMemoryCache(result)
			return result, nil
		}
	}

	return s.fetchAndAggregate(ctx)
}

func (s *PricingService) fetchAndAggregate(ctx context.Context) (*exchange_entities.PricingResult, error) {
	type providerResult struct {
		point *exchange_entities.PricePoint
		err   error
	}

	results := make(chan providerResult, len(s.providers))
	for _, provider := range s.providers {
		go func(p exchange_out.PriceFeedProvider) {
			point, err := p.GetBTCUSDPrice(ctx)
			results <- providerResult{point: point, err: err}
		}(provider)
	}

	var points []exchange_entities.PricePoint
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(s.providers); i++ {
		select {
		case r := <-results:
			if r.err != nil {
				log.Printf("[PricingService] Provider error: %v", r.err)
				continue
			}
			if r.point != nil {
				points = append(points, *r.point)
			}
		case <-timeout:
			log.Printf("[PricingService] Timeout waiting for providers, got %d/%d", len(points), len(s.providers))
		}
	}

	if len(points) < s.minSources {
		return nil, fmt.Errorf("insufficient price sources: got %d, need %d", len(points), s.minSources)
	}

	prices := make([]float64, len(points))
	for i, p := range points {
		prices[i] = p.Price
	}
	sort.Float64s(prices)

	medianPrice := computeMedian(prices)

	var maxDeviation float64
	for _, price := range prices {
		deviation := math.Abs(price-medianPrice) / medianPrice * 100.0
		if deviation > maxDeviation {
			maxDeviation = deviation
		}
	}

	if maxDeviation > s.maxSpreadPct {
		return nil, fmt.Errorf("price spread too wide (%.2f%% > %.2f%%): possible price manipulation", maxDeviation, s.maxSpreadPct)
	}

	confidence := 1.0 - (maxDeviation / s.maxSpreadPct)
	if confidence < 0 {
		confidence = 0
	}

	result := &exchange_entities.PricingResult{
		MedianPrice: medianPrice,
		Sources:     points,
		Timestamp:   time.Now().UTC(),
		Confidence:  confidence,
		Spread:      maxDeviation,
	}

	s.persistRate(ctx, result)
	s.updateInMemoryCache(result)

	return result, nil
}

func (s *PricingService) persistRate(ctx context.Context, result *exchange_entities.PricingResult) {
	if s.rateRepo != nil {
		rate := exchange_entities.NewExchangeRate(
			s.resourceOwner,
			exchange_vo.PairBTCUSD,
			result.MedianPrice,
			result.Sources,
			result.Confidence,
			result.Spread,
		)
		if err := s.rateRepo.Save(ctx, rate); err != nil {
			log.Printf("[PricingService] Failed to persist rate: %v", err)
		}
		result.RateID = rate.ID

		if s.rateCache != nil {
			if err := s.rateCache.SetRate(ctx, exchange_vo.PairBTCUSD, rate); err != nil {
				log.Printf("[PricingService] Failed to cache rate: %v", err)
			}
		}
	}
}

func (s *PricingService) updateInMemoryCache(result *exchange_entities.PricingResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastResult = result
	s.lastResultTime = time.Now()
}

func computeMedian(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2.0
	}
	return sorted[n/2]
}
