package blockchain_ports

import (
	"context"
	"math/big"

	"github.com/google/uuid"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// ChainClient defines the interface for interacting with a blockchain
type ChainClient interface {
	// Connection
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool
	GetChainID() blockchain_vo.ChainID

	// Block info
	GetLatestBlockNumber(ctx context.Context) (uint64, error)
	GetBlockByNumber(ctx context.Context, blockNumber uint64) (*BlockInfo, error)

	// Transaction
	GetTransaction(ctx context.Context, txHash blockchain_vo.TxHash) (*TransactionInfo, error)
	GetTransactionReceipt(ctx context.Context, txHash blockchain_vo.TxHash) (*TransactionReceipt, error)
	SendRawTransaction(ctx context.Context, signedTx []byte) (blockchain_vo.TxHash, error)
	EstimateGas(ctx context.Context, from, to wallet_vo.EVMAddress, data []byte, value *big.Int) (uint64, error)
	GetGasPrice(ctx context.Context) (*big.Int, error)

	// Account
	GetBalance(ctx context.Context, address wallet_vo.EVMAddress) (*big.Int, error)
	GetNonce(ctx context.Context, address wallet_vo.EVMAddress) (uint64, error)

	// Token
	GetTokenBalance(ctx context.Context, tokenAddr, accountAddr wallet_vo.EVMAddress) (*big.Int, error)
	GetTokenDecimals(ctx context.Context, tokenAddr wallet_vo.EVMAddress) (uint8, error)

	// Events
	SubscribeToEvents(ctx context.Context, contractAddr wallet_vo.EVMAddress, topics [][]byte) (<-chan ContractEvent, error)
	GetLogs(ctx context.Context, filter LogFilter) ([]ContractEvent, error)
}

// BlockInfo represents block data
type BlockInfo struct {
	Number     uint64
	Hash       blockchain_vo.BlockHash
	ParentHash blockchain_vo.BlockHash
	Timestamp  uint64
	GasLimit   uint64
	GasUsed    uint64
	TxCount    int
}

// TransactionInfo represents on-chain transaction data
type TransactionInfo struct {
	Hash        blockchain_vo.TxHash
	BlockNumber uint64
	BlockHash   blockchain_vo.BlockHash
	From        wallet_vo.EVMAddress
	To          *wallet_vo.EVMAddress
	Value       *big.Int
	GasPrice    *big.Int
	GasLimit    uint64
	Nonce       uint64
	Input       []byte
}

// TransactionReceipt represents transaction receipt
type TransactionReceipt struct {
	TxHash          blockchain_vo.TxHash
	BlockNumber     uint64
	BlockHash       blockchain_vo.BlockHash
	Status          uint64 // 1 = success, 0 = failure
	GasUsed         uint64
	ContractAddress *wallet_vo.EVMAddress
	Logs            []ContractEvent
}

// ContractEvent represents a contract event/log
type ContractEvent struct {
	Address     wallet_vo.EVMAddress
	Topics      [][]byte
	Data        []byte
	BlockNumber uint64
	TxHash      blockchain_vo.TxHash
	LogIndex    uint
}

// LogFilter defines parameters for filtering logs
type LogFilter struct {
	FromBlock   uint64
	ToBlock     *uint64
	Addresses   []wallet_vo.EVMAddress
	Topics      [][]byte
}

// MultiChainClient manages multiple chain clients
type MultiChainClient interface {
	// Get client for specific chain
	GetClient(chainID blockchain_vo.ChainID) (ChainClient, error)

	// Get primary chain client
	GetPrimaryClient() ChainClient

	// Health check across all chains
	HealthCheck(ctx context.Context) map[blockchain_vo.ChainID]bool

	// Get best chain for transaction (based on gas, speed, etc.)
	GetOptimalChain(ctx context.Context) blockchain_vo.ChainID
}

// VaultContract defines interface for LeetVault contract interactions
type VaultContract interface {
	// Prize Pool Management
	CreatePrizePool(ctx context.Context, matchID uuid.UUID, tokenAddr wallet_vo.EVMAddress, entryFee *big.Int, platformFeeBPS uint16) (blockchain_vo.TxHash, error)
	DepositEntryFee(ctx context.Context, matchID uuid.UUID, playerAddr wallet_vo.EVMAddress) (blockchain_vo.TxHash, error)
	LockPrizePool(ctx context.Context, matchID uuid.UUID) (blockchain_vo.TxHash, error)
	StartEscrow(ctx context.Context, matchID uuid.UUID) (blockchain_vo.TxHash, error)
	DistributePrizes(ctx context.Context, matchID uuid.UUID, winners []wallet_vo.EVMAddress, sharesBPS []uint16) (blockchain_vo.TxHash, error)
	CancelPrizePool(ctx context.Context, matchID uuid.UUID) (blockchain_vo.TxHash, error)

	// User Operations
	Deposit(ctx context.Context, userAddr wallet_vo.EVMAddress, tokenAddr wallet_vo.EVMAddress, amount *big.Int) (blockchain_vo.TxHash, error)
	Withdraw(ctx context.Context, userAddr wallet_vo.EVMAddress, tokenAddr wallet_vo.EVMAddress, amount *big.Int) (blockchain_vo.TxHash, error)

	// View Functions
	GetPrizePoolInfo(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error)
	GetUserBalance(ctx context.Context, userAddr, tokenAddr wallet_vo.EVMAddress) (*big.Int, error)
	GetParticipants(ctx context.Context, matchID uuid.UUID) ([]wallet_vo.EVMAddress, error)
}

// LedgerContract defines interface for LeetLedger contract interactions
type LedgerContract interface {
	// Record entries
	RecordEntry(ctx context.Context, txID uuid.UUID, account wallet_vo.EVMAddress, token wallet_vo.EVMAddress, amount *big.Int, category string, matchID *uuid.UUID) (blockchain_vo.TxHash, error)
	RecordBatch(ctx context.Context, batchID uuid.UUID, entries []LedgerEntryInput) (blockchain_vo.TxHash, error)
	RecordTransfer(ctx context.Context, txID uuid.UUID, from, to wallet_vo.EVMAddress, token wallet_vo.EVMAddress, amount *big.Int, category string, matchID *uuid.UUID) (blockchain_vo.TxHash, error)

	// View Functions
	GetEntry(ctx context.Context, index uint64) (*OnChainLedgerEntry, error)
	GetEntryByTxID(ctx context.Context, txID uuid.UUID) (*OnChainLedgerEntry, error)
	GetAccountEntries(ctx context.Context, account wallet_vo.EVMAddress, start, limit uint64) ([]OnChainLedgerEntry, error)
	GetAccountBalance(ctx context.Context, account, token wallet_vo.EVMAddress) (*big.Int, error)
	GetMatchEntries(ctx context.Context, matchID uuid.UUID) ([]OnChainLedgerEntry, error)
	GetCurrentMerkleRoot(ctx context.Context) ([]byte, error)
	VerifyChainIntegrity(ctx context.Context, startIndex, endIndex uint64) (bool, error)
}

// LedgerEntryInput represents input for batch recording
type LedgerEntryInput struct {
	TransactionID uuid.UUID
	Account       wallet_vo.EVMAddress
	Token         wallet_vo.EVMAddress
	Amount        *big.Int
	Category      string
	MatchID       *uuid.UUID
}

// OnChainLedgerEntry represents a ledger entry from the blockchain
type OnChainLedgerEntry struct {
	TransactionID [32]byte
	Account       wallet_vo.EVMAddress
	Token         wallet_vo.EVMAddress
	Amount        *big.Int
	Category      [32]byte
	MatchID       [32]byte
	TournamentID  [32]byte
	Timestamp     uint64
	BlockNumber   uint64
	PreviousHash  [32]byte
	MerkleRoot    [32]byte
}

// TransactionRepository stores blockchain transactions in cache
type TransactionRepository interface {
	Save(ctx context.Context, tx *blockchain_entities.BlockchainTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*blockchain_entities.BlockchainTransaction, error)
	FindByTxHash(ctx context.Context, chainID blockchain_vo.ChainID, txHash blockchain_vo.TxHash) (*blockchain_entities.BlockchainTransaction, error)
	FindPending(ctx context.Context, chainID blockchain_vo.ChainID) ([]*blockchain_entities.BlockchainTransaction, error)
	FindByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error)
	FindByMatch(ctx context.Context, matchID uuid.UUID) ([]*blockchain_entities.BlockchainTransaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status blockchain_entities.TransactionStatus, confirmations uint64) error
}

// PrizePoolRepository stores on-chain prize pools in cache
type PrizePoolRepository interface {
	Save(ctx context.Context, pool *blockchain_entities.OnChainPrizePool) error
	FindByID(ctx context.Context, id uuid.UUID) (*blockchain_entities.OnChainPrizePool, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error)
	FindByStatus(ctx context.Context, status blockchain_entities.OnChainPrizePoolStatus) ([]*blockchain_entities.OnChainPrizePool, error)
	FindPendingDistribution(ctx context.Context) ([]*blockchain_entities.OnChainPrizePool, error)
	UpdateSyncState(ctx context.Context, id uuid.UUID, blockNumber uint64, synced bool) error
}

// SyncStateRepository tracks blockchain sync state
type SyncStateRepository interface {
	GetLastSyncedBlock(ctx context.Context, chainID blockchain_vo.ChainID, contractAddr wallet_vo.EVMAddress) (uint64, error)
	SetLastSyncedBlock(ctx context.Context, chainID blockchain_vo.ChainID, contractAddr wallet_vo.EVMAddress, blockNumber uint64) error
}
