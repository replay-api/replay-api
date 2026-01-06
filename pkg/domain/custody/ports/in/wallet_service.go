package custody_in

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// WalletService defines the interface for smart wallet operations
type WalletService interface {
	// Wallet Lifecycle
	CreateWallet(ctx context.Context, req *CreateWalletRequest) (*CreateWalletResult, error)
	GetWallet(ctx context.Context, walletID uuid.UUID) (*custody_entities.SmartWallet, error)
	GetWalletByAddress(ctx context.Context, chainID custody_vo.ChainID, address string) (*custody_entities.SmartWallet, error)
	GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*custody_entities.SmartWallet, error)
	DeployWallet(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID) (*DeployWalletResult, error)

	// Balance Operations
	GetBalance(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID) (*WalletBalance, error)
	GetAllBalances(ctx context.Context, walletID uuid.UUID) ([]*WalletBalance, error)
	GetTokenBalance(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID, tokenAddress string) (*TokenBalance, error)

	// Transaction Operations
	Transfer(ctx context.Context, req *TransferRequest) (*TransferResult, error)
	TransferToken(ctx context.Context, req *TokenTransferRequest) (*TransferResult, error)
	ExecuteTransaction(ctx context.Context, req *ExecuteTxRequest) (*ExecuteTxResult, error)
	BatchExecute(ctx context.Context, req *BatchExecuteRequest) (*BatchExecuteResult, error)

	// Transaction Status
	GetTransactionStatus(ctx context.Context, txID uuid.UUID) (*TxStatusResult, error)
	GetPendingTransactions(ctx context.Context, walletID uuid.UUID) ([]*PendingTx, error)

	// Spending Limits
	GetSpendingStatus(ctx context.Context, walletID uuid.UUID) (*SpendingStatus, error)
	UpdateSpendingLimits(ctx context.Context, walletID uuid.UUID, limits *custody_entities.TransactionLimits) error

	// Session Keys
	AddSessionKey(ctx context.Context, walletID uuid.UUID, req *AddSessionKeyRequest) (*SessionKeyResult, error)
	RevokeSessionKey(ctx context.Context, walletID uuid.UUID, keyAddress string) error
	GetSessionKeys(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.SessionKey, error)

	// Wallet Management
	FreezeWallet(ctx context.Context, walletID uuid.UUID, reason string) error
	UnfreezeWallet(ctx context.Context, walletID uuid.UUID) error
	UpdateKYCStatus(ctx context.Context, walletID uuid.UUID, status custody_entities.KYCStatus) error
}

// CreateWalletRequest for creating a new smart wallet
type CreateWalletRequest struct {
	ResourceOwner shared.ResourceOwner
	UserID        uuid.UUID
	TenantID      uuid.UUID
	WalletType    custody_entities.WalletType
	PrimaryChain  custody_vo.ChainID
	Chains        []custody_vo.ChainID // Additional chains
	Label         string
	Limits        *custody_entities.TransactionLimits
	KYCStatus     custody_entities.KYCStatus
	Metadata      map[string]string
}

// CreateWalletResult contains the created wallet information
type CreateWalletResult struct {
	Wallet     *custody_entities.SmartWallet
	MPCKeyID   string
	Addresses  map[custody_vo.ChainID]string
	DeployTx   map[custody_vo.ChainID]string // Deployment tx hashes
}

// DeployWalletResult contains wallet deployment result
type DeployWalletResult struct {
	ChainID     custody_vo.ChainID
	Address     string
	TxHash      string
	BlockNumber uint64
	GasUsed     uint64
}

// WalletBalance represents wallet balance
type WalletBalance struct {
	ChainID  custody_vo.ChainID
	Address  string
	Balance  *big.Int
	Symbol   string
	Decimals uint8
	USD      string // USD equivalent
}

// TokenBalance represents token balance
type TokenBalance struct {
	ChainID      custody_vo.ChainID
	Address      string
	TokenAddress string
	Symbol       string
	Balance      *big.Int
	Decimals     uint8
	USD          string
}

// TransferRequest for native token transfer
type TransferRequest struct {
	WalletID      uuid.UUID
	ChainID       custody_vo.ChainID
	To            string
	Amount        *big.Int
	GasLimit      *uint64
	MaxFeePerGas  *big.Int
	Note          string
	Metadata      map[string]string
	IdempotencyKey string
}

