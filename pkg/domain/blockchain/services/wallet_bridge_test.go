package blockchain_services

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	common "github.com/replay-api/replay-api/pkg/domain"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_in "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/in"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// Test helper functions to create value objects (handle errors in test context)

func mustNewEVMAddress(address string) wallet_vo.EVMAddress {
	addr, err := wallet_vo.NewEVMAddress(address)
	if err != nil {
		panic("invalid test EVM address: " + err.Error())
	}
	return addr
}

func mustNewTxHash(hash string) blockchain_vo.TxHash {
	txHash, err := blockchain_vo.NewTxHash(hash)
	if err != nil {
		panic("invalid test tx hash: " + err.Error())
	}
	return txHash
}

// Mock implementations

type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(ctx context.Context, cmd wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletService) Deposit(ctx context.Context, cmd wallet_in.DepositCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockWalletService) Withdraw(ctx context.Context, cmd wallet_in.WithdrawCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockWalletService) DeductEntryFee(ctx context.Context, cmd wallet_in.DeductEntryFeeCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockWalletService) AddPrize(ctx context.Context, cmd wallet_in.AddPrizeCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockWalletService) Refund(ctx context.Context, cmd wallet_in.RefundCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockWalletService) DebitWallet(ctx context.Context, cmd wallet_in.DebitWalletCommand) (*wallet_entities.WalletTransaction, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.WalletTransaction), args.Error(1)
}

func (m *MockWalletService) CreditWallet(ctx context.Context, cmd wallet_in.CreditWalletCommand) (*wallet_entities.WalletTransaction, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.WalletTransaction), args.Error(1)
}

type MockBlockchainService struct {
	mock.Mock
}

func (m *MockBlockchainService) CreatePrizePool(ctx context.Context, cmd blockchain_in.CreatePrizePoolCommand) (*blockchain_entities.OnChainPrizePool, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blockchain_entities.OnChainPrizePool), args.Error(1)
}

func (m *MockBlockchainService) JoinPrizePool(ctx context.Context, cmd blockchain_in.JoinPrizePoolCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockBlockchainService) LockPrizePool(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

func (m *MockBlockchainService) DistributePrizes(ctx context.Context, cmd blockchain_in.DistributePrizesCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockBlockchainService) CancelPrizePool(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

func (m *MockBlockchainService) DepositToVault(ctx context.Context, cmd blockchain_in.DepositCommand) (*blockchain_entities.BlockchainTransaction, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blockchain_entities.BlockchainTransaction), args.Error(1)
}

func (m *MockBlockchainService) WithdrawFromVault(ctx context.Context, cmd blockchain_in.WithdrawCommand) (*blockchain_entities.BlockchainTransaction, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blockchain_entities.BlockchainTransaction), args.Error(1)
}

func (m *MockBlockchainService) RecordTransaction(ctx context.Context, cmd blockchain_in.RecordLedgerEntryCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockBlockchainService) GetLedgerBalance(ctx context.Context, account wallet_vo.EVMAddress, token wallet_vo.EVMAddress) (*big.Int, error) {
	args := m.Called(ctx, account, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBlockchainService) SyncPrizePool(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

func (m *MockBlockchainService) SyncAllPendingPools(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBlockchainService) GetPrizePool(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blockchain_entities.OnChainPrizePool), args.Error(1)
}

func (m *MockBlockchainService) GetTransaction(ctx context.Context, txID uuid.UUID) (*blockchain_entities.BlockchainTransaction, error) {
	args := m.Called(ctx, txID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blockchain_entities.BlockchainTransaction), args.Error(1)
}

func (m *MockBlockchainService) GetTransactionsByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error) {
	args := m.Called(ctx, walletID, limit, offset)
	return args.Get(0).([]*blockchain_entities.BlockchainTransaction), args.Get(1).(int64), args.Error(2)
}

// Test helpers

func createTestBridge() (*WalletBlockchainBridge, *MockWalletService, *MockBlockchainService) {
	walletService := new(MockWalletService)
	blockchainService := new(MockBlockchainService)
	bridge := NewWalletBlockchainBridge(walletService, blockchainService)
	return bridge, walletService, blockchainService
}

