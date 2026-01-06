//go:build smoke || short
// +build smoke short

package smoke_test

import (
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// TestSmoke_WalletCreation tests wallet entity creation without external dependencies
func TestSmoke_WalletCreation(t *testing.T) {
	// Create test user
	userID := uuid.New()
	resourceOwner := shared.ResourceOwner{UserID: userID}

	// Create EVM address
	evmAddress, err := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	require.NoError(t, err, "Valid EVM address should be accepted")

	// Create wallet
	wallet, err := wallet_entities.NewUserWallet(resourceOwner, evmAddress)
	require.NoError(t, err, "Wallet creation should succeed")

	// Verify wallet properties
	assert.NotEqual(t, uuid.Nil, wallet.ID)
	assert.Equal(t, userID, wallet.ResourceOwner.UserID)
	assert.Equal(t, evmAddress.String(), wallet.EVMAddress.String())
	assert.False(t, wallet.IsLocked)

	// Verify default balances
	assert.Equal(t, int64(0), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())
	assert.Equal(t, int64(0), wallet.GetBalance(wallet_vo.CurrencyUSDC).ToCents())
	assert.Equal(t, int64(0), wallet.GetBalance(wallet_vo.CurrencyUSDT).ToCents())
}

// TestSmoke_WalletDeposit tests deposit logic without persistence
func TestSmoke_WalletDeposit(t *testing.T) {
	wallet := createTestWallet(t)

	// Deposit $100
	depositAmount := wallet_vo.NewAmount(100.00)
	err := wallet.Deposit(wallet_vo.CurrencyUSD, depositAmount)
	require.NoError(t, err)

	// Verify balance
	assert.Equal(t, int64(10000), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())
	assert.Equal(t, int64(10000), wallet.TotalDeposited.ToCents())
}

// TestSmoke_WalletWithdrawal tests withdrawal logic
func TestSmoke_WalletWithdrawal(t *testing.T) {
	wallet := createTestWallet(t)

	// Deposit $100 first
	wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(100.00))

	// Withdraw $30
	withdrawAmount := wallet_vo.NewAmount(30.00)
	err := wallet.Withdraw(wallet_vo.CurrencyUSD, withdrawAmount)
	require.NoError(t, err)

	// Verify balance
	assert.Equal(t, int64(7000), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())
	assert.Equal(t, int64(3000), wallet.TotalWithdrawn.ToCents())
}

// TestSmoke_WalletInsufficientBalance tests insufficient balance check
func TestSmoke_WalletInsufficientBalance(t *testing.T) {
	wallet := createTestWallet(t)

	// Try to withdraw without deposit
	withdrawAmount := wallet_vo.NewAmount(50.00)
	err := wallet.Withdraw(wallet_vo.CurrencyUSD, withdrawAmount)
	require.Error(t, err, "Withdrawal should fail with insufficient balance")
	assert.Contains(t, err.Error(), "insufficient balance")
}

// TestSmoke_WalletLock tests wallet locking
func TestSmoke_WalletLock(t *testing.T) {
	wallet := createTestWallet(t)

	// Deposit when unlocked
	wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(100.00))
	assert.Equal(t, int64(10000), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())

	// Lock wallet
	wallet.Lock("Fraud investigation")
	assert.True(t, wallet.IsLocked)
	assert.Equal(t, "Fraud investigation", wallet.LockReason)

	// Try to deposit when locked
	err := wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(50.00))
	require.Error(t, err, "Deposit should fail when wallet is locked")
	assert.Contains(t, err.Error(), "wallet is locked")

	// Unlock wallet
	wallet.Unlock()
	assert.False(t, wallet.IsLocked)
	assert.Equal(t, "", wallet.LockReason)

	// Deposit should work after unlock
	err = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(50.00))
	require.NoError(t, err)
}

// TestSmoke_WalletPrizeWithLimit tests prize addition with daily limit
func TestSmoke_WalletPrizeWithLimit(t *testing.T) {
	wallet := createTestWallet(t)

	maxDaily := wallet_vo.NewAmount(100.00)

	// Add first prize
	err := wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(60.00), maxDaily)
	require.NoError(t, err)
	assert.Equal(t, int64(6000), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())
	assert.Equal(t, int64(6000), wallet.DailyPrizeWinnings.ToCents())

	// Add second prize - should fail (exceeds daily limit)
	err = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(50.00), maxDaily)
	require.Error(t, err, "Prize should be rejected due to daily limit")
	assert.Contains(t, err.Error(), "daily prize limit exceeded")

	// Balance should not change
	assert.Equal(t, int64(6000), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())
}

