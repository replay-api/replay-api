package blockchain_services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_in "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/in"
	blockchain_out "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/out"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// =============================================================================
// Test Helpers
// =============================================================================

// testResourceOwner creates a test resource owner
func testResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   uuid.New(),
	}
}

// testContext creates a context with resource owner
func testContext() context.Context {
	ro := testResourceOwner()
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, ro.TenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, ro.ClientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, ro.UserID)
	return ctx
}

// testEVMAddress creates the first test EVM address
func testEVMAddress() wallet_vo.EVMAddress {
	addr, _ := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	return addr
}

// testEVMAddress2 creates the second test EVM address
func testEVMAddress2() wallet_vo.EVMAddress {
	addr, _ := wallet_vo.NewEVMAddress("0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199")
	return addr
}

// testTokenAddress creates a test token address
func testTokenAddress() wallet_vo.EVMAddress {
	addr, _ := wallet_vo.NewEVMAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	return addr
}

// testTxHash creates a test transaction hash
func testTxHash() blockchain_vo.TxHash {
	hash, _ := blockchain_vo.NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	return hash
}

// Suppress unused import
var _ = fmt.Sprint

// Mock implementations for testing

type mockMultiChainClient struct {
	latestBlockNumber uint64
	getClientErr      error
	primaryChain      *mockChainClient
}

func (m *mockMultiChainClient) GetClient(chainID blockchain_vo.ChainID) (blockchain_out.ChainClient, error) {
	if m.getClientErr != nil {
		return nil, m.getClientErr
	}
	return &mockChainClient{blockNumber: m.latestBlockNumber, chainID: chainID}, nil
}

func (m *mockMultiChainClient) GetPrimaryClient() blockchain_out.ChainClient {
	if m.primaryChain != nil {
		return m.primaryChain
	}
	return &mockChainClient{blockNumber: m.latestBlockNumber, chainID: blockchain_vo.PrimaryChain()}
}

func (m *mockMultiChainClient) HealthCheck(ctx context.Context) map[blockchain_vo.ChainID]bool {
	return map[blockchain_vo.ChainID]bool{blockchain_vo.PrimaryChain(): true}
}

func (m *mockMultiChainClient) GetOptimalChain(ctx context.Context) blockchain_vo.ChainID {
	return blockchain_vo.PrimaryChain()
}

type mockChainClient struct {
	blockNumber uint64
	chainID     blockchain_vo.ChainID
}

func (m *mockChainClient) Connect(ctx context.Context) error { return nil }
func (m *mockChainClient) Disconnect() error                 { return nil }
func (m *mockChainClient) IsConnected() bool                 { return true }
func (m *mockChainClient) GetChainID() blockchain_vo.ChainID { return m.chainID }
func (m *mockChainClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	return m.blockNumber, nil
}
func (m *mockChainClient) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*blockchain_out.BlockInfo, error) {
	return &blockchain_out.BlockInfo{Number: blockNumber}, nil
}
func (m *mockChainClient) GetTransaction(ctx context.Context, txHash blockchain_vo.TxHash) (*blockchain_out.TransactionInfo, error) {
	return &blockchain_out.TransactionInfo{Hash: txHash}, nil
}
func (m *mockChainClient) GetTransactionReceipt(ctx context.Context, txHash blockchain_vo.TxHash) (*blockchain_out.TransactionReceipt, error) {
	return &blockchain_out.TransactionReceipt{TxHash: txHash, Status: 1}, nil
}
func (m *mockChainClient) SendRawTransaction(ctx context.Context, signedTx []byte) (blockchain_vo.TxHash, error) {
	return blockchain_vo.TxHash{}, nil
}
func (m *mockChainClient) EstimateGas(ctx context.Context, from, to wallet_vo.EVMAddress, data []byte, value *big.Int) (uint64, error) {
	return 21000, nil
}
func (m *mockChainClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1000000000), nil
}
func (m *mockChainClient) GetBalance(ctx context.Context, address wallet_vo.EVMAddress) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil
}
func (m *mockChainClient) GetNonce(ctx context.Context, address wallet_vo.EVMAddress) (uint64, error) {
	return 0, nil
}
func (m *mockChainClient) GetTokenBalance(ctx context.Context, tokenAddr, accountAddr wallet_vo.EVMAddress) (*big.Int, error) {
	return big.NewInt(1000000), nil
}
func (m *mockChainClient) GetTokenDecimals(ctx context.Context, tokenAddr wallet_vo.EVMAddress) (uint8, error) {
	return 6, nil
}
func (m *mockChainClient) SubscribeToEvents(ctx context.Context, contractAddr wallet_vo.EVMAddress, topics [][]byte) (<-chan blockchain_out.ContractEvent, error) {
	return nil, nil
}
func (m *mockChainClient) GetLogs(ctx context.Context, filter blockchain_out.LogFilter) ([]blockchain_out.ContractEvent, error) {
	return nil, nil
}

