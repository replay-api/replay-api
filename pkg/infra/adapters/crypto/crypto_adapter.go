// Package crypto provides the cryptocurrency payment provider adapter implementation
// Supports EVM-compatible chains (Ethereum, Polygon, etc.) for USDC/USDT payments
// Follows SOLID principles - implements the same PaymentProviderAdapter interface
package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// Supported chains enum
type ChainID int64

const (
	ChainEthereum ChainID = 1
	ChainPolygon  ChainID = 137
	ChainArbitrum ChainID = 42161
	ChainBase     ChainID = 8453
)

// Token addresses on different chains (USDC)
var USDCAddresses = map[ChainID]string{
	ChainEthereum: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
	ChainPolygon:  "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
	ChainArbitrum: "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
	ChainBase:     "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
}

// USDT addresses
var USDTAddresses = map[ChainID]string{
	ChainEthereum: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
	ChainPolygon:  "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
	ChainArbitrum: "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9",
}

// ERC20 Transfer event signature
const ERC20TransferEventSig = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// CryptoAdapter implements the PaymentProviderAdapter interface for crypto payments
type CryptoAdapter struct {
	depositAddress string              // Platform deposit address
	rpcEndpoints   map[ChainID]string  // RPC endpoints for each chain
	httpClient     *http.Client
	defaultChain   ChainID
}

// NewCryptoAdapter creates a new crypto adapter
func NewCryptoAdapter() *CryptoAdapter {
	depositAddress := os.Getenv("CRYPTO_DEPOSIT_ADDRESS")
	if depositAddress == "" {
		depositAddress = "0x0000000000000000000000000000000000000000" // Placeholder
	}

	// Initialize RPC endpoints from environment
	rpcEndpoints := make(map[ChainID]string)
	rpcEndpoints[ChainEthereum] = os.Getenv("ETH_RPC_URL")
	rpcEndpoints[ChainPolygon] = os.Getenv("POLYGON_RPC_URL")
	rpcEndpoints[ChainArbitrum] = os.Getenv("ARBITRUM_RPC_URL")
	rpcEndpoints[ChainBase] = os.Getenv("BASE_RPC_URL")

	// Set defaults if not configured
	if rpcEndpoints[ChainEthereum] == "" {
		rpcEndpoints[ChainEthereum] = "https://eth.llamarpc.com"
	}
	if rpcEndpoints[ChainPolygon] == "" {
		rpcEndpoints[ChainPolygon] = "https://polygon.llamarpc.com"
	}

	return &CryptoAdapter{
		depositAddress: depositAddress,
		rpcEndpoints:   rpcEndpoints,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		defaultChain: ChainPolygon, // Default to Polygon for lower fees
	}
}

// GetProvider returns the provider type
func (c *CryptoAdapter) GetProvider() payment_entities.PaymentProvider {
	return payment_entities.PaymentProviderCrypto
}

// CreatePaymentIntent creates a crypto payment intent
// Returns the deposit address and expected amount
func (c *CryptoAdapter) CreatePaymentIntent(ctx context.Context, req payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error) {
	// Generate unique payment ID for tracking
	paymentID := "crypto_" + uuid.New().String()

	// For crypto, we return the deposit address
	// The frontend will prompt user to send to this address
	// We track by amount and memo/reference

	return &payment_out.CreateIntentResponse{
		ProviderPaymentID: paymentID,
		CryptoAddress:     c.depositAddress,
		Status:            "awaiting_payment",
	}, nil
}

// ConfirmPayment verifies a crypto payment on-chain
func (c *CryptoAdapter) ConfirmPayment(ctx context.Context, req payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error) {
	// The PaymentMethodID for crypto is the transaction hash
	txHash := req.PaymentMethodID
	if txHash == "" {
		return nil, fmt.Errorf("transaction hash is required for crypto confirmation")
	}

	// Verify transaction on-chain
	verified, err := c.verifyTransaction(ctx, txHash, c.defaultChain)
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	if !verified {
		return &payment_out.ConfirmPaymentResponse{
			Status:      "failed",
			ProviderFee: 0, // No fee on failure
		}, nil
	}

	return &payment_out.ConfirmPaymentResponse{
		Status:      "succeeded",
		ProviderFee: 0, // No provider fee for crypto (gas paid by user)
	}, nil
}

// verifyTransaction verifies a transaction on-chain
func (c *CryptoAdapter) verifyTransaction(ctx context.Context, txHash string, chainID ChainID) (bool, error) {
	rpcURL, ok := c.rpcEndpoints[chainID]
	if !ok || rpcURL == "" {
		return false, fmt.Errorf("no RPC endpoint configured for chain %d", chainID)
	}

	// Get transaction receipt
	receipt, err := c.getTransactionReceipt(ctx, rpcURL, txHash)
	if err != nil {
		return false, err
	}

	// Check if transaction succeeded (status = 1)
	if receipt.Status != "0x1" {
		return false, fmt.Errorf("transaction failed on-chain")
	}

	// Verify the transfer was to our deposit address
	depositAddressLower := strings.ToLower(c.depositAddress)

	for _, log := range receipt.Logs {
		// Check if this is an ERC20 Transfer event
		if len(log.Topics) >= 3 && log.Topics[0] == ERC20TransferEventSig {
			// Third topic is the "to" address (padded to 32 bytes)
			toAddress := "0x" + log.Topics[2][26:] // Remove padding
			if strings.ToLower(toAddress) == depositAddressLower {
				return true, nil
			}
		}
	}

	return false, fmt.Errorf("no transfer to deposit address found in transaction")
}

