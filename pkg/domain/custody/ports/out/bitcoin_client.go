package custody_out

import (
	"context"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID          string `json:"txid"`
	Vout          uint32 `json:"vout"`
	AmountSats    int64  `json:"amount_sats"`
	ScriptPubKey  string `json:"script_pubkey"`
	Confirmations int64  `json:"confirmations"`
}

// BuildBTCTxRequest contains the parameters for building a Bitcoin transaction
type BuildBTCTxRequest struct {
	FromAddress        string `json:"from_address"`
	ToAddress          string `json:"to_address"`
	AmountSats         int64  `json:"amount_sats"`
	FeeRateSatPerVByte int64  `json:"fee_rate_sat_per_vbyte"`
	UTXOs              []UTXO `json:"utxos,omitempty"` // If empty, client will fetch UTXOs
}

// UnsignedBTCTx represents a Bitcoin transaction ready for signing
type UnsignedBTCTx struct {
	RawTx        []byte   `json:"raw_tx"`
	TxHash       string   `json:"tx_hash"`
	InputHashes  [][]byte `json:"input_hashes"` // Hashes to sign for each input
	EstimatedFee int64    `json:"estimated_fee_sats"`
}

// BTCTransaction represents an on-chain Bitcoin transaction
type BTCTransaction struct {
	TxHash        string `json:"tx_hash"`
	BlockHash     string `json:"block_hash,omitempty"`
	BlockHeight   int64  `json:"block_height,omitempty"`
	Confirmations int64  `json:"confirmations"`
	AmountSats    int64  `json:"amount_sats"`
	FeeSats       int64  `json:"fee_sats"`
	Status        string `json:"status"` // "pending", "confirmed", "failed"
	Timestamp     int64  `json:"timestamp"`
}

// BTCAddressType represents the type of Bitcoin address
type BTCAddressType string

const (
	BTCAddrP2PKH   BTCAddressType = "P2PKH"
	BTCAddrP2SH    BTCAddressType = "P2SH"
	BTCAddrBech32  BTCAddressType = "Bech32"
	BTCAddrTaproot BTCAddressType = "Taproot"
)

// BitcoinClient provides Bitcoin blockchain operations
type BitcoinClient interface {
	// GetBalance returns the balance of an address in satoshis
	GetBalance(ctx context.Context, address string) (int64, error)

	// GetUTXOs returns unspent transaction outputs for an address
	GetUTXOs(ctx context.Context, address string) ([]UTXO, error)

	// BuildTransaction creates an unsigned transaction
	BuildTransaction(ctx context.Context, req BuildBTCTxRequest) (*UnsignedBTCTx, error)

	// EstimateFee returns the recommended fee rate in sat/vByte for target confirmation blocks
	EstimateFee(ctx context.Context, targetBlocks int) (int64, error)

	// BroadcastTransaction broadcasts a signed transaction to the network
	BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error)

	// GetTransaction returns details of a transaction
	GetTransaction(ctx context.Context, txHash string) (*BTCTransaction, error)

	// WaitForConfirmations waits until a transaction has the specified number of confirmations
	WaitForConfirmations(ctx context.Context, txHash string, confirmations int) error

	// ValidateAddress validates a Bitcoin address and returns its type
	ValidateAddress(ctx context.Context, address string) (bool, BTCAddressType, error)

	// HealthCheck verifies the Bitcoin node/API connection
	HealthCheck(ctx context.Context) error
}