type mockVaultContract struct {
	createPoolTxHash blockchain_vo.TxHash
	createPoolErr    error
	depositTxHash    blockchain_vo.TxHash
	depositErr       error
	withdrawTxHash   blockchain_vo.TxHash
	withdrawErr      error
	lockPoolTxHash   blockchain_vo.TxHash
	lockPoolErr      error
	distributeTxHash blockchain_vo.TxHash
	distributeErr    error
	cancelTxHash     blockchain_vo.TxHash
	cancelErr        error
	prizePoolInfo    *blockchain_entities.OnChainPrizePool
	prizePoolInfoErr error
}

func (m *mockVaultContract) CreatePrizePool(ctx context.Context, matchID uuid.UUID, tokenAddress wallet_vo.EVMAddress, entryFee *big.Int, platformFeePct uint16) (blockchain_vo.TxHash, error) {
	return m.createPoolTxHash, m.createPoolErr
}

func (m *mockVaultContract) DepositEntryFee(ctx context.Context, matchID uuid.UUID, player wallet_vo.EVMAddress) (blockchain_vo.TxHash, error) {
	return m.depositTxHash, m.depositErr
}

func (m *mockVaultContract) LockPrizePool(ctx context.Context, matchID uuid.UUID) (blockchain_vo.TxHash, error) {
	return m.lockPoolTxHash, m.lockPoolErr
}

func (m *mockVaultContract) DistributePrizes(ctx context.Context, matchID uuid.UUID, winners []wallet_vo.EVMAddress, shares []uint16) (blockchain_vo.TxHash, error) {
	return m.distributeTxHash, m.distributeErr
}

func (m *mockVaultContract) CancelPrizePool(ctx context.Context, matchID uuid.UUID) (blockchain_vo.TxHash, error) {
	return m.cancelTxHash, m.cancelErr
}

func (m *mockVaultContract) Deposit(ctx context.Context, user wallet_vo.EVMAddress, token wallet_vo.EVMAddress, amount *big.Int) (blockchain_vo.TxHash, error) {
	return m.depositTxHash, m.depositErr
}

func (m *mockVaultContract) Withdraw(ctx context.Context, user wallet_vo.EVMAddress, token wallet_vo.EVMAddress, amount *big.Int) (blockchain_vo.TxHash, error) {
	return m.withdrawTxHash, m.withdrawErr
}

func (m *mockVaultContract) GetPrizePoolInfo(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	return m.prizePoolInfo, m.prizePoolInfoErr
}

func (m *mockVaultContract) StartEscrow(ctx context.Context, matchID uuid.UUID) (blockchain_vo.TxHash, error) {
	return m.lockPoolTxHash, m.lockPoolErr
}

func (m *mockVaultContract) GetUserBalance(ctx context.Context, userAddr, tokenAddr wallet_vo.EVMAddress) (*big.Int, error) {
	return big.NewInt(1000000), nil
}

func (m *mockVaultContract) GetParticipants(ctx context.Context, matchID uuid.UUID) ([]wallet_vo.EVMAddress, error) {
	return []wallet_vo.EVMAddress{}, nil
}

