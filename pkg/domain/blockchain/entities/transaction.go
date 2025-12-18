package blockchain_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// TransactionStatus represents the on-chain transaction status
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "Pending"
	TxStatusConfirmed TransactionStatus = "Confirmed"
	TxStatusFailed    TransactionStatus = "Failed"
	TxStatusReplaced  TransactionStatus = "Replaced"
)

// TransactionType categorizes blockchain transactions
type TransactionType string

const (
	TxTypeDeposit      TransactionType = "Deposit"
	TxTypeWithdrawal   TransactionType = "Withdrawal"
	TxTypeEntryFee     TransactionType = "EntryFee"
	TxTypePrize        TransactionType = "Prize"
	TxTypeRefund       TransactionType = "Refund"
	TxTypePlatformFee  TransactionType = "PlatformFee"
	TxTypeBridge       TransactionType = "Bridge"
	TxTypeContractCall TransactionType = "ContractCall"
)

// BlockchainTransaction represents a transaction on the blockchain
type BlockchainTransaction struct {
	common.BaseEntity

	// Chain info
	ChainID     blockchain_vo.ChainID `json:"chain_id" bson:"chain_id"`
	TxHash      blockchain_vo.TxHash  `json:"tx_hash" bson:"tx_hash"`
	BlockNumber uint64                `json:"block_number" bson:"block_number"`
	BlockHash   blockchain_vo.TxHash  `json:"block_hash" bson:"block_hash"`

	// Transaction details
	Type        TransactionType   `json:"type" bson:"type"`
	Status      TransactionStatus `json:"status" bson:"status"`
	FromAddress wallet_vo.EVMAddress `json:"from_address" bson:"from_address"`
	ToAddress   wallet_vo.EVMAddress `json:"to_address" bson:"to_address"`
	TokenAddress *wallet_vo.EVMAddress `json:"token_address,omitempty" bson:"token_address,omitempty"`

	// Value
	Currency wallet_vo.Currency `json:"currency" bson:"currency"`
	Amount   wallet_vo.Amount   `json:"amount" bson:"amount"`

	// Gas
	GasLimit uint64           `json:"gas_limit" bson:"gas_limit"`
	GasUsed  uint64           `json:"gas_used" bson:"gas_used"`
	GasPrice wallet_vo.Amount `json:"gas_price" bson:"gas_price"`

	// Related entities
	WalletID     *uuid.UUID `json:"wallet_id,omitempty" bson:"wallet_id,omitempty"`
	MatchID      *uuid.UUID `json:"match_id,omitempty" bson:"match_id,omitempty"`
	TournamentID *uuid.UUID `json:"tournament_id,omitempty" bson:"tournament_id,omitempty"`
	LedgerTxID   *uuid.UUID `json:"ledger_tx_id,omitempty" bson:"ledger_tx_id,omitempty"`

	// Timestamps
	SubmittedAt *time.Time `json:"submitted_at,omitempty" bson:"submitted_at,omitempty"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`

	// Confirmations
	Confirmations      uint64 `json:"confirmations" bson:"confirmations"`
	RequiredConfirms   uint64 `json:"required_confirms" bson:"required_confirms"`

	// Error handling
	ErrorMessage string `json:"error_message,omitempty" bson:"error_message,omitempty"`
	RetryCount   int    `json:"retry_count" bson:"retry_count"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// NewBlockchainTransaction creates a new blockchain transaction
func NewBlockchainTransaction(
	resourceOwner common.ResourceOwner,
	chainID blockchain_vo.ChainID,
	txType TransactionType,
	from wallet_vo.EVMAddress,
	to wallet_vo.EVMAddress,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
) *BlockchainTransaction {
	baseEntity := common.NewPrivateEntity(resourceOwner)
	now := time.Now()

	return &BlockchainTransaction{
		BaseEntity:        baseEntity,
		ChainID:           chainID,
		Type:              txType,
		Status:            TxStatusPending,
		FromAddress:       from,
		ToAddress:         to,
		Currency:          currency,
		Amount:            amount,
		SubmittedAt:       &now,
		RequiredConfirms:  12, // Default for most chains
		Metadata:          make(map[string]interface{}),
	}
}

// SetTxHash sets the transaction hash after submission
func (t *BlockchainTransaction) SetTxHash(hash blockchain_vo.TxHash) {
	t.TxHash = hash
	t.UpdatedAt = time.Now()
}

// Confirm marks the transaction as confirmed
func (t *BlockchainTransaction) Confirm(blockNumber uint64, blockHash blockchain_vo.TxHash, gasUsed uint64) {
	now := time.Now()
	t.Status = TxStatusConfirmed
	t.BlockNumber = blockNumber
	t.BlockHash = blockHash
	t.GasUsed = gasUsed
	t.ConfirmedAt = &now
	t.UpdatedAt = now
}

// Fail marks the transaction as failed
func (t *BlockchainTransaction) Fail(errorMsg string) {
	t.Status = TxStatusFailed
	t.ErrorMessage = errorMsg
	t.UpdatedAt = time.Now()
}

// IncrementRetry increments the retry count
func (t *BlockchainTransaction) IncrementRetry() {
	t.RetryCount++
	t.UpdatedAt = time.Now()
}

// UpdateConfirmations updates the confirmation count
func (t *BlockchainTransaction) UpdateConfirmations(confirmations uint64) {
	t.Confirmations = confirmations
	t.UpdatedAt = time.Now()
}

// IsConfirmed checks if transaction has enough confirmations
func (t *BlockchainTransaction) IsConfirmed() bool {
	return t.Status == TxStatusConfirmed && t.Confirmations >= t.RequiredConfirms
}

// IsPending checks if transaction is still pending
func (t *BlockchainTransaction) IsPending() bool {
	return t.Status == TxStatusPending
}

// SetRelatedMatch sets the related match ID
func (t *BlockchainTransaction) SetRelatedMatch(matchID uuid.UUID) {
	t.MatchID = &matchID
}

// SetRelatedTournament sets the related tournament ID
func (t *BlockchainTransaction) SetRelatedTournament(tournamentID uuid.UUID) {
	t.TournamentID = &tournamentID
}

// SetWallet sets the related wallet ID
func (t *BlockchainTransaction) SetWallet(walletID uuid.UUID) {
	t.WalletID = &walletID
}

// SetLedgerTransaction sets the related ledger transaction ID
func (t *BlockchainTransaction) SetLedgerTransaction(ledgerTxID uuid.UUID) {
	t.LedgerTxID = &ledgerTxID
}