// TransactionReceipt represents an Ethereum transaction receipt
type TransactionReceipt struct {
	Status string `json:"status"`
	Logs   []struct {
		Address string   `json:"address"`
		Topics  []string `json:"topics"`
		Data    string   `json:"data"`
	} `json:"logs"`
	GasUsed string `json:"gasUsed"`
}

// getTransactionReceipt fetches transaction receipt from RPC
func (c *CryptoAdapter) getTransactionReceipt(ctx context.Context, rpcURL, txHash string) (*TransactionReceipt, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []string{txHash},
		"id":      1,
	}

	body, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", rpcURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Result *TransactionReceipt `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	if rpcResp.Result == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	return rpcResp.Result, nil
}

// RefundPayment initiates a crypto refund (manual process)
func (c *CryptoAdapter) RefundPayment(ctx context.Context, req payment_out.RefundRequest) (*payment_out.RefundResponse, error) {
	// Crypto refunds typically require manual processing
	// We create a refund record that will be processed by treasury

	refundID := "refund_" + uuid.New().String()

	return &payment_out.RefundResponse{
		RefundID: refundID,
		Status:   "pending_manual_review",
		Amount:   req.Amount,
	}, nil
}

// CancelPayment cancels a crypto payment intent (no-op if not paid)
func (c *CryptoAdapter) CancelPayment(ctx context.Context, req payment_out.CancelRequest) (*payment_out.CancelResponse, error) {
	// For crypto, we just mark as canceled
	// If user already sent funds, they need to request refund separately

	return &payment_out.CancelResponse{
		Status: "canceled",
	}, nil
}

// ParseWebhook parses blockchain events (typically from indexer)
func (c *CryptoAdapter) ParseWebhook(payload []byte, signature string) (*payment_out.WebhookEvent, error) {
	// Crypto webhooks typically come from blockchain indexers like:
	// - Alchemy webhooks
	// - QuickNode streams
	// - Custom indexer

	var event struct {
		EventType     string `json:"event_type"`
		TransactionHash string `json:"transaction_hash"`
		ToAddress     string `json:"to_address"`
		Amount        string `json:"amount"`
		Token         string `json:"token"`
		ChainID       int64  `json:"chain_id"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Verify the transfer is to our deposit address
	if strings.ToLower(event.ToAddress) != strings.ToLower(c.depositAddress) {
		return nil, fmt.Errorf("transfer not to our deposit address")
	}

	webhookEvent := &payment_out.WebhookEvent{
		EventType:         event.EventType,
		ProviderPaymentID: event.TransactionHash,
		Status:            payment_entities.PaymentStatusSucceeded,
		Metadata: map[string]any{
			"chain_id": event.ChainID,
			"token":    event.Token,
			"amount":   event.Amount,
		},
	}

	return webhookEvent, nil
}

// CreateOrGetCustomer is not applicable for crypto
// Users are identified by their wallet address
func (c *CryptoAdapter) CreateOrGetCustomer(ctx context.Context, req payment_out.CustomerRequest) (*payment_out.CustomerResponse, error) {
	// For crypto, the "customer" is their wallet address
	// Return metadata wallet address if available
	if walletAddr, ok := req.Metadata["wallet_address"].(string); ok {
		return &payment_out.CustomerResponse{
			CustomerID: walletAddr,
		}, nil
	}

	return &payment_out.CustomerResponse{
		CustomerID: req.Email, // Fallback to email
	}, nil
}

// GetDepositAddress returns the platform deposit address for a specific chain
func (c *CryptoAdapter) GetDepositAddress(chainID ChainID) string {
	return c.depositAddress
}

// GetSupportedTokens returns supported stablecoin addresses for a chain
func (c *CryptoAdapter) GetSupportedTokens(chainID ChainID) map[string]string {
	tokens := make(map[string]string)

	if addr, ok := USDCAddresses[chainID]; ok {
		tokens["USDC"] = addr
	}
	if addr, ok := USDTAddresses[chainID]; ok {
		tokens["USDT"] = addr
	}

	return tokens
}

// ParseTransferAmount parses the amount from ERC20 transfer data
func ParseTransferAmount(data string) (*big.Int, error) {
	// Data is 32 bytes (64 hex chars) amount
	if len(data) < 2 || !strings.HasPrefix(data, "0x") {
		return nil, fmt.Errorf("invalid data format")
	}

	amountHex := data[2:] // Remove 0x prefix

	amount := new(big.Int)
	_, ok := amount.SetString(amountHex, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse amount")
	}

	return amount, nil
}

// FormatAmount formats a big.Int amount to human readable format
func FormatAmount(amount *big.Int, decimals int) string {
	// USDC/USDT have 6 decimals
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	whole := new(big.Int).Div(amount, divisor)
	remainder := new(big.Int).Mod(amount, divisor)

	return fmt.Sprintf("%s.%0*d", whole.String(), decimals, remainder)
}

// Ensure CryptoAdapter implements PaymentProviderAdapter
var _ payment_out.PaymentProviderAdapter = (*CryptoAdapter)(nil)
