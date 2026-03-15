package coinbase

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
)

// CoinbaseAdapter implements ExchangeAdapter using Coinbase Advanced Trade API
type CoinbaseAdapter struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	apiSecret  string
}

// NewCoinbaseAdapter creates a new Coinbase exchange adapter
func NewCoinbaseAdapter(apiKey, apiSecret string) *CoinbaseAdapter {
	return &CoinbaseAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.coinbase.com/api/v3/brokerage",
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}
}

// GetProvider returns the exchange provider identifier
func (a *CoinbaseAdapter) GetProvider() exchange_vo.ExchangeProvider {
	return exchange_vo.ExchangeProviderCoinbase
}

// signRequest generates HMAC-SHA256 signature for Coinbase API
func (a *CoinbaseAdapter) signRequest(timestamp, method, path, body string) string {
	message := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(a.apiSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// doRequest performs an authenticated request to the Coinbase API
func (a *CoinbaseAdapter) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("coinbase: failed to marshal request: %w", err)
		}
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := a.signRequest(timestamp, method, path, string(bodyBytes))

	url := a.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("coinbase: failed to create request: %w", err)
	}

	req.Header.Set("CB-ACCESS-KEY", a.apiKey)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coinbase: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("coinbase: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("coinbase: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// PlaceMarketBuyOrder places a market buy order on Coinbase
func (a *CoinbaseAdapter) PlaceMarketBuyOrder(ctx context.Context, pair exchange_vo.ExchangePair, amountUSD float64) (*exchange_out.ExchangeOrderResult, error) {
	orderReq := map[string]interface{}{
		"client_order_id": fmt.Sprintf("leet-%d", time.Now().UnixNano()),
		"product_id":      "BTC-USD",
		"side":            "BUY",
		"order_configuration": map[string]interface{}{
			"market_market_ioc": map[string]interface{}{
				"quote_size": fmt.Sprintf("%.2f", amountUSD),
			},
		},
	}

	respBody, err := a.doRequest(ctx, "POST", "/orders", orderReq)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Success       bool   `json:"success"`
		OrderID       string `json:"order_id"`
		FailureReason string `json:"failure_reason"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse order response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("coinbase: order failed: %s", resp.FailureReason)
	}

	// Poll for fill (market orders usually fill immediately)
	return a.GetOrderStatus(ctx, resp.OrderID)
}

// PlaceMarketSellOrder places a market sell order on Coinbase
func (a *CoinbaseAdapter) PlaceMarketSellOrder(ctx context.Context, pair exchange_vo.ExchangePair, amountBTC float64) (*exchange_out.ExchangeOrderResult, error) {
	orderReq := map[string]interface{}{
		"client_order_id": fmt.Sprintf("leet-%d", time.Now().UnixNano()),
		"product_id":      "BTC-USD",
		"side":            "SELL",
		"order_configuration": map[string]interface{}{
			"market_market_ioc": map[string]interface{}{
				"base_size": fmt.Sprintf("%.8f", amountBTC),
			},
		},
	}

	respBody, err := a.doRequest(ctx, "POST", "/orders", orderReq)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Success       bool   `json:"success"`
		OrderID       string `json:"order_id"`
		FailureReason string `json:"failure_reason"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse order response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("coinbase: order failed: %s", resp.FailureReason)
	}

	return a.GetOrderStatus(ctx, resp.OrderID)
}