type mockLedgerContract struct {
	recordEntryTxHash blockchain_vo.TxHash
	recordEntryErr    error
	balance           *big.Int
	balanceErr        error
	entries           []blockchain_out.OnChainLedgerEntry
}

func (m *mockLedgerContract) RecordEntry(ctx context.Context, txID uuid.UUID, account, token wallet_vo.EVMAddress, amount *big.Int, category string, matchID *uuid.UUID) (blockchain_vo.TxHash, error) {
	return m.recordEntryTxHash, m.recordEntryErr
}

func (m *mockLedgerContract) RecordBatch(ctx context.Context, batchID uuid.UUID, entries []blockchain_out.LedgerEntryInput) (blockchain_vo.TxHash, error) {
	return m.recordEntryTxHash, m.recordEntryErr
}

func (m *mockLedgerContract) RecordTransfer(ctx context.Context, txID uuid.UUID, from, to wallet_vo.EVMAddress, token wallet_vo.EVMAddress, amount *big.Int, category string, matchID *uuid.UUID) (blockchain_vo.TxHash, error) {
	return m.recordEntryTxHash, m.recordEntryErr
}

func (m *mockLedgerContract) GetEntry(ctx context.Context, index uint64) (*blockchain_out.OnChainLedgerEntry, error) {
	if len(m.entries) > int(index) {
		return &m.entries[index], nil
	}
	return nil, errors.New("entry not found")
}

func (m *mockLedgerContract) GetEntryByTxID(ctx context.Context, txID uuid.UUID) (*blockchain_out.OnChainLedgerEntry, error) {
	if len(m.entries) > 0 {
		return &m.entries[0], nil
	}
	return nil, errors.New("entry not found")
}

func (m *mockLedgerContract) GetAccountEntries(ctx context.Context, account wallet_vo.EVMAddress, start, limit uint64) ([]blockchain_out.OnChainLedgerEntry, error) {
	return m.entries, nil
}

func (m *mockLedgerContract) GetAccountBalance(ctx context.Context, account, token wallet_vo.EVMAddress) (*big.Int, error) {
	return m.balance, m.balanceErr
}

func (m *mockLedgerContract) GetMatchEntries(ctx context.Context, matchID uuid.UUID) ([]blockchain_out.OnChainLedgerEntry, error) {
	return m.entries, nil
}

func (m *mockLedgerContract) GetCurrentMerkleRoot(ctx context.Context) ([]byte, error) {
	return make([]byte, 32), nil
}

func (m *mockLedgerContract) VerifyChainIntegrity(ctx context.Context, startIndex, endIndex uint64) (bool, error) {
	return true, nil
}

type mockTransactionRepository struct {
	savedTx     *blockchain_entities.BlockchainTransaction
	saveErr     error
	foundTx     *blockchain_entities.BlockchainTransaction
	findErr     error
	walletTxs   []*blockchain_entities.BlockchainTransaction
	walletTotal int64
	matchTxs    []*blockchain_entities.BlockchainTransaction
	pendingTxs  []*blockchain_entities.BlockchainTransaction
}

func (m *mockTransactionRepository) Save(ctx context.Context, tx *blockchain_entities.BlockchainTransaction) error {
	m.savedTx = tx
	return m.saveErr
}

func (m *mockTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*blockchain_entities.BlockchainTransaction, error) {
	return m.foundTx, m.findErr
}

func (m *mockTransactionRepository) FindByTxHash(ctx context.Context, chainID blockchain_vo.ChainID, txHash blockchain_vo.TxHash) (*blockchain_entities.BlockchainTransaction, error) {
	return m.foundTx, m.findErr
}

func (m *mockTransactionRepository) FindPending(ctx context.Context, chainID blockchain_vo.ChainID) ([]*blockchain_entities.BlockchainTransaction, error) {
	return m.pendingTxs, m.findErr
}

func (m *mockTransactionRepository) FindByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error) {
	return m.walletTxs, m.walletTotal, m.findErr
}

func (m *mockTransactionRepository) FindByMatch(ctx context.Context, matchID uuid.UUID) ([]*blockchain_entities.BlockchainTransaction, error) {
	return m.matchTxs, m.findErr
}