func createTestDepositCommand() DepositWithBlockchainCommand {
	return DepositWithBlockchainCommand{
		UserID:       uuid.New(),
		WalletID:     uuid.New(),
		EVMAddress:   mustNewEVMAddress("0x1234567890123456789012345678901234567890"),
		TokenAddress: mustNewEVMAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
		Currency:     wallet_vo.CurrencyUSD,
		Amount:       wallet_vo.NewAmount(100), // $100
		PaymentID:    uuid.New(),
		ChainID:      blockchain_vo.ChainIDEthereum,
	}
}

func createTestWithdrawCommand() WithdrawWithBlockchainCommand {
	return WithdrawWithBlockchainCommand{
		UserID:       uuid.New(),
		WalletID:     uuid.New(),
		EVMAddress:   mustNewEVMAddress("0x1234567890123456789012345678901234567890"),
		TokenAddress: mustNewEVMAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		Currency:     wallet_vo.CurrencyUSD,
		Amount:       wallet_vo.NewAmount(50), // $50
		ChainID:      blockchain_vo.ChainIDEthereum,
	}
}

func createTestEntryFeeCommand() DeductEntryFeeWithBlockchainCommand {
	return DeductEntryFeeWithBlockchainCommand{
		UserID:     uuid.New(),
		WalletID:   uuid.New(),
		EVMAddress: mustNewEVMAddress("0x1234567890123456789012345678901234567890"),
		MatchID:    uuid.New(),
		Currency:   wallet_vo.CurrencyUSD,
		Amount:     wallet_vo.NewAmount(10), // $10 entry fee
	}
}

func createTestPrizeCommand() AddPrizeWithBlockchainCommand {
	return AddPrizeWithBlockchainCommand{
		UserID:     uuid.New(),
		WalletID:   uuid.New(),
		EVMAddress: mustNewEVMAddress("0x1234567890123456789012345678901234567890"),
		MatchID:    uuid.New(),
		Currency:   wallet_vo.CurrencyUSD,
		Amount:     wallet_vo.NewAmount(500), // $500 prize
		Rank:       1,
	}
}

// Tests

func TestNewWalletBlockchainBridge(t *testing.T) {
	bridge, walletService, blockchainService := createTestBridge()

	assert.NotNil(t, bridge)
	assert.Equal(t, walletService, bridge.walletService)
	assert.Equal(t, blockchainService, bridge.blockchainService)
	assert.True(t, bridge.syncToBlockchain)
	assert.True(t, bridge.verifyFromChain)
	assert.False(t, bridge.useChainAsSource)
}

func TestEnableChainAsSource(t *testing.T) {
	bridge, _, _ := createTestBridge()

	assert.False(t, bridge.useChainAsSource)
	bridge.EnableChainAsSource()
	assert.True(t, bridge.useChainAsSource)
}

// DepositWithBlockchain Tests

func TestDepositWithBlockchain_Success_BothSucceed(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, blockchainService := createTestBridge()
	cmd := createTestDepositCommand()

	txHash := mustNewTxHash("0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc1")
	mockTx := &blockchain_entities.BlockchainTransaction{
		TxHash: txHash,
	}

	// Setup expectations
	walletService.On("Deposit", ctx, mock.MatchedBy(func(c wallet_in.DepositCommand) bool {
		return c.UserID == cmd.UserID && c.Amount == cmd.Amount.Dollars()
	})).Return(nil)

	blockchainService.On("DepositToVault", ctx, mock.MatchedBy(func(c blockchain_in.DepositCommand) bool {
		return c.WalletID == cmd.WalletID && c.Amount.Equals(cmd.Amount)
	})).Return(mockTx, nil)

	// Execute
	result, err := bridge.DepositWithBlockchain(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.OffChainSuccess)
	assert.True(t, result.OnChainSuccess)
	assert.Nil(t, result.OffChainError)
	assert.Nil(t, result.OnChainError)
	assert.Equal(t, &txHash, result.TxHash)
	assert.True(t, result.IsFullySuccessful())
	assert.False(t, result.NeedsReconciliation())
	assert.Equal(t, "Deposit", result.Operation)

	walletService.AssertExpectations(t)
	blockchainService.AssertExpectations(t)
}

