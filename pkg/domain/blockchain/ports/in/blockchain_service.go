package blockchain_ports

import (
	"context"
	"math/big"

	"github.com/google/uuid"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// BlockchainService defines the main service interface for blockchain operations
type BlockchainService interface {
	// Prize Pool Operations
	CreatePrizePool(ctx context.Context, cmd CreatePrizePoolCommand) (*blockchain_entities.OnChainPrizePool, error)
	JoinPrizePool(ctx context.Context, cmd JoinPrizePoolCommand) error
	LockPrizePool(ctx context.Context, matchID uuid.UUID) error
	DistributePrizes(ctx context.Context, cmd DistributePrizesCommand) error
	CancelPrizePool(ctx context.Context, matchID uuid.UUID) error

	// User Operations
	DepositToVault(ctx context.Context, cmd DepositCommand) (*blockchain_entities.BlockchainTransaction, error)
	WithdrawFromVault(ctx context.Context, cmd WithdrawCommand) (*blockchain_entities.BlockchainTransaction, error)

	// Ledger Operations
	RecordTransaction(ctx context.Context, cmd RecordLedgerEntryCommand) error
	GetLedgerBalance(ctx context.Context, account wallet_vo.EVMAddress, token wallet_vo.EVMAddress) (*big.Int, error)

	// Sync Operations
	SyncPrizePool(ctx context.Context, matchID uuid.UUID) error
	SyncAllPendingPools(ctx context.Context) error

	// Query Operations
	GetPrizePool(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error)
	GetTransaction(ctx context.Context, txID uuid.UUID) (*blockchain_entities.BlockchainTransaction, error)
	GetTransactionsByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error)
}

// CreatePrizePoolCommand contains data for creating a prize pool
type CreatePrizePoolCommand struct {
	MatchID            uuid.UUID
	TokenAddress       wallet_vo.EVMAddress
	Currency           wallet_vo.Currency
	EntryFee           wallet_vo.Amount
	PlatformFeePercent uint16 // basis points (100 = 1%)
	ChainID            blockchain_vo.ChainID
}

// JoinPrizePoolCommand contains data for joining a prize pool
type JoinPrizePoolCommand struct {
	MatchID       uuid.UUID
	PlayerAddress wallet_vo.EVMAddress
	PlayerWalletID uuid.UUID
}

// DistributePrizesCommand contains data for distributing prizes
type DistributePrizesCommand struct {
	MatchID   uuid.UUID
	Winners   []WinnerShare
}

// WinnerShare represents a winner's share of the prize pool
type WinnerShare struct {
	Address   wallet_vo.EVMAddress
	WalletID  uuid.UUID
	Rank      uint8
	ShareBPS  uint16 // basis points (10000 = 100%)
	IsMVP     bool
}

// DepositCommand contains data for depositing to vault
type DepositCommand struct {
	UserAddress  wallet_vo.EVMAddress
	WalletID     uuid.UUID
	TokenAddress wallet_vo.EVMAddress
	Currency     wallet_vo.Currency
	Amount       wallet_vo.Amount
	ChainID      blockchain_vo.ChainID
}

// WithdrawCommand contains data for withdrawing from vault
type WithdrawCommand struct {
	UserAddress  wallet_vo.EVMAddress
	WalletID     uuid.UUID
	TokenAddress wallet_vo.EVMAddress
	Currency     wallet_vo.Currency
	Amount       wallet_vo.Amount
	ChainID      blockchain_vo.ChainID
}

// RecordLedgerEntryCommand contains data for recording a ledger entry
type RecordLedgerEntryCommand struct {
	TransactionID uuid.UUID
	Account       wallet_vo.EVMAddress
	Token         wallet_vo.EVMAddress
	Amount        *big.Int // positive = credit, negative = debit
	Category      string
	MatchID       *uuid.UUID
	TournamentID  *uuid.UUID
}

// BlockchainEventListener listens for blockchain events
type BlockchainEventListener interface {
	// Start listening for events
	Start(ctx context.Context) error

	// Stop listening
	Stop() error

	// Subscribe to specific event types
	OnPrizePoolCreated(handler func(ctx context.Context, event PrizePoolCreatedEvent) error)
	OnEntryFeeDeposited(handler func(ctx context.Context, event EntryFeeDepositedEvent) error)
	OnPrizePoolLocked(handler func(ctx context.Context, event PrizePoolLockedEvent) error)
	OnPrizeDistributed(handler func(ctx context.Context, event PrizeDistributedEvent) error)
	OnUserWithdrawal(handler func(ctx context.Context, event UserWithdrawalEvent) error)
	OnLedgerEntry(handler func(ctx context.Context, event LedgerEntryEvent) error)
}

// Event types
type PrizePoolCreatedEvent struct {
	MatchID     [32]byte
	Token       wallet_vo.EVMAddress
	EntryFee    *big.Int
	PlatformFee uint16
	TxHash      blockchain_vo.TxHash
	BlockNumber uint64
}

type EntryFeeDepositedEvent struct {
	MatchID     [32]byte
	Player      wallet_vo.EVMAddress
	Amount      *big.Int
	TxHash      blockchain_vo.TxHash
	BlockNumber uint64
}

type PrizePoolLockedEvent struct {
	MatchID     [32]byte
	TotalAmount *big.Int
	TxHash      blockchain_vo.TxHash
	BlockNumber uint64
}

type PrizeDistributedEvent struct {
	MatchID     [32]byte
	Winner      wallet_vo.EVMAddress
	Amount      *big.Int
	Rank        uint8
	TxHash      blockchain_vo.TxHash
	BlockNumber uint64
}

type UserWithdrawalEvent struct {
	User        wallet_vo.EVMAddress
	Token       wallet_vo.EVMAddress
	Amount      *big.Int
	TxHash      blockchain_vo.TxHash
	BlockNumber uint64
}

type LedgerEntryEvent struct {
	TransactionID [32]byte
	Account       wallet_vo.EVMAddress
	Token         wallet_vo.EVMAddress
	Amount        *big.Int
	Category      [32]byte
	EntryIndex    uint64
	TxHash        blockchain_vo.TxHash
	BlockNumber   uint64
}

// TransactionMonitor monitors pending transactions
type TransactionMonitor interface {
	// Start monitoring
	Start(ctx context.Context) error

	// Stop monitoring
	Stop() error

	// Add transaction to monitor
	Monitor(tx *blockchain_entities.BlockchainTransaction) error

	// Get pending count
	PendingCount() int

	// Set confirmation callback
	OnConfirmed(handler func(ctx context.Context, tx *blockchain_entities.BlockchainTransaction) error)
	OnFailed(handler func(ctx context.Context, tx *blockchain_entities.BlockchainTransaction, err error) error)
}

// GasEstimator provides gas estimation and pricing
type GasEstimator interface {
	// Estimate gas for operation
	EstimateGas(ctx context.Context, chainID blockchain_vo.ChainID, operation string, params map[string]interface{}) (uint64, error)

	// Get current gas price
	GetGasPrice(ctx context.Context, chainID blockchain_vo.ChainID) (*big.Int, error)

	// Get priority fee (EIP-1559)
	GetPriorityFee(ctx context.Context, chainID blockchain_vo.ChainID) (*big.Int, error)

	// Calculate total cost in USD
	CalculateCostUSD(ctx context.Context, chainID blockchain_vo.ChainID, gasLimit uint64) (float64, error)
}
