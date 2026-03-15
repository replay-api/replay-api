package exchange_services

import (
	"context"
	"fmt"
	"log"

	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
)

// SmartRouter routes orders to the best available exchange based on price and availability
type SmartRouter struct {
	exchanges []exchange_out.ExchangeAdapter
}

// NewSmartRouter creates a new smart router with the given exchange adapters
func NewSmartRouter(exchanges []exchange_out.ExchangeAdapter) *SmartRouter {
	return &SmartRouter{exchanges: exchanges}
}

// RouteOrder selects the best exchange for an order based on current ticker prices
// For buys: picks the exchange with the lowest ask price
// For sells: picks the exchange with the highest bid price
func (r *SmartRouter) RouteOrder(ctx context.Context, side exchange_vo.OrderSide, pair exchange_vo.ExchangePair) (exchange_out.ExchangeAdapter, error) {
	if len(r.exchanges) == 0 {
		return nil, fmt.Errorf("no exchanges configured")
	}

	if len(r.exchanges) == 1 {
		if err := r.exchanges[0].HealthCheck(ctx); err != nil {
			return nil, fmt.Errorf("only exchange %s is unhealthy: %w", r.exchanges[0].GetProvider(), err)
		}
		return r.exchanges[0], nil
	}

	type exchangeQuote struct {
		adapter exchange_out.ExchangeAdapter
		ticker  *exchange_out.TickerResult
		err     error
	}

	results := make(chan exchangeQuote, len(r.exchanges))
	for _, ex := range r.exchanges {
		go func(adapter exchange_out.ExchangeAdapter) {
			ticker, err := adapter.GetTicker(ctx, pair)
			results <- exchangeQuote{adapter: adapter, ticker: ticker, err: err}
		}(ex)
	}

	var bestAdapter exchange_out.ExchangeAdapter
	var bestPrice float64
	for i := 0; i < len(r.exchanges); i++ {
		result := <-results
		if result.err != nil {
			log.Printf("[SmartRouter] Exchange %s error: %v", result.adapter.GetProvider(), result.err)
			continue
		}

		var price float64
		if side == exchange_vo.OrderSideBuy {
			price = result.ticker.Ask
		} else {
			price = result.ticker.Bid
		}

		if bestAdapter == nil {
			bestAdapter = result.adapter
			bestPrice = price
			continue
		}

		if side == exchange_vo.OrderSideBuy && price < bestPrice {
			bestAdapter = result.adapter
			bestPrice = price
		} else if side == exchange_vo.OrderSideSell && price > bestPrice {
			bestAdapter = result.adapter
			bestPrice = price
		}
	}

	if bestAdapter == nil {
		return nil, fmt.Errorf("no healthy exchanges available")
	}

	log.Printf("[SmartRouter] Selected %s for %s order (price: $%.2f)", bestAdapter.GetProvider(), side, bestPrice)
	return bestAdapter, nil
}

// GetAllTickers returns tickers from all exchanges for comparison
func (r *SmartRouter) GetAllTickers(ctx context.Context, pair exchange_vo.ExchangePair) []*exchange_out.TickerResult {
	var tickers []*exchange_out.TickerResult
	for _, ex := range r.exchanges {
		ticker, err := ex.GetTicker(ctx, pair)
		if err != nil {
			log.Printf("[SmartRouter] Ticker error from %s: %v", ex.GetProvider(), err)
			continue
		}
		tickers = append(tickers, ticker)
	}
	return tickers
}