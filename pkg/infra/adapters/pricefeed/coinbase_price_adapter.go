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

// CoinbasePriceAdapter implements PriceFeedProvider using Coinbase public API
type CoinbasePriceAdapter struct {
	httpClient *http.Client
	baseURL    string
}

// NewCoinbasePriceAdapter creates a new Coinbase price feed adapter
func NewCoinbasePriceAdapter() *CoinbasePriceAdapter {
	return &CoinbasePriceAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.coinbase.com/v2",
	}
}

// coinbaseSpotResponse represents the Coinbase API spot price response
type coinbaseSpotResponse struct {
	Data struct {
		Base     string `json:"base"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	} `json:"data"`
}

// GetBTCUSDPrice returns the current BTC/USD price from Coinbase
func (a *CoinbasePriceAdapter) GetBTCUSDPrice(ctx context.Context) (*exchange_entities.PricePoint, error) {
	url := fmt.Sprintf("%s/prices/BTC-USD/spot", a.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("coinbase: failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coinbase: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coinbase: API returned %d: %s", resp.StatusCode, string(body))
	}

	var result coinbaseSpotResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("coinbase: failed to decode response: %w", err)
	}

	price, err := strconv.ParseFloat(result.Data.Amount, 64)
	if err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse price '%s': %w", result.Data.Amount, err)
	}

	return &exchange_entities.PricePoint{
		Provider:  "coinbase",
		Price:     price,
		Timestamp: time.Now().UTC(),
	}, nil
}

// GetProvider returns the provider name
func (a *CoinbasePriceAdapter) GetProvider() string {
	return "coinbase"
}