func TestDepositWithBlockchain_OffChainFails(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, _ := createTestBridge()
	cmd := createTestDepositCommand()

	expectedErr := errors.New("insufficient funds")
	walletService.On("Deposit", ctx, mock.Anything).Return(expectedErr)

	// Execute
	result, err := bridge.DepositWithBlockchain(ctx, cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "off-chain deposit failed")
	assert.NotNil(t, result)
	assert.False(t, result.OffChainSuccess)
	assert.False(t, result.OnChainSuccess)
	assert.Equal(t, expectedErr, result.OffChainError)

	walletService.AssertExpectations(t)
}

func TestDepositWithBlockchain_OnChainFails_OffChainSucceeds(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, blockchainService := createTestBridge()
	cmd := createTestDepositCommand()

	onChainErr := errors.New("blockchain network error")

	walletService.On("Deposit", ctx, mock.Anything).Return(nil)
	blockchainService.On("DepositToVault", ctx, mock.Anything).Return(nil, onChainErr)

	// Execute
	result, err := bridge.DepositWithBlockchain(ctx, cmd)

	// Assert - should NOT return error because off-chain succeeded
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.OffChainSuccess)
	assert.False(t, result.OnChainSuccess)
	assert.Nil(t, result.OffChainError)
	assert.Equal(t, onChainErr, result.OnChainError)
	assert.False(t, result.IsFullySuccessful())
	assert.True(t, result.NeedsReconciliation())

	walletService.AssertExpectations(t)
	blockchainService.AssertExpectations(t)
}

func TestDepositWithBlockchain_SyncDisabled(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, _ := createTestBridge()
	bridge.syncToBlockchain = false // Disable blockchain sync
	cmd := createTestDepositCommand()

	walletService.On("Deposit", ctx, mock.Anything).Return(nil)

	// Execute
	result, err := bridge.DepositWithBlockchain(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.OffChainSuccess)
	assert.False(t, result.OnChainSuccess) // Was not attempted
	assert.Nil(t, result.TxHash)

	walletService.AssertExpectations(t)
}

// WithdrawWithBlockchain Tests

func TestWithdrawWithBlockchain_Success_BothSucceed(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, blockchainService := createTestBridge()
	cmd := createTestWithdrawCommand()

	txHash := mustNewTxHash("0xdef456def456def456def456def456def456def456def456def456def456def4")
	mockTx := &blockchain_entities.BlockchainTransaction{
		TxHash: txHash,
	}

	blockchainService.On("WithdrawFromVault", ctx, mock.MatchedBy(func(c blockchain_in.WithdrawCommand) bool {
		return c.WalletID == cmd.WalletID
	})).Return(mockTx, nil)

	walletService.On("Withdraw", ctx, mock.MatchedBy(func(c wallet_in.WithdrawCommand) bool {
		return c.UserID == cmd.UserID && c.Amount == cmd.Amount.Dollars()
	})).Return(nil)

	// Execute
	result, err := bridge.WithdrawWithBlockchain(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.OffChainSuccess)
	assert.True(t, result.OnChainSuccess)
	assert.Equal(t, &txHash, result.TxHash)
	assert.True(t, result.IsFullySuccessful())
	assert.Equal(t, "Withdrawal", result.Operation)

	walletService.AssertExpectations(t)
	blockchainService.AssertExpectations(t)
}

func TestWithdrawWithBlockchain_OnChainFails(t *testing.T) {
	ctx := context.Background()
	bridge, _, blockchainService := createTestBridge()
	cmd := createTestWithdrawCommand()

	onChainErr := errors.New("insufficient vault balance")
	blockchainService.On("WithdrawFromVault", ctx, mock.Anything).Return(nil, onChainErr)

	// Execute
	result, err := bridge.WithdrawWithBlockchain(ctx, cmd)

	// Assert - should fail because on-chain is required for withdrawals
	require.Error(t, err)
	assert.Contains(t, err.Error(), "on-chain withdrawal failed")
	assert.False(t, result.OnChainSuccess)
	assert.Equal(t, onChainErr, result.OnChainError)

	blockchainService.AssertExpectations(t)
}