// GetOrderStatus checks the status of a Coinbase order
func (a *CoinbaseAdapter) GetOrderStatus(ctx context.Context, orderID string) (*exchange_out.ExchangeOrderResult, error) {
	path := fmt.Sprintf("/orders/historical/%s", orderID)
	respBody, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Order struct {
			OrderID            string `json:"order_id"`
			Status             string `json:"status"`
			FilledSize         string `json:"filled_size"`
			AverageFilledPrice string `json:"average_filled_price"`
			TotalFees          string `json:"total_fees"`
		} `json:"order"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse order status: %w", err)
	}

	filledBTC, _ := strconv.ParseFloat(resp.Order.FilledSize, 64)
	avgPrice, _ := strconv.ParseFloat(resp.Order.AverageFilledPrice, 64)
	fees, _ := strconv.ParseFloat(resp.Order.TotalFees, 64)

	// Map Coinbase status to our status
	status := resp.Order.Status
	switch status {
	case "FILLED":
		status = "filled"
	case "CANCELLED":
		status = "cancelled"
	case "PENDING", "OPEN":
		status = "open"
	case "FAILED":
		status = "failed"
	}

	return &exchange_out.ExchangeOrderResult{
		OrderID:      resp.Order.OrderID,
		Status:       status,
		FilledQtyBTC: filledBTC,
		AvgPriceUSD:  avgPrice,
		FeeUSD:       fees,
		FeeCurrency:  "USD",
		Provider:     exchange_vo.ExchangeProviderCoinbase,
	}, nil
}

// GetTicker returns the current BTC-USD ticker from Coinbase
func (a *CoinbaseAdapter) GetTicker(ctx context.Context, pair exchange_vo.ExchangePair) (*exchange_out.TickerResult, error) {
	respBody, err := a.doRequest(ctx, "GET", "/products/BTC-USD", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		ProductID string `json:"product_id"`
		Price     string `json:"price"`
		Bid       string `json:"bid"`
		Ask       string `json:"ask"`
		Volume24h string `json:"volume_24h"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse ticker: %w", err)
	}

	last, _ := strconv.ParseFloat(resp.Price, 64)
	bid, _ := strconv.ParseFloat(resp.Bid, 64)
	ask, _ := strconv.ParseFloat(resp.Ask, 64)
	vol, _ := strconv.ParseFloat(resp.Volume24h, 64)

	return &exchange_out.TickerResult{
		Pair:      pair,
		Bid:       bid,
		Ask:       ask,
		Last:      last,
		Volume24h: vol,
		Provider:  exchange_vo.ExchangeProviderCoinbase,
	}, nil
}

// GetAccountBalance returns Coinbase account balances
func (a *CoinbaseAdapter) GetAccountBalance(ctx context.Context) (map[string]float64, error) {
	respBody, err := a.doRequest(ctx, "GET", "/accounts", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Accounts []struct {
			Currency         string `json:"currency"`
			AvailableBalance struct {
				Value string `json:"value"`
			} `json:"available_balance"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse accounts: %w", err)
	}

	balances := make(map[string]float64)
	for _, acc := range resp.Accounts {
		val, _ := strconv.ParseFloat(acc.AvailableBalance.Value, 64)
		if val > 0 {
			balances[acc.Currency] = val
		}
	}

	return balances, nil
}

// WithdrawBTC withdraws BTC from Coinbase to an external address
func (a *CoinbaseAdapter) WithdrawBTC(ctx context.Context, address string, amountBTC float64) (*exchange_out.WithdrawResult, error) {
	withdrawReq := map[string]interface{}{
		"amount":   fmt.Sprintf("%.8f", amountBTC),
		"currency": "BTC",
		"crypto_address": map[string]interface{}{
			"address": address,
		},
	}

	// Use Coinbase Send endpoint (v2 API for withdrawals)
	// Note: This uses a different auth path for the main Coinbase API
	respBody, err := a.doRequest(ctx, "POST", "/withdrawals/crypto", withdrawReq)
	if err != nil {
		return nil, err
	}

	var resp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Amount string `json:"amount"`
		} `json:"amount"`
		Fee struct {
			Amount string `json:"amount"`
		} `json:"fee"`
		Network struct {
			Hash string `json:"hash"`
		} `json:"network"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("coinbase: failed to parse withdrawal response: %w", err)
	}

	feeBTC, _ := strconv.ParseFloat(resp.Fee.Amount, 64)

	return &exchange_out.WithdrawResult{
		WithdrawID: resp.ID,
		Status:     resp.Status,
		AmountBTC:  amountBTC,
		FeeBTC:     feeBTC,
		TxHash:     resp.Network.Hash,
	}, nil
}

// HealthCheck verifies the Coinbase API connection
func (a *CoinbaseAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.doRequest(ctx, "GET", "/products/BTC-USD", nil)
	return err
}
