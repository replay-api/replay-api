package custody_out

import (
	"context"
	"math/big"
	"time"

	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// ChainClient defines the interface for blockchain interactions
// Implementations exist for Solana, EVM chains (Polygon, Base, Arbitrum)
type ChainClient interface {
	// Chain Info
	GetChainID() custody_vo.ChainID
	GetChainInfo(ctx context.Context) (*ChainInfo, error)
	GetLatestBlock(ctx context.Context) (*BlockInfo, error)

	// Account Operations
	GetBalance(ctx context.Context, address string) (*Balance, error)
	GetTokenBalance(ctx context.Context, address string, tokenAddress string) (*TokenBalance, error)
	GetNonce(ctx context.Context, address string) (uint64, error)

	// Transaction Building
	BuildTransaction(ctx context.Context, req *TxBuildRequest) (*UnsignedTransaction, error)
	EstimateGas(ctx context.Context, tx *UnsignedTransaction) (*GasEstimate, error)
	GetGasPrice(ctx context.Context) (*GasPriceInfo, error)

	// Transaction Submission
	SubmitTransaction(ctx context.Context, signedTx []byte) (*TxSubmitResult, error)
	GetTransaction(ctx context.Context, txHash string) (*TransactionInfo, error)
	WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) (*TransactionReceipt, error)

	// Smart Contract Interactions
	CallContract(ctx context.Context, req *ContractCallRequest) ([]byte, error)
	GetContractLogs(ctx context.Context, filter *LogFilter) ([]*ContractLog, error)

	// Health
	HealthCheck(ctx context.Context) error
}

// SolanaClient extends ChainClient with Solana-specific operations
type SolanaClient interface {
	ChainClient

	// Solana-specific
	GetAccountInfo(ctx context.Context, address string) (*SolanaAccountInfo, error)
	GetTokenAccounts(ctx context.Context, owner string) ([]*SolanaTokenAccount, error)
	CreateAssociatedTokenAccount(ctx context.Context, owner string, mint string) (*UnsignedTransaction, error)

	// Program Interactions
	BuildProgramInstruction(ctx context.Context, req *ProgramInstructionRequest) (*UnsignedTransaction, error)
	GetProgramAccounts(ctx context.Context, programID string, filter *ProgramAccountFilter) ([]*SolanaAccountInfo, error)

	// SPL Token Operations
	BuildSPLTransfer(ctx context.Context, req *SPLTransferRequest) (*UnsignedTransaction, error)

	// Priority Fees
	GetRecentPriorityFees(ctx context.Context) (*PriorityFeeInfo, error)
}

// EVMClient extends ChainClient with EVM-specific operations
type EVMClient interface {
	ChainClient

	// EVM-specific
	GetCode(ctx context.Context, address string) ([]byte, error)
	GetStorageAt(ctx context.Context, address string, slot string) ([]byte, error)

	// ERC-4337 Account Abstraction
	BuildUserOperation(ctx context.Context, req *UserOpRequest) (*UserOperation, error)
	EstimateUserOperationGas(ctx context.Context, userOp *UserOperation) (*UserOpGasEstimate, error)
	SubmitUserOperation(ctx context.Context, userOp *UserOperation) (*UserOpResult, error)
	GetUserOperationReceipt(ctx context.Context, userOpHash string) (*UserOpReceipt, error)

	// ERC-20 Token Operations
	BuildERC20Transfer(ctx context.Context, req *ERC20TransferRequest) (*UnsignedTransaction, error)
	BuildERC20Approve(ctx context.Context, req *ERC20ApproveRequest) (*UnsignedTransaction, error)
	GetERC20Allowance(ctx context.Context, token string, owner string, spender string) (*big.Int, error)
}

// ChainInfo contains chain information
type ChainInfo struct {
	ChainID       custody_vo.ChainID
	Name          string
	NativeCurrency string
	BlockTime     time.Duration
	IsTestnet     bool
}

// BlockInfo contains block information
type BlockInfo struct {
	Number    uint64
	Hash      string
	Timestamp time.Time
	ParentHash string
}

// Balance represents native token balance
type Balance struct {
	Address  string
	Balance  *big.Int
	Decimals uint8
	Symbol   string
}

// TokenBalance represents token balance
type TokenBalance struct {
	Address      string
	TokenAddress string
	Balance      *big.Int
	Decimals     uint8
	Symbol       string
}

// TxBuildRequest represents a transaction build request
type TxBuildRequest struct {
	From        string
	To          string
	Value       *big.Int
	Data        []byte
	GasLimit    uint64
	GasPrice    *big.Int
	MaxFeePerGas *big.Int
	MaxPriorityFee *big.Int
	Nonce       *uint64
}