func TestWithdrawWithBlockchain_OnChainSucceeds_OffChainFails(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, blockchainService := createTestBridge()
	cmd := createTestWithdrawCommand()

	txHash := mustNewTxHash("0xabc789abc789abc789abc789abc789abc789abc789abc789abc789abc789abc7")
	mockTx := &blockchain_entities.BlockchainTransaction{
		TxHash: txHash,
	}

	offChainErr := errors.New("database connection failed")

	blockchainService.On("WithdrawFromVault", ctx, mock.Anything).Return(mockTx, nil)
	walletService.On("Withdraw", ctx, mock.Anything).Return(offChainErr)

	// Execute
	result, err := bridge.WithdrawWithBlockchain(ctx, cmd)

	// Assert - no error returned but reconciliation needed
	require.NoError(t, err)
	assert.True(t, result.OnChainSuccess)
	assert.False(t, result.OffChainSuccess)
	assert.Equal(t, offChainErr, result.OffChainError)
	assert.True(t, result.NeedsReconciliation())
	assert.Equal(t, &txHash, result.TxHash)

	walletService.AssertExpectations(t)
	blockchainService.AssertExpectations(t)
}

// DeductEntryFeeWithBlockchain Tests

func TestDeductEntryFeeWithBlockchain_Success(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, blockchainService := createTestBridge()
	cmd := createTestEntryFeeCommand()

	blockchainService.On("JoinPrizePool", ctx, mock.MatchedBy(func(c blockchain_in.JoinPrizePoolCommand) bool {
		return c.MatchID == cmd.MatchID && c.PlayerWalletID == cmd.WalletID
	})).Return(nil)

	walletService.On("DeductEntryFee", ctx, mock.MatchedBy(func(c wallet_in.DeductEntryFeeCommand) bool {
		return c.UserID == cmd.UserID && c.Amount == cmd.Amount.Dollars()
	})).Return(nil)

	// Execute
	result, err := bridge.DeductEntryFeeWithBlockchain(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.OnChainSuccess)
	assert.True(t, result.OffChainSuccess)
	assert.Equal(t, "EntryFee", result.Operation)

	walletService.AssertExpectations(t)
	blockchainService.AssertExpectations(t)
}

func TestDeductEntryFeeWithBlockchain_OnChainFails(t *testing.T) {
	ctx := context.Background()
	bridge, _, blockchainService := createTestBridge()
	cmd := createTestEntryFeeCommand()

	onChainErr := errors.New("prize pool is locked")
	blockchainService.On("JoinPrizePool", ctx, mock.Anything).Return(onChainErr)

	// Execute
	result, err := bridge.DeductEntryFeeWithBlockchain(ctx, cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "on-chain entry fee failed")
	assert.Equal(t, onChainErr, result.OnChainError)

	blockchainService.AssertExpectations(t)
}

func TestDeductEntryFeeWithBlockchain_OffChainFails(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, blockchainService := createTestBridge()
	cmd := createTestEntryFeeCommand()

	offChainErr := errors.New("insufficient wallet balance")

	blockchainService.On("JoinPrizePool", ctx, mock.Anything).Return(nil)
	walletService.On("DeductEntryFee", ctx, mock.Anything).Return(offChainErr)

	// Execute
	result, err := bridge.DeductEntryFeeWithBlockchain(ctx, cmd)

	// Assert - no error returned but needs reconciliation
	require.NoError(t, err)
	assert.True(t, result.OnChainSuccess)
	assert.False(t, result.OffChainSuccess)
	assert.Equal(t, offChainErr, result.OffChainError)
	assert.True(t, result.NeedsReconciliation())

	walletService.AssertExpectations(t)
	blockchainService.AssertExpectations(t)
}

// AddPrizeWithBlockchain Tests