// TestSmoke_LedgerEntry tests ledger entry creation
func TestSmoke_LedgerEntry(t *testing.T) {
	txID := uuid.New()
	walletID := uuid.New()
	userID := uuid.New()
	amount := big.NewFloat(100.00)

	entry := wallet_entities.NewLedgerEntry(
		txID,
		walletID,
		wallet_entities.AccountTypeAsset,
		wallet_entities.EntryTypeDebit,
		wallet_entities.AssetTypeFiat,
		amount,
		"Test deposit",
		"test_idempotency_key",
		userID,
	)

	// Verify entry properties
	assert.NotEqual(t, uuid.Nil, entry.ID)
	assert.Equal(t, txID, entry.TransactionID)
	assert.Equal(t, walletID, entry.AccountID)
	assert.Equal(t, wallet_entities.EntryTypeDebit, entry.EntryType)
	expectedAmount, _ := amount.Float64()
	actualAmount, _ := entry.Amount.Float64()
	assert.Equal(t, expectedAmount, actualAmount)
	assert.False(t, entry.IsReversed)
}

// TestSmoke_LedgerEntryReversal tests ledger entry reversal
func TestSmoke_LedgerEntryReversal(t *testing.T) {
	txID := uuid.New()
	walletID := uuid.New()
	userID := uuid.New()
	amount := big.NewFloat(50.00)

	// Create debit entry
	originalEntry := wallet_entities.NewLedgerEntry(
		txID,
		walletID,
		wallet_entities.AccountTypeAsset,
		wallet_entities.EntryTypeDebit,
		wallet_entities.AssetTypeFiat,
		amount,
		"Original transaction",
		"original_key",
		userID,
	)

	// Reverse it
	reversalEntry := originalEntry.Reverse("Test reversal", userID)

	// Verify reversal
	assert.NotEqual(t, originalEntry.TransactionID, reversalEntry.TransactionID, "Reversal should have new transaction ID")
	assert.Equal(t, originalEntry.AccountID, reversalEntry.AccountID)
	assert.Equal(t, wallet_entities.EntryTypeCredit, reversalEntry.EntryType, "Debit should be reversed to credit")
	originalAmount, _ := originalEntry.Amount.Float64()
	reversalAmount, _ := reversalEntry.Amount.Float64()
	assert.Equal(t, originalAmount, reversalAmount)
	assert.True(t, originalEntry.IsReversed, "Original should be marked as reversed")
	assert.Contains(t, reversalEntry.Description, "REVERSAL")
}

// TestSmoke_IdempotentOperation tests idempotent operation tracking
func TestSmoke_IdempotentOperation(t *testing.T) {
	key := "test_operation_123"
	opType := "Deposit"

	op := wallet_entities.NewIdempotentOperation(key, opType, nil)

	// Verify initial state
	assert.Equal(t, key, op.Key)
	assert.Equal(t, opType, op.OperationType)
	assert.Equal(t, wallet_entities.OperationStatusProcessing, op.Status)
	assert.Equal(t, 1, op.AttemptCount)
	assert.True(t, op.IsProcessing())
	assert.False(t, op.IsCompleted())

	// Mark as completed
	resultID := uuid.New()
	op.MarkCompleted(resultID, nil)

	assert.Equal(t, wallet_entities.OperationStatusCompleted, op.Status)
	assert.Equal(t, resultID, *op.ResultID)
	assert.False(t, op.IsProcessing())
	assert.True(t, op.IsCompleted())
}

// TestSmoke_CurrencyValidation tests currency validation
func TestSmoke_CurrencyValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid USD", "USD", false},
		{"Valid USDC", "USDC", false},
		{"Valid USDT", "USDT", false},
		{"Invalid currency", "INVALID", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := wallet_vo.ParseCurrency(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSmoke_EVMAddressValidation tests EVM address validation
func TestSmoke_EVMAddressValidation(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"Valid address", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", false},
		{"Valid lowercase", "0x742d35cc6634c0532925a3b844bc9e7595f0beb0", false},
		{"Invalid - too short", "0x742d35", true},
		{"Invalid - no 0x prefix", "742d35Cc6634C0532925a3b844Bc9e7595f0bEb", true},
		{"Invalid - not hex", "0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := wallet_vo.NewEVMAddress(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create a test wallet
func createTestWallet(t *testing.T) *wallet_entities.UserWallet {
	t.Helper()

	userID := uuid.New()
	resourceOwner := shared.ResourceOwner{UserID: userID}
	evmAddress, err := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	require.NoError(t, err)

	wallet, err := wallet_entities.NewUserWallet(resourceOwner, evmAddress)
	require.NoError(t, err)

	return wallet
}

// Benchmark tests
func BenchmarkWalletDeposit(b *testing.B) {
	userID := uuid.New()
	resourceOwner := shared.ResourceOwner{UserID: userID}
	evmAddress, _ := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	wallet, _ := wallet_entities.NewUserWallet(resourceOwner, evmAddress)
	amount := wallet_vo.NewAmount(10.00)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wallet.Deposit(wallet_vo.CurrencyUSD, amount)
	}
}

func BenchmarkLedgerEntryCreation(b *testing.B) {
	txID := uuid.New()
	walletID := uuid.New()
	userID := uuid.New()
	amount := big.NewFloat(100.00)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wallet_entities.NewLedgerEntry(
			txID,
			walletID,
			wallet_entities.AccountTypeAsset,
			wallet_entities.EntryTypeDebit,
			wallet_entities.AssetTypeFiat,
			amount,
			"Benchmark test",
			"bench_key",
			userID,
		)
	}
}