// UnsignedTransaction represents an unsigned transaction
type UnsignedTransaction struct {
	ChainID     custody_vo.ChainID
	From        string
	To          string
	Value       *big.Int
	Data        []byte
	GasLimit    uint64
	GasPrice    *big.Int
	MaxFeePerGas *big.Int
	MaxPriorityFee *big.Int
	Nonce       uint64
	RawTx       []byte // Serialized unsigned transaction
	MessageHash []byte // Hash to sign
}

// GasEstimate contains gas estimation
type GasEstimate struct {
	GasLimit     uint64
	GasPrice     *big.Int
	MaxFeePerGas *big.Int
	MaxPriorityFee *big.Int
	EstimatedCost *big.Int
}

// GasPriceInfo contains current gas prices
type GasPriceInfo struct {
	BaseFee    *big.Int
	SafeLow    *big.Int
	Standard   *big.Int
	Fast       *big.Int
	Instant    *big.Int
}

// TxSubmitResult contains transaction submission result
type TxSubmitResult struct {
	TxHash      string
	SubmittedAt time.Time
}

// TransactionInfo contains transaction information
type TransactionInfo struct {
	Hash        string
	From        string
	To          string
	Value       *big.Int
	Data        []byte
	GasLimit    uint64
	GasPrice    *big.Int
	Nonce       uint64
	BlockNumber *uint64
	BlockHash   *string
	Status      TxStatus
}

type TxStatus string

const (
	TxStatusPending   TxStatus = "Pending"
	TxStatusConfirmed TxStatus = "Confirmed"
	TxStatusFailed    TxStatus = "Failed"
)

// TransactionReceipt contains transaction receipt
type TransactionReceipt struct {
	TxHash          string
	BlockNumber     uint64
	BlockHash       string
	GasUsed         uint64
	EffectiveGasPrice *big.Int
	Status          bool
	Logs            []*ContractLog
	ContractAddress *string
}

// ContractCallRequest for calling smart contract methods
type ContractCallRequest struct {
	To       string
	Data     []byte
	From     *string
	BlockNum *uint64
}

// LogFilter for filtering contract logs
type LogFilter struct {
	Address   []string
	Topics    [][]string
	FromBlock uint64
	ToBlock   uint64
}

// ContractLog represents a contract event log
type ContractLog struct {
	Address     string
	Topics      []string
	Data        []byte
	BlockNumber uint64
	TxHash      string
	LogIndex    uint32
}

// Solana-specific types

type SolanaAccountInfo struct {
	Address    string
	Lamports   uint64
	Owner      string
	Executable bool
	RentEpoch  uint64
	Data       []byte
}

type SolanaTokenAccount struct {
	Address  string
	Mint     string
	Owner    string
	Amount   uint64
	Decimals uint8
}

type ProgramInstructionRequest struct {
	ProgramID string
	Accounts  []AccountMeta
	Data      []byte
}

type AccountMeta struct {
	Address    string
	IsSigner   bool
	IsWritable bool
}

type ProgramAccountFilter struct {
	DataSize   *uint64
	Memcmp     []MemcmpFilter
}

type MemcmpFilter struct {
	Offset uint64
	Bytes  []byte
}

type SPLTransferRequest struct {
	From       string
	To         string
	Mint       string
	Amount     uint64
	CreateATA  bool // Create Associated Token Account if needed
}

type PriorityFeeInfo struct {
	Min     uint64
	Low     uint64
	Medium  uint64
	High    uint64
	VeryHigh uint64
}

// EVM ERC-4337 types

type UserOpRequest struct {
	Sender        string
	Target        string
	Value         *big.Int
	CallData      []byte
	Paymaster     *string
	PaymasterData []byte
}

type UserOperation struct {
	Sender               string
	Nonce                *big.Int
	InitCode             []byte
	CallData             []byte
	AccountGasLimits     [32]byte
	PreVerificationGas   *big.Int
	GasFees              [32]byte
	PaymasterAndData     []byte
	Signature            []byte
	UserOpHash           []byte // Hash to sign
}

type UserOpGasEstimate struct {
	PreVerificationGas uint64
	VerificationGasLimit uint64
	CallGasLimit       uint64
	PaymasterVerificationGas uint64
	PaymasterPostOpGas uint64
	MaxFeePerGas       *big.Int
	MaxPriorityFeePerGas *big.Int
}

type UserOpResult struct {
	UserOpHash  string
	SubmittedAt time.Time
}

type UserOpReceipt struct {
	UserOpHash    string
	Success       bool
	ActualGasCost *big.Int
	ActualGasUsed uint64
	TxHash        string
	BlockNumber   uint64
}

// ERC-20 types

type ERC20TransferRequest struct {
	Token   string
	From    string
	To      string
	Amount  *big.Int
}

type ERC20ApproveRequest struct {
	Token   string
	Owner   string
	Spender string
	Amount  *big.Int
}