func TestAddPrizeWithBlockchain_Success(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, _ := createTestBridge()
	cmd := createTestPrizeCommand()

	walletService.On("AddPrize", ctx, mock.MatchedBy(func(c wallet_in.AddPrizeCommand) bool {
		return c.UserID == cmd.UserID && c.Amount == cmd.Amount.Dollars()
	})).Return(nil)

	// Execute
	result, err := bridge.AddPrizeWithBlockchain(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.OffChainSuccess)
	assert.Equal(t, "Prize", result.Operation)

	walletService.AssertExpectations(t)
}

func TestAddPrizeWithBlockchain_OffChainFails(t *testing.T) {
	ctx := context.Background()
	bridge, walletService, _ := createTestBridge()
	cmd := createTestPrizeCommand()

	offChainErr := errors.New("wallet not found")
	walletService.On("AddPrize", ctx, mock.Anything).Return(offChainErr)

	// Execute
	result, err := bridge.AddPrizeWithBlockchain(ctx, cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "off-chain prize add failed")
	assert.False(t, result.OffChainSuccess)
	assert.Equal(t, offChainErr, result.OffChainError)

	walletService.AssertExpectations(t)
}

// ReconcileWallet Tests

func TestReconcileWallet_Success(t *testing.T) {
	ctx := context.Background()
	bridge, _, blockchainService := createTestBridge()
	walletID := uuid.New()
	evmAddress := mustNewEVMAddress("0x1234567890123456789012345678901234567890")
	tokenAddress := mustNewEVMAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	expectedBalance := big.NewInt(1000000) // 1 USDC (6 decimals)

	blockchainService.On("GetLedgerBalance", ctx, evmAddress, tokenAddress).Return(expectedBalance, nil)

	// Execute
	result, err := bridge.ReconcileWallet(ctx, walletID, evmAddress, tokenAddress)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, walletID, result.WalletID)
	assert.Equal(t, expectedBalance, result.OnChainBalance)
	assert.True(t, result.IsReconciled)

	blockchainService.AssertExpectations(t)
}

func TestReconcileWallet_GetBalanceFails(t *testing.T) {
	ctx := context.Background()
	bridge, _, blockchainService := createTestBridge()
	walletID := uuid.New()
	evmAddress := mustNewEVMAddress("0x1234567890123456789012345678901234567890")
	tokenAddress := mustNewEVMAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	expectedErr := errors.New("blockchain RPC error")
	blockchainService.On("GetLedgerBalance", ctx, evmAddress, tokenAddress).Return(nil, expectedErr)

	// Execute
	result, err := bridge.ReconcileWallet(ctx, walletID, evmAddress, tokenAddress)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get on-chain balance")
	assert.Nil(t, result)

	blockchainService.AssertExpectations(t)
}

// VerifyTransactionOnChain Tests

func TestVerifyTransactionOnChain(t *testing.T) {
	ctx := context.Background()
	bridge, _, _ := createTestBridge()
	txHash := mustNewTxHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")
	chainID := blockchain_vo.ChainIDEthereum

	// Execute
	verified, err := bridge.VerifyTransactionOnChain(ctx, txHash, chainID)

	// Assert - placeholder implementation returns true
	require.NoError(t, err)
	assert.True(t, verified)
}

// BlockchainSyncResult Tests

func TestBlockchainSyncResult_IsFullySuccessful(t *testing.T) {
	tests := []struct {
		name     string
		result   BlockchainSyncResult
		expected bool
	}{
		{
			name: "both succeed",
			result: BlockchainSyncResult{
				OffChainSuccess: true,
				OnChainSuccess:  true,
			},
			expected: true,
		},
		{
			name: "off-chain only",
			result: BlockchainSyncResult{
				OffChainSuccess: true,
				OnChainSuccess:  false,
			},
			expected: false,
		},
		{
			name: "on-chain only",
			result: BlockchainSyncResult{
				OffChainSuccess: false,
				OnChainSuccess:  true,
			},
			expected: false,
		},
		{
			name: "both fail",
			result: BlockchainSyncResult{
				OffChainSuccess: false,
				OnChainSuccess:  false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsFullySuccessful())
		})
	}
}

