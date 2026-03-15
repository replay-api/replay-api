package pricefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
)

// KrakenPriceAdapter implements PriceFeedProvider using Kraken public API
type KrakenPriceAdapter struct {
	httpClient *http.Client
	baseURL    string
}

// NewKrakenPriceAdapter creates a new Kraken price feed adapter
func NewKrakenPriceAdapter() *KrakenPriceAdapter {
	return &KrakenPriceAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.kraken.com/0/public",
	}
}

// krakenTickerResponse represents the Kraken API ticker response
type krakenTickerResponse struct {
	Error  []string                          `json:"error"`
	Result map[string]krakenTickerPairResult `json:"result"`
}

type krakenTickerPairResult struct {
	C []string `json:"c"` // Last trade closed [price, lot volume]
	V []string `json:"v"` // Volume [today, last 24 hours]
	B []string `json:"b"` // Bid [price, whole lot volume, lot volume]
	A []string `json:"a"` // Ask [price, whole lot volume, lot volume]
	H []string `json:"h"` // High [today, last 24 hours]
	L []string `json:"l"` // Low [today, last 24 hours]
}

// GetBTCUSDPrice returns the current BTC/USD price from Kraken
func (a *KrakenPriceAdapter) GetBTCUSDPrice(ctx context.Context) (*exchange_entities.PricePoint, error) {
	url := fmt.Sprintf("%s/Ticker?pair=XXBTZUSD", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kraken: failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kraken: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("kraken: API returned %d: %s", resp.StatusCode, string(body))
	}

	var result krakenTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("kraken: failed to decode response: %w", err)
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf("kraken: API error: %v", result.Error)
	}

	pair, ok := result.Result["XXBTZUSD"]
	if !ok {
		return nil, fmt.Errorf("kraken: XXBTZUSD pair not found in response")
	}

	if len(pair.C) == 0 {
		return nil, fmt.Errorf("kraken: no last trade price available")
	}

	price, err := strconv.ParseFloat(pair.C[0], 64)
	if err != nil {
		return nil, fmt.Errorf("kraken: failed to parse price '%s': %w", pair.C[0], err)
	}

	var volume float64
	if len(pair.V) > 1 {
		volume, _ = strconv.ParseFloat(pair.V[1], 64)
	}

	var bid, ask float64
	if len(pair.B) > 0 {
		bid, _ = strconv.ParseFloat(pair.B[0], 64)
	}
	if len(pair.A) > 0 {
		ask, _ = strconv.ParseFloat(pair.A[0], 64)
	}

	return &exchange_entities.PricePoint{
		Provider:  "kraken",
		Price:     price,
		Timestamp: time.Now().UTC(),
		Volume24h: volume,
		Bid:       bid,
		Ask:       ask,
	}, nil
}

// GetProvider returns the provider name
func (a *KrakenPriceAdapter) GetProvider() string {
	return "kraken"
}
