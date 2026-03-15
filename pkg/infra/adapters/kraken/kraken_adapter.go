package kraken

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
)

// KrakenAdapter implements ExchangeAdapter using Kraken REST API
type KrakenAdapter struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	apiSecret  string
}

// NewKrakenAdapter creates a new Kraken exchange adapter
func NewKrakenAdapter(apiKey, apiSecret string) *KrakenAdapter {
	return &KrakenAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.kraken.com",
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}
}

// GetProvider returns the exchange provider identifier
func (a *KrakenAdapter) GetProvider() exchange_vo.ExchangeProvider {
	return exchange_vo.ExchangeProviderKraken
}

// signRequest generates Kraken API signature
func (a *KrakenAdapter) signRequest(path string, nonce int64, data url.Values) string {
	sha := sha256.Sum256([]byte(fmt.Sprintf("%d%s", nonce, data.Encode())))
	decodedSecret, _ := base64.StdEncoding.DecodeString(a.apiSecret)
	mac := hmac.New(sha512.New, decodedSecret)
	mac.Write(append([]byte(path), sha[:]...))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// doPrivateRequest performs an authenticated request to Kraken
func (a *KrakenAdapter) doPrivateRequest(ctx context.Context, path string, data url.Values) ([]byte, error) {
	nonce := time.Now().UnixMicro()
	if data == nil {
		data = url.Values{}
	}
	data.Set("nonce", strconv.FormatInt(nonce, 10))

	signature := a.signRequest(path, nonce, data)

	reqURL := a.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("kraken: failed to create request: %w", err)
	}

	req.Header.Set("API-Key", a.apiKey)
	req.Header.Set("API-Sign", signature)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kraken: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kraken: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kraken: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// doPublicRequest performs a public (unauthenticated) request to Kraken
func (a *KrakenAdapter) doPublicRequest(ctx context.Context, path string) ([]byte, error) {
	reqURL := a.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("kraken: failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kraken: request failed: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// PlaceMarketBuyOrder places a market buy order on Kraken
func (a *KrakenAdapter) PlaceMarketBuyOrder(ctx context.Context, pair exchange_vo.ExchangePair, amountUSD float64) (*exchange_out.ExchangeOrderResult, error) {
	data := url.Values{}
	data.Set("pair", "XXBTZUSD")
	data.Set("type", "buy")
	data.Set("ordertype", "market")
	// Kraken uses "volume" for base currency amount; for market buy with USD, we use oflags=viqc
	data.Set("volume", fmt.Sprintf("%.2f", amountUSD))
	data.Set("oflags", "viqc") // Volume In Quote Currency

	respBody, err := a.doPrivateRequest(ctx, "/0/private/AddOrder", data)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Error  []string `json:"error"`
		Result struct {
			Descr struct {
				Order string `json:"order"`
			} `json:"descr"`
			Txid []string `json:"txid"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse order response: %w", err)
	}

	if len(resp.Error) > 0 {
		return nil, fmt.Errorf("kraken: order error: %v", resp.Error)
	}

	if len(resp.Result.Txid) == 0 {
		return nil, fmt.Errorf("kraken: no transaction ID returned")
	}

	txid := resp.Result.Txid[0]

	// Brief wait then check order status
	time.Sleep(2 * time.Second)
	return a.GetOrderStatus(ctx, txid)
}

// PlaceMarketSellOrder places a market sell order on Kraken
func (a *KrakenAdapter) PlaceMarketSellOrder(ctx context.Context, pair exchange_vo.ExchangePair, amountBTC float64) (*exchange_out.ExchangeOrderResult, error) {
	data := url.Values{}
	data.Set("pair", "XXBTZUSD")
	data.Set("type", "sell")
	data.Set("ordertype", "market")
	data.Set("volume", fmt.Sprintf("%.8f", amountBTC))

	respBody, err := a.doPrivateRequest(ctx, "/0/private/AddOrder", data)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Error  []string `json:"error"`
		Result struct {
			Txid []string `json:"txid"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse order response: %w", err)
	}

	if len(resp.Error) > 0 {
		return nil, fmt.Errorf("kraken: order error: %v", resp.Error)
	}

	if len(resp.Result.Txid) == 0 {
		return nil, fmt.Errorf("kraken: no transaction ID returned")
	}

	txid := resp.Result.Txid[0]
	time.Sleep(2 * time.Second)
	return a.GetOrderStatus(ctx, txid)
}

// GetOrderStatus checks the status of a Kraken order
func (a *KrakenAdapter) GetOrderStatus(ctx context.Context, orderID string) (*exchange_out.ExchangeOrderResult, error) {
	data := url.Values{}
	data.Set("txid", orderID)

	respBody, err := a.doPrivateRequest(ctx, "/0/private/QueryOrders", data)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Error  []string                   `json:"error"`
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse order status: %w", err)
	}

	if len(resp.Error) > 0 {
		return nil, fmt.Errorf("kraken: query error: %v", resp.Error)
	}

	orderData, ok := resp.Result[orderID]
	if !ok {
		return nil, fmt.Errorf("kraken: order %s not found", orderID)
	}

	var order struct {
		Status  string `json:"status"`
		Vol     string `json:"vol"`
		VolExec string `json:"vol_exec"`
		Cost    string `json:"cost"`
		Fee     string `json:"fee"`
		Price   string `json:"price"`
	}
	if err := json.Unmarshal(orderData, &order); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse order data: %w", err)
	}

	filledBTC, _ := strconv.ParseFloat(order.VolExec, 64)
	cost, _ := strconv.ParseFloat(order.Cost, 64)
	fee, _ := strconv.ParseFloat(order.Fee, 64)

	var avgPrice float64
	if filledBTC > 0 {
		avgPrice = cost / filledBTC
	}

	status := order.Status
	switch status {
	case "closed":
		status = "filled"
	case "canceled", "expired":
		status = "cancelled"
	case "open", "pending":
		status = "open"
	}

	return &exchange_out.ExchangeOrderResult{
		OrderID:      orderID,
		Status:       status,
		FilledQtyBTC: filledBTC,
		AvgPriceUSD:  avgPrice,
		FeeUSD:       fee,
		FeeCurrency:  "USD",
		Provider:     exchange_vo.ExchangeProviderKraken,
	}, nil
}

// GetTicker returns the current BTC/USD ticker from Kraken
func (a *KrakenAdapter) GetTicker(ctx context.Context, pair exchange_vo.ExchangePair) (*exchange_out.TickerResult, error) {
	respBody, err := a.doPublicRequest(ctx, "/0/public/Ticker?pair=XXBTZUSD")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Error  []string                   `json:"error"`
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse ticker: %w", err)
	}

	pairData, ok := resp.Result["XXBTZUSD"]
	if !ok {
		return nil, fmt.Errorf("kraken: XXBTZUSD pair not found")
	}

	var ticker struct {
		A []string `json:"a"` // Ask
		B []string `json:"b"` // Bid
		C []string `json:"c"` // Last trade
		V []string `json:"v"` // Volume
		H []string `json:"h"` // High
		L []string `json:"l"` // Low
	}
	if err := json.Unmarshal(pairData, &ticker); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse ticker data: %w", err)
	}

	ask, _ := strconv.ParseFloat(ticker.A[0], 64)
	bid, _ := strconv.ParseFloat(ticker.B[0], 64)
	last, _ := strconv.ParseFloat(ticker.C[0], 64)
	vol, _ := strconv.ParseFloat(ticker.V[1], 64)
	high, _ := strconv.ParseFloat(ticker.H[1], 64)
	low, _ := strconv.ParseFloat(ticker.L[1], 64)

	return &exchange_out.TickerResult{
		Pair:      pair,
		Bid:       bid,
		Ask:       ask,
		Last:      last,
		Volume24h: vol,
		High24h:   high,
		Low24h:    low,
		Provider:  exchange_vo.ExchangeProviderKraken,
	}, nil
}

