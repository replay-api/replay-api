package pricefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
)

// CoinGeckoAdapter implements PriceFeedProvider using CoinGecko API
type CoinGeckoAdapter struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string // Optional, for Pro API
}

// NewCoinGeckoAdapter creates a new CoinGecko price feed adapter
func NewCoinGeckoAdapter(apiKey string) *CoinGeckoAdapter {
	baseURL := "https://api.coingecko.com/api/v3"
	if apiKey != "" {
		baseURL = "https://pro-api.coingecko.com/api/v3"
	}
	return &CoinGeckoAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
	}
}

// GetBTCUSDPrice returns the current BTC/USD price from CoinGecko
func (a *CoinGeckoAdapter) GetBTCUSDPrice(ctx context.Context) (*exchange_entities.PricePoint, error) {
	url := fmt.Sprintf("%s/simple/price?ids=bitcoin&vs_currencies=usd&include_24hr_vol=true", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("coingecko: failed to create request: %w", err)
	}

	if a.apiKey != "" {
		req.Header.Set("x-cg-pro-api-key", a.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coingecko: API returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("coingecko: failed to decode response: %w", err)
	}

	btcData, ok := result["bitcoin"]
	if !ok {
		return nil, fmt.Errorf("coingecko: bitcoin data not found in response")
	}

	price, ok := btcData["usd"]
	if !ok {
		return nil, fmt.Errorf("coingecko: USD price not found in response")
	}

	volume := btcData["usd_24h_vol"]

	return &exchange_entities.PricePoint{
		Provider:  "coingecko",
		Price:     price,
		Timestamp: time.Now().UTC(),
		Volume24h: volume,
	}, nil
}

// GetProvider returns the provider name
func (a *CoinGeckoAdapter) GetProvider() string {
	return "coingecko"
}