// TokenTransferRequest for ERC-20/SPL token transfer
type TokenTransferRequest struct {
	WalletID      uuid.UUID
	ChainID       custody_vo.ChainID
	TokenAddress  string
	To            string
	Amount        *big.Int
	GasLimit      *uint64
	MaxFeePerGas  *big.Int
	Note          string
	Metadata      map[string]string
	IdempotencyKey string
}

// TransferResult contains transfer result
type TransferResult struct {
	TxID        uuid.UUID
	TxHash      string
	ChainID     custody_vo.ChainID
	From        string
	To          string
	Amount      *big.Int
	TokenAddress *string
	Status      string
	GasUsed     *uint64
	GasCost     *big.Int
	SubmittedAt time.Time
	ConfirmedAt *time.Time
}

// ExecuteTxRequest for arbitrary transaction execution
type ExecuteTxRequest struct {
	WalletID      uuid.UUID
	ChainID       custody_vo.ChainID
	To            string
	Value         *big.Int
	Data          []byte
	GasLimit      *uint64
	MaxFeePerGas  *big.Int
	Note          string
	Metadata      map[string]string
	IdempotencyKey string
}

// ExecuteTxResult contains execution result
type ExecuteTxResult struct {
	TxID        uuid.UUID
	TxHash      string
	ChainID     custody_vo.ChainID
	Status      string
	GasUsed     *uint64
	GasCost     *big.Int
	ReturnData  []byte
	SubmittedAt time.Time
	ConfirmedAt *time.Time
}

// BatchExecuteRequest for batch transaction execution
type BatchExecuteRequest struct {
	WalletID      uuid.UUID
	ChainID       custody_vo.ChainID
	Transactions  []SingleTx
	GasLimit      *uint64
	MaxFeePerGas  *big.Int
	Note          string
	IdempotencyKey string
}

type SingleTx struct {
	To    string
	Value *big.Int
	Data  []byte
}

// BatchExecuteResult contains batch execution result
type BatchExecuteResult struct {
	TxID        uuid.UUID
	TxHash      string
	ChainID     custody_vo.ChainID
	Status      string
	Results     []SingleTxResult
	TotalGas    uint64
	GasCost     *big.Int
	SubmittedAt time.Time
	ConfirmedAt *time.Time
}

type SingleTxResult struct {
	Success    bool
	ReturnData []byte
	GasUsed    uint64
}

// TxStatusResult contains transaction status
type TxStatusResult struct {
	TxID          uuid.UUID
	TxHash        string
	ChainID       custody_vo.ChainID
	Status        string
	Confirmations uint64
	BlockNumber   *uint64
	GasUsed       *uint64
	GasCost       *big.Int
	FailureReason *string
	SubmittedAt   time.Time
	ConfirmedAt   *time.Time
}

// PendingTx represents a pending transaction
type PendingTx struct {
	TxID       uuid.UUID
	ChainID    custody_vo.ChainID
	TxHash     string
	To         string
	Value      *big.Int
	Status     string
	SubmittedAt time.Time
}

// SpendingStatus represents current spending status
type SpendingStatus struct {
	WalletID       uuid.UUID
	DailyLimit     *big.Int
	DailySpent     *big.Int
	DailyRemaining *big.Int
	WeeklyLimit    *big.Int
	WeeklySpent    *big.Int
	WeeklyRemaining *big.Int
	MonthlyLimit   *big.Int
	MonthlySpent   *big.Int
	MonthlyRemaining *big.Int
	PerTxLimit     *big.Int
	NextDailyReset time.Time
	NextWeeklyReset time.Time
	NextMonthlyReset time.Time
}

// AddSessionKeyRequest for adding a session key
type AddSessionKeyRequest struct {
	KeyAddress     string
	Label          string
	ValidFrom      time.Time
	ValidUntil     time.Time
	SpendingLimit  *big.Int
	AllowedTokens  []string
	AllowedTargets []string
}

// SessionKeyResult contains session key result
type SessionKeyResult struct {
	KeyAddress  string
	Label       string
	ValidFrom   time.Time
	ValidUntil  time.Time
	SpendingLimit *big.Int
	TxHash      *string // If on-chain registration needed
}