// GetAccountBalance returns Kraken account balances
func (a *KrakenAdapter) GetAccountBalance(ctx context.Context) (map[string]float64, error) {
	respBody, err := a.doPrivateRequest(ctx, "/0/private/Balance", nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Error  []string          `json:"error"`
		Result map[string]string `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse balance: %w", err)
	}

	if len(resp.Error) > 0 {
		return nil, fmt.Errorf("kraken: balance error: %v", resp.Error)
	}

	balances := make(map[string]float64)
	for currency, amount := range resp.Result {
		val, _ := strconv.ParseFloat(amount, 64)
		if val > 0 {
			// Kraken uses XXBT for BTC
			if currency == "XXBT" {
				currency = "BTC"
			} else if currency == "ZUSD" {
				currency = "USD"
			}
			balances[currency] = val
		}
	}

	return balances, nil
}

// WithdrawBTC withdraws BTC from Kraken to an external address
func (a *KrakenAdapter) WithdrawBTC(ctx context.Context, address string, amountBTC float64) (*exchange_out.WithdrawResult, error) {
	data := url.Values{}
	data.Set("asset", "XBT")
	data.Set("key", address) // Kraken requires pre-registered withdrawal addresses by "key" name
	data.Set("amount", fmt.Sprintf("%.8f", amountBTC))

	respBody, err := a.doPrivateRequest(ctx, "/0/private/Withdraw", data)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Error  []string `json:"error"`
		Result struct {
			RefID string `json:"refid"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("kraken: failed to parse withdrawal response: %w", err)
	}

	if len(resp.Error) > 0 {
		return nil, fmt.Errorf("kraken: withdrawal error: %v", resp.Error)
	}

	return &exchange_out.WithdrawResult{
		WithdrawID: resp.Result.RefID,
		Status:     "pending",
		AmountBTC:  amountBTC,
	}, nil
}

// HealthCheck verifies the Kraken API connection
func (a *KrakenAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.doPublicRequest(ctx, "/0/public/Time")
	return err
}