func (m *mockTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status blockchain_entities.TransactionStatus, confirmations uint64) error {
	return m.saveErr
}

type mockPrizePoolRepository struct {
	savedPool    *blockchain_entities.OnChainPrizePool
	saveErr      error
	foundPool    *blockchain_entities.OnChainPrizePool
	findErr      error
	pendingPools []*blockchain_entities.OnChainPrizePool
	statusPools  []*blockchain_entities.OnChainPrizePool
}

func (m *mockPrizePoolRepository) Save(ctx context.Context, pool *blockchain_entities.OnChainPrizePool) error {
	m.savedPool = pool
	return m.saveErr
}

func (m *mockPrizePoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	return m.foundPool, m.findErr
}

func (m *mockPrizePoolRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	return m.foundPool, m.findErr
}

func (m *mockPrizePoolRepository) FindByStatus(ctx context.Context, status blockchain_entities.OnChainPrizePoolStatus) ([]*blockchain_entities.OnChainPrizePool, error) {
	return m.statusPools, nil
}

func (m *mockPrizePoolRepository) FindPendingDistribution(ctx context.Context) ([]*blockchain_entities.OnChainPrizePool, error) {
	return m.pendingPools, nil
}

func (m *mockPrizePoolRepository) UpdateSyncState(ctx context.Context, id uuid.UUID, blockNumber uint64, synced bool) error {
	return m.saveErr
}

type mockSyncStateRepository struct {
	lastSyncedBlock uint64
}

func (m *mockSyncStateRepository) GetLastSyncedBlock(ctx context.Context, chainID blockchain_vo.ChainID, contractAddr wallet_vo.EVMAddress) (uint64, error) {
	return m.lastSyncedBlock, nil
}

func (m *mockSyncStateRepository) SetLastSyncedBlock(ctx context.Context, chainID blockchain_vo.ChainID, contractAddr wallet_vo.EVMAddress, blockNumber uint64) error {
	m.lastSyncedBlock = blockNumber
	return nil
}

// Tests

func TestNewBlockchainService(t *testing.T) {
	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		&mockTransactionRepository{},
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.escrowDuration != 72*time.Hour {
		t.Errorf("expected escrow duration 72h, got %v", svc.escrowDuration)
	}
}

func TestSetContractAddresses(t *testing.T) {
	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		&mockTransactionRepository{},
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	chainID := blockchain_vo.ChainID(1)
	vaultAddr := testEVMAddress()
	ledgerAddr := testEVMAddress2()
	tokenAddr := testTokenAddress()

	svc.SetVaultAddress(chainID, vaultAddr)
	svc.SetLedgerAddress(chainID, ledgerAddr)
	svc.SetTokenAddress(chainID, wallet_vo.CurrencyUSDC, tokenAddr)

	if svc.vaultAddresses[chainID] != vaultAddr {
		t.Error("vault address not set correctly")
	}
	if svc.ledgerAddresses[chainID] != ledgerAddr {
		t.Error("ledger address not set correctly")
	}
	if svc.tokenAddresses[chainID][wallet_vo.CurrencyUSDC] != tokenAddr {
		t.Error("token address not set correctly")
	}
}