func TestBlockchainSyncResult_NeedsReconciliation(t *testing.T) {
	tests := []struct {
		name     string
		result   BlockchainSyncResult
		expected bool
	}{
		{
			name: "both succeed - no reconciliation",
			result: BlockchainSyncResult{
				OffChainSuccess: true,
				OnChainSuccess:  true,
			},
			expected: false,
		},
		{
			name: "both fail - no reconciliation",
			result: BlockchainSyncResult{
				OffChainSuccess: false,
				OnChainSuccess:  false,
			},
			expected: false,
		},
		{
			name: "off-chain only - needs reconciliation",
			result: BlockchainSyncResult{
				OffChainSuccess: true,
				OnChainSuccess:  false,
			},
			expected: true,
		},
		{
			name: "on-chain only - needs reconciliation",
			result: BlockchainSyncResult{
				OffChainSuccess: false,
				OnChainSuccess:  true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.NeedsReconciliation())
		})
	}
}

// BlockchainWalletEventHandler Tests

func TestNewBlockchainWalletEventHandler(t *testing.T) {
	walletService := new(MockWalletService)
	handler := NewBlockchainWalletEventHandler(walletService)

	assert.NotNil(t, handler)
	assert.Equal(t, walletService, handler.walletService)
}

func TestBlockchainWalletEventHandler_OnPrizeDistributed(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)
	handler := NewBlockchainWalletEventHandler(walletService)

	event := blockchain_in.PrizeDistributedEvent{
		MatchID:     [32]byte{1, 2, 3},
		Winner:      mustNewEVMAddress("0x1234567890123456789012345678901234567890"),
		Amount:      big.NewInt(1000000),
		Rank:        1,
		TxHash:      mustNewTxHash("0xabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc1"),
		BlockNumber: 12345678,
	}

	// Execute
	err := handler.OnPrizeDistributed(ctx, event)

	// Assert - placeholder returns nil
	require.NoError(t, err)
}

func TestBlockchainWalletEventHandler_OnUserWithdrawal(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)
	handler := NewBlockchainWalletEventHandler(walletService)

	event := blockchain_in.UserWithdrawalEvent{
		User:        mustNewEVMAddress("0x1234567890123456789012345678901234567890"),
		Token:       mustNewEVMAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		Amount:      big.NewInt(500000),
		TxHash:      mustNewTxHash("0xdef456def456def456def456def456def456def456def456def456def456def4"),
		BlockNumber: 12345679,
	}

	// Execute
	err := handler.OnUserWithdrawal(ctx, event)

	// Assert - placeholder returns nil
	require.NoError(t, err)
}

// WalletTxToBlockchainTx Tests

func TestWalletTxToBlockchainTx(t *testing.T) {
	tests := []struct {
		name       string
		txType     string
		expected   blockchain_entities.TransactionType
	}{
		{"deposit", "Deposit", blockchain_entities.TxTypeDeposit},
		{"withdrawal", "Withdrawal", blockchain_entities.TxTypeWithdrawal},
		{"debit", "Debit", blockchain_entities.TxTypeEntryFee},
		{"credit", "Credit", blockchain_entities.TxTypePrize},
		{"unknown", "Unknown", blockchain_entities.TxTypeContractCall},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walletTx := &wallet_entities.WalletTransaction{
				Type: tt.txType,
			}
			chainID := blockchain_vo.ChainIDEthereum
			from := mustNewEVMAddress("0x1111111111111111111111111111111111111111")
			to := mustNewEVMAddress("0x2222222222222222222222222222222222222222")
			currency := wallet_vo.CurrencyUSD
			amount := wallet_vo.NewAmount(100)
			resourceOwner := bridgeTestResourceOwner()

			result := WalletTxToBlockchainTx(walletTx, chainID, from, to, currency, amount, resourceOwner)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, result.Type)
			assert.Equal(t, chainID, result.ChainID)
			assert.Equal(t, from, result.FromAddress)
			assert.Equal(t, to, result.ToAddress)
		})
	}
}

// Helper to create test resource owner for wallet bridge tests
func bridgeTestResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		ClientID: uuid.New(),
	}
}