func TestCreatePrizePool_Success(t *testing.T) {
	txHash := testTxHash()
	vaultContract := &mockVaultContract{createPoolTxHash: txHash}
	prizePoolRepo := &mockPrizePoolRepository{}
	txRepo := &mockTransactionRepository{}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		&mockLedgerContract{},
		txRepo,
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	chainID := blockchain_vo.PrimaryChain()
	vaultAddr := testEVMAddress()
	svc.SetVaultAddress(chainID, vaultAddr)

	ctx := testContext()

	cmd := blockchain_in.CreatePrizePoolCommand{
		MatchID:            uuid.New(),
		ChainID:            chainID,
		TokenAddress:       testTokenAddress(),
		Currency:           wallet_vo.CurrencyUSDC,
		EntryFee:           wallet_vo.NewAmount(10),
		PlatformFeePercent: 1000, // 10%
	}

	pool, err := svc.CreatePrizePool(ctx, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if pool.MatchID != cmd.MatchID {
		t.Errorf("expected match ID %v, got %v", cmd.MatchID, pool.MatchID)
	}
	if pool.Status != blockchain_entities.PoolStatusAccumulating {
		t.Errorf("expected status Accumulating, got %v", pool.Status)
	}
	if prizePoolRepo.savedPool == nil {
		t.Error("pool was not saved")
	}
	if txRepo.savedTx == nil {
		t.Error("transaction was not saved")
	}
}

func TestCreatePrizePool_VaultNotDeployed(t *testing.T) {
	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		&mockTransactionRepository{},
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	// Don't set vault address
	ctx := testContext()

	cmd := blockchain_in.CreatePrizePoolCommand{
		MatchID:  uuid.New(),
		ChainID:  blockchain_vo.ChainID(1),
		EntryFee: wallet_vo.NewAmount(10),
	}

	_, err := svc.CreatePrizePool(ctx, cmd)
	if err == nil {
		t.Fatal("expected error for vault not deployed")
	}
}

func TestCreatePrizePool_ContractError(t *testing.T) {
	vaultContract := &mockVaultContract{
		createPoolErr: errors.New("contract error"),
	}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		&mockLedgerContract{},
		&mockTransactionRepository{},
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	chainID := blockchain_vo.PrimaryChain()
	svc.SetVaultAddress(chainID, testEVMAddress())

	ctx := testContext()

	cmd := blockchain_in.CreatePrizePoolCommand{
		MatchID:  uuid.New(),
		ChainID:  chainID,
		EntryFee: wallet_vo.NewAmount(10),
	}

	_, err := svc.CreatePrizePool(ctx, cmd)
	if err == nil {
		t.Fatal("expected error from contract")
	}
}

func TestJoinPrizePool_Success(t *testing.T) {
	matchID := uuid.New()
	existingPool := blockchain_entities.NewOnChainPrizePool(
		testResourceOwner(),
		matchID,
		blockchain_vo.PrimaryChain(),
		testEVMAddress(),
		testTokenAddress(),
		wallet_vo.CurrencyUSDC,
		wallet_vo.NewAmount(10),
		1000,
	)
	existingPool.Status = blockchain_entities.PoolStatusAccumulating

	txHash := testTxHash()
	vaultContract := &mockVaultContract{depositTxHash: txHash}
	prizePoolRepo := &mockPrizePoolRepository{foundPool: existingPool}
	txRepo := &mockTransactionRepository{}
	ledgerContract := &mockLedgerContract{}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		ledgerContract,
		txRepo,
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	ctx := testContext()

	cmd := blockchain_in.JoinPrizePoolCommand{
		MatchID:        matchID,
		PlayerAddress:  testEVMAddress2(),
		PlayerWalletID: uuid.New(),
	}

	err := svc.JoinPrizePool(ctx, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prizePoolRepo.savedPool == nil {
		t.Error("pool was not saved")
	}
	if txRepo.savedTx == nil {
		t.Error("transaction was not saved")
	}
}

func TestJoinPrizePool_PoolNotFound(t *testing.T) {
	prizePoolRepo := &mockPrizePoolRepository{findErr: errors.New("not found")}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		&mockTransactionRepository{},
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	ctx := context.Background()
	cmd := blockchain_in.JoinPrizePoolCommand{
		MatchID:       uuid.New(),
		PlayerAddress: testEVMAddress(),
	}

	err := svc.JoinPrizePool(ctx, cmd)
	if err == nil {
		t.Fatal("expected error for pool not found")
	}
}

func TestJoinPrizePool_PoolNotAccumulating(t *testing.T) {
	matchID := uuid.New()
	existingPool := blockchain_entities.NewOnChainPrizePool(
		testResourceOwner(),
		matchID,
		blockchain_vo.PrimaryChain(),
		testEVMAddress(),
		testTokenAddress(),
		wallet_vo.CurrencyUSDC,
		wallet_vo.NewAmount(10),
		1000,
	)
	existingPool.Status = blockchain_entities.PoolStatusLocked

	prizePoolRepo := &mockPrizePoolRepository{foundPool: existingPool}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		&mockTransactionRepository{},
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	ctx := context.Background()
	cmd := blockchain_in.JoinPrizePoolCommand{
		MatchID:       matchID,
		PlayerAddress: testEVMAddress2(),
	}

	err := svc.JoinPrizePool(ctx, cmd)
	if err == nil {
		t.Fatal("expected error for pool not accepting entries")
	}
}

func TestLockPrizePool_Success(t *testing.T) {
	matchID := uuid.New()
	existingPool := blockchain_entities.NewOnChainPrizePool(
		testResourceOwner(),
		matchID,
		blockchain_vo.PrimaryChain(),
		testEVMAddress(),
		testTokenAddress(),
		wallet_vo.CurrencyUSDC,
		wallet_vo.NewAmount(10),
		1000,
	)
	existingPool.Status = blockchain_entities.PoolStatusAccumulating
	// Add participants
	existingPool.AddParticipant(testEVMAddress(), wallet_vo.NewAmount(10))
	existingPool.AddParticipant(testEVMAddress2(), wallet_vo.NewAmount(10))

	txHash := testTxHash()
	vaultContract := &mockVaultContract{lockPoolTxHash: txHash}
	prizePoolRepo := &mockPrizePoolRepository{foundPool: existingPool}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		&mockLedgerContract{},
		&mockTransactionRepository{},
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	err := svc.LockPrizePool(context.Background(), matchID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prizePoolRepo.savedPool == nil {
		t.Error("pool was not saved")
	}
	if prizePoolRepo.savedPool.Status != blockchain_entities.PoolStatusLocked {
		t.Errorf("expected status Locked, got %v", prizePoolRepo.savedPool.Status)
	}
}

func TestDepositToVault_Success(t *testing.T) {
	txHash := testTxHash()
	vaultContract := &mockVaultContract{depositTxHash: txHash}
	txRepo := &mockTransactionRepository{}
	ledgerContract := &mockLedgerContract{}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		ledgerContract,
		txRepo,
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	chainID := blockchain_vo.PrimaryChain()
	svc.SetVaultAddress(chainID, testEVMAddress())

	ctx := testContext()

	cmd := blockchain_in.DepositCommand{
		UserAddress:  testEVMAddress2(),
		WalletID:     uuid.New(),
		TokenAddress: testTokenAddress(),
		Currency:     wallet_vo.CurrencyUSDC,
		Amount:       wallet_vo.NewAmount(100),
		ChainID:      chainID,
	}

	tx, err := svc.DepositToVault(ctx, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
	if tx.Type != blockchain_entities.TxTypeDeposit {
		t.Errorf("expected type Deposit, got %v", tx.Type)
	}
	if txRepo.savedTx == nil {
		t.Error("transaction was not saved")
	}
}

func TestWithdrawFromVault_Success(t *testing.T) {
	txHash := testTxHash()
	vaultContract := &mockVaultContract{withdrawTxHash: txHash}
	txRepo := &mockTransactionRepository{}
	ledgerContract := &mockLedgerContract{}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		ledgerContract,
		txRepo,
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	chainID := blockchain_vo.PrimaryChain()
	svc.SetVaultAddress(chainID, testEVMAddress())

	ctx := testContext()

	cmd := blockchain_in.WithdrawCommand{
		UserAddress:  testEVMAddress2(),
		WalletID:     uuid.New(),
		TokenAddress: testTokenAddress(),
		Currency:     wallet_vo.CurrencyUSDC,
		Amount:       wallet_vo.NewAmount(50),
		ChainID:      chainID,
	}

	tx, err := svc.WithdrawFromVault(ctx, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
	if tx.Type != blockchain_entities.TxTypeWithdrawal {
		t.Errorf("expected type Withdrawal, got %v", tx.Type)
	}
}

func TestGetLedgerBalance_Success(t *testing.T) {
	expectedBalance := big.NewInt(1000000)
	ledgerContract := &mockLedgerContract{balance: expectedBalance}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		ledgerContract,
		&mockTransactionRepository{},
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	balance, err := svc.GetLedgerBalance(context.Background(),
		testEVMAddress(),
		testTokenAddress())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.Cmp(expectedBalance) != 0 {
		t.Errorf("expected balance %v, got %v", expectedBalance, balance)
	}
}

func TestGetPrizePool_Success(t *testing.T) {
	matchID := uuid.New()
	expectedPool := blockchain_entities.NewOnChainPrizePool(
		testResourceOwner(),
		matchID,
		blockchain_vo.PrimaryChain(),
		testEVMAddress(),
		testTokenAddress(),
		wallet_vo.CurrencyUSDC,
		wallet_vo.NewAmount(10),
		1000,
	)

	prizePoolRepo := &mockPrizePoolRepository{foundPool: expectedPool}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		&mockTransactionRepository{},
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	pool, err := svc.GetPrizePool(context.Background(), matchID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.MatchID != matchID {
		t.Errorf("expected match ID %v, got %v", matchID, pool.MatchID)
	}
}

func TestGetTransaction_Success(t *testing.T) {
	txID := uuid.New()
	expectedTx := blockchain_entities.NewBlockchainTransaction(
		testResourceOwner(),
		blockchain_vo.PrimaryChain(),
		blockchain_entities.TxTypeDeposit,
		testEVMAddress(),
		testEVMAddress2(),
		wallet_vo.CurrencyUSDC,
		wallet_vo.NewAmount(100),
	)
	expectedTx.BaseEntity.ID = txID

	txRepo := &mockTransactionRepository{foundTx: expectedTx}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		txRepo,
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	tx, err := svc.GetTransaction(context.Background(), txID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.BaseEntity.ID != txID {
		t.Errorf("expected tx ID %v, got %v", txID, tx.BaseEntity.ID)
	}
}

func TestGetTransactionsByWallet_Success(t *testing.T) {
	walletID := uuid.New()
	expectedTxs := []*blockchain_entities.BlockchainTransaction{
		blockchain_entities.NewBlockchainTransaction(
			testResourceOwner(),
			blockchain_vo.PrimaryChain(),
			blockchain_entities.TxTypeDeposit,
			testEVMAddress(),
			testEVMAddress2(),
			wallet_vo.CurrencyUSDC,
			wallet_vo.NewAmount(100),
		),
	}

	txRepo := &mockTransactionRepository{walletTxs: expectedTxs, walletTotal: 1}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		&mockVaultContract{},
		&mockLedgerContract{},
		txRepo,
		&mockPrizePoolRepository{},
		&mockSyncStateRepository{},
	)

	txs, total, err := svc.GetTransactionsByWallet(context.Background(), walletID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestCancelPrizePool_Success(t *testing.T) {
	matchID := uuid.New()
	existingPool := blockchain_entities.NewOnChainPrizePool(
		testResourceOwner(),
		matchID,
		blockchain_vo.PrimaryChain(),
		testEVMAddress(),
		testTokenAddress(),
		wallet_vo.CurrencyUSDC,
		wallet_vo.NewAmount(10),
		1000,
	)
	existingPool.Status = blockchain_entities.PoolStatusAccumulating
	existingPool.AddParticipant(testEVMAddress2(), wallet_vo.NewAmount(10))

	txHash := testTxHash()
	vaultContract := &mockVaultContract{cancelTxHash: txHash}
	prizePoolRepo := &mockPrizePoolRepository{foundPool: existingPool}
	ledgerContract := &mockLedgerContract{}

	svc := NewBlockchainService(
		&mockMultiChainClient{},
		vaultContract,
		ledgerContract,
		&mockTransactionRepository{},
		prizePoolRepo,
		&mockSyncStateRepository{},
	)

	err := svc.CancelPrizePool(context.Background(), matchID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prizePoolRepo.savedPool == nil {
		t.Error("pool was not saved")
	}
	if prizePoolRepo.savedPool.Status != blockchain_entities.PoolStatusCancelled {
		t.Errorf("expected status Cancelled, got %v", prizePoolRepo.savedPool.Status)
	}
}
