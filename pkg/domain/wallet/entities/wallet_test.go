package wallet_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/assert"
)

// testWalletResourceOwner creates a test resource owner for wallet tests
func testWalletResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   uuid.New(),
	}
}

// validEVMAddress returns a valid EVM address for testing
func validEVMAddress() wallet_vo.EVMAddress {
	addr, _ := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	return addr
}

// =============================================================================
// TransactionStatus Constants Tests
// =============================================================================

func TestTransactionStatus_Constants(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(TransactionStatus("Pending"), TransactionStatusPending)
	assert.Equal(TransactionStatus("Completed"), TransactionStatusCompleted)
	assert.Equal(TransactionStatus("Failed"), TransactionStatusFailed)
	assert.Equal(TransactionStatus("RolledBack"), TransactionStatusRolledBack)
}

// =============================================================================
// NewUserWallet Tests
// =============================================================================

func TestNewUserWallet_CreatesValidWallet(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()
	evmAddress := validEVMAddress()

	wallet, err := NewUserWallet(rxn, evmAddress)

	assert.NoError(err)
	assert.NotNil(wallet)
	assert.Equal(evmAddress, wallet.EVMAddress)
	assert.NotEqual(uuid.Nil, wallet.ID)
	assert.False(wallet.IsLocked)
	assert.Empty(wallet.LockReason)
}

func TestNewUserWallet_InitializesBalances(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, err := NewUserWallet(rxn, validEVMAddress())

	assert.NoError(err)
	assert.NotNil(wallet.Balances)

	// Should initialize with zero balances for supported currencies
	assert.True(wallet.GetBalance(wallet_vo.CurrencyUSD).IsZero())
	assert.True(wallet.GetBalance(wallet_vo.CurrencyUSDC).IsZero())
	assert.True(wallet.GetBalance(wallet_vo.CurrencyUSDT).IsZero())
}

func TestNewUserWallet_InitializesTotals(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, err := NewUserWallet(rxn, validEVMAddress())

	assert.NoError(err)
	assert.True(wallet.TotalDeposited.IsZero())
	assert.True(wallet.TotalWithdrawn.IsZero())
	assert.True(wallet.TotalPrizesWon.IsZero())
	assert.True(wallet.DailyPrizeWinnings.IsZero())
	assert.Empty(wallet.PendingTransactions)
}

func TestNewUserWallet_FailsWithInvalidAddress(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()
	// Create an empty EVMAddress which is invalid
	invalidAddress := wallet_vo.EVMAddress{}

	wallet, err := NewUserWallet(rxn, invalidAddress)

	assert.Error(err)
	assert.Nil(wallet)
	assert.Contains(err.Error(), "invalid EVM address")
}

// =============================================================================
// GetBalance Tests
// =============================================================================

func TestGetBalance_ReturnsCorrectBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	wallet.Balances[wallet_vo.CurrencyUSD] = wallet_vo.NewAmount(10000) // $100.00

	balance := wallet.GetBalance(wallet_vo.CurrencyUSD)

	assert.Equal(wallet_vo.NewAmount(10000), balance)
}

func TestGetBalance_ReturnsZeroForUnknownCurrency(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())

	balance := wallet.GetBalance(wallet_vo.Currency("UNKNOWN"))

	assert.True(balance.IsZero())
}

// =============================================================================
// Deposit Tests
// =============================================================================

func TestDeposit_IncreasesBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	amount := wallet_vo.NewAmount(10000) // $100.00

	err := wallet.Deposit(wallet_vo.CurrencyUSD, amount)

	assert.NoError(err)
	assert.Equal(amount, wallet.GetBalance(wallet_vo.CurrencyUSD))
}

func TestDeposit_AccumulatesMultipleDeposits(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())

	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000))
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(3000))

	assert.Equal(wallet_vo.NewAmount(8000), wallet.GetBalance(wallet_vo.CurrencyUSD))
}

func TestDeposit_UpdatesTotalDeposited(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	amount := wallet_vo.NewAmount(10000)

	_ = wallet.Deposit(wallet_vo.CurrencyUSD, amount)

	assert.Equal(amount, wallet.TotalDeposited)
}

func TestDeposit_FailsWithNegativeAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	negativeAmount := wallet_vo.NewAmount(-100)

	err := wallet.Deposit(wallet_vo.CurrencyUSD, negativeAmount)

	assert.Error(err)
	assert.Contains(err.Error(), "positive")
}

func TestDeposit_FailsWithZeroAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())

	err := wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(0))

	assert.Error(err)
	assert.Contains(err.Error(), "positive")
}

func TestDeposit_FailsWhenLocked(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	wallet.Lock("Fraud investigation")

	err := wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(1000))

	assert.Error(err)
	assert.Contains(err.Error(), "locked")
	assert.Contains(err.Error(), "Fraud investigation")
}

// =============================================================================
// Withdraw Tests
// =============================================================================

func TestWithdraw_DecreasesBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))

	err := wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(3000))

	assert.NoError(err)
	assert.Equal(wallet_vo.NewAmount(7000), wallet.GetBalance(wallet_vo.CurrencyUSD))
}

func TestWithdraw_UpdatesTotalWithdrawn(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))

	_ = wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(3000))

	assert.Equal(wallet_vo.NewAmount(3000), wallet.TotalWithdrawn)
}

func TestWithdraw_FailsWithInsufficientBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000))

	err := wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))

	assert.Error(err)
	assert.Contains(err.Error(), "insufficient balance")
}

func TestWithdraw_FailsWithNegativeAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))

	err := wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(-100))

	assert.Error(err)
	assert.Contains(err.Error(), "positive")
}

func TestWithdraw_FailsWithZeroAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))

	err := wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(0))

	assert.Error(err)
	assert.Contains(err.Error(), "positive")
}

func TestWithdraw_FailsWhenLocked(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))
	wallet.Lock("Security review")

	err := wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000))

	assert.Error(err)
	assert.Contains(err.Error(), "locked")
}

// =============================================================================
// DeductEntryFee Tests
// =============================================================================

func TestDeductEntryFee_DeductsCorrectly(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000))

	err := wallet.DeductEntryFee(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(500))

	assert.NoError(err)
	assert.Equal(wallet_vo.NewAmount(4500), wallet.GetBalance(wallet_vo.CurrencyUSD))
}

func TestDeductEntryFee_FailsWithInsufficientBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(100))

	err := wallet.DeductEntryFee(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(500))

	assert.Error(err)
	assert.Contains(err.Error(), "entry fee")
}

// =============================================================================
// AddPrize Tests
// =============================================================================

func TestAddPrize_AddsPrizeToBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	maxDaily := wallet_vo.NewAmount(1000000) // $10,000 daily limit

	err := wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000), maxDaily)

	assert.NoError(err)
	assert.Equal(wallet_vo.NewAmount(10000), wallet.GetBalance(wallet_vo.CurrencyUSD))
	assert.Equal(wallet_vo.NewAmount(10000), wallet.TotalPrizesWon)
	assert.Equal(wallet_vo.NewAmount(10000), wallet.DailyPrizeWinnings)
}

func TestAddPrize_AccumulatesMultiplePrizes(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	maxDaily := wallet_vo.NewAmount(1000000)

	_ = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000), maxDaily)
	_ = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(3000), maxDaily)

	assert.Equal(wallet_vo.NewAmount(8000), wallet.GetBalance(wallet_vo.CurrencyUSD))
	assert.Equal(wallet_vo.NewAmount(8000), wallet.TotalPrizesWon)
	assert.Equal(wallet_vo.NewAmount(8000), wallet.DailyPrizeWinnings)
}

func TestAddPrize_FailsWhenDailyLimitExceeded(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	maxDaily := wallet_vo.NewAmount(10000) // $100 daily limit

	// First prize brings us to limit
	_ = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000), maxDaily)

	// Second prize should fail
	err := wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(1), maxDaily)

	assert.Error(err)
	assert.Contains(err.Error(), "daily prize limit exceeded")
}

func TestAddPrize_FailsWithNegativeAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	maxDaily := wallet_vo.NewAmount(1000000)

	err := wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(-100), maxDaily)

	assert.Error(err)
	assert.Contains(err.Error(), "positive")
}

func TestAddPrize_FailsWhenLocked(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	wallet.Lock("Investigation")
	maxDaily := wallet_vo.NewAmount(1000000)

	err := wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000), maxDaily)

	assert.Error(err)
	assert.Contains(err.Error(), "locked")
}

// =============================================================================
// Lock/Unlock Tests
// =============================================================================

func TestLock_LocksWallet(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())

	wallet.Lock("Suspicious activity detected")

	assert.True(wallet.IsLocked)
	assert.Equal("Suspicious activity detected", wallet.LockReason)
}

func TestUnlock_UnlocksWallet(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	wallet.Lock("Testing")

	wallet.Unlock()

	assert.False(wallet.IsLocked)
	assert.Empty(wallet.LockReason)
}

// =============================================================================
// Pending Transaction Tests
// =============================================================================

func TestAddPendingTransaction_AddsToList(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	txID := uuid.New()

	wallet.AddPendingTransaction(txID)

	assert.Len(wallet.PendingTransactions, 1)
	assert.Equal(txID, wallet.PendingTransactions[0])
}

func TestRemovePendingTransaction_RemovesFromList(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	tx1 := uuid.New()
	tx2 := uuid.New()
	wallet.AddPendingTransaction(tx1)
	wallet.AddPendingTransaction(tx2)

	wallet.RemovePendingTransaction(tx1)

	assert.Len(wallet.PendingTransactions, 1)
	assert.Equal(tx2, wallet.PendingTransactions[0])
}

func TestRemovePendingTransaction_NoOpIfNotFound(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	tx1 := uuid.New()
	tx2 := uuid.New()
	wallet.AddPendingTransaction(tx1)

	wallet.RemovePendingTransaction(tx2) // Not in list

	assert.Len(wallet.PendingTransactions, 1)
	assert.Equal(tx1, wallet.PendingTransactions[0])
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestValidate_SucceedsForValidWallet(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))

	err := wallet.Validate()

	assert.NoError(err)
}

func TestValidate_FailsWithInvalidEVMAddress(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	wallet.EVMAddress = wallet_vo.EVMAddress{} // Empty address is invalid

	err := wallet.Validate()

	assert.Error(err)
	assert.Contains(err.Error(), "invalid EVM address")
}

func TestValidate_FailsWithNegativeBalance(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	wallet.Balances[wallet_vo.CurrencyUSD] = wallet_vo.NewAmount(-100)

	err := wallet.Validate()

	assert.Error(err)
	assert.Contains(err.Error(), "negative balance")
}

// =============================================================================
// isSameDay Tests
// =============================================================================

func TestIsSameDay_ReturnsTrueForSameDay(t *testing.T) {
	assert := assert.New(t)

	t1 := time.Date(2024, 12, 19, 10, 30, 0, 0, time.UTC)
	t2 := time.Date(2024, 12, 19, 23, 59, 59, 0, time.UTC)

	assert.True(isSameDay(t1, t2))
}

func TestIsSameDay_ReturnsFalseForDifferentDay(t *testing.T) {
	assert := assert.New(t)

	t1 := time.Date(2024, 12, 19, 23, 59, 59, 0, time.UTC)
	t2 := time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC)

	assert.False(isSameDay(t1, t2))
}

// =============================================================================
// WalletTransaction Struct Tests
// =============================================================================

func TestWalletTransaction_StructFields(t *testing.T) {
	assert := assert.New(t)
	id := uuid.New()
	walletID := uuid.New()
	ledgerID := uuid.New()
	now := time.Now()

	tx := WalletTransaction{
		ID:           id,
		WalletID:     walletID,
		Type:         "Deposit",
		Status:       TransactionStatusCompleted,
		LedgerTxID:   &ledgerID,
		StartedAt:    now,
		CompletedAt:  &now,
		ErrorMessage: "",
		Metadata:     map[string]interface{}{"source": "bank_transfer"},
	}

	assert.Equal(id, tx.ID)
	assert.Equal(walletID, tx.WalletID)
	assert.Equal("Deposit", tx.Type)
	assert.Equal(TransactionStatusCompleted, tx.Status)
	assert.Equal(&ledgerID, tx.LedgerTxID)
	assert.Equal("bank_transfer", tx.Metadata["source"])
}

// =============================================================================
// Business Scenario Tests - E-Sports Platform Specific
// =============================================================================

func TestScenario_TournamentParticipation(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	// Player creates wallet and deposits funds
	wallet, err := NewUserWallet(rxn, validEVMAddress())
	assert.NoError(err)

	// Deposit $50
	err = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000))
	assert.NoError(err)
	assert.Equal(wallet_vo.NewAmount(5000), wallet.GetBalance(wallet_vo.CurrencyUSD))

	// Pay $5 entry fee
	err = wallet.DeductEntryFee(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(500))
	assert.NoError(err)
	assert.Equal(wallet_vo.NewAmount(4500), wallet.GetBalance(wallet_vo.CurrencyUSD))

	// Win $25 prize
	maxDaily := wallet_vo.NewAmount(100000) // $1000 daily limit
	err = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(2500), maxDaily)
	assert.NoError(err)
	assert.Equal(wallet_vo.NewAmount(7000), wallet.GetBalance(wallet_vo.CurrencyUSD))

	// Total prizes won should be $25
	assert.Equal(wallet_vo.NewAmount(2500), wallet.TotalPrizesWon)
}

func TestScenario_FraudPrevention(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(100000))

	// Anti-fraud: Lock wallet due to suspicious activity
	wallet.Lock("Multiple failed verification attempts")
	assert.True(wallet.IsLocked)

	// All operations should fail while locked
	err := wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(1000))
	assert.Error(err)
	assert.Contains(err.Error(), "locked")

	err = wallet.Withdraw(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(1000))
	assert.Error(err)
	assert.Contains(err.Error(), "locked")

	err = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(1000), wallet_vo.NewAmount(1000000))
	assert.Error(err)
	assert.Contains(err.Error(), "locked")

	// After investigation, unlock
	wallet.Unlock()
	assert.False(wallet.IsLocked)

	// Operations should work again
	err = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(1000))
	assert.NoError(err)
}

func TestScenario_DailyPrizeLimitReset(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())
	maxDaily := wallet_vo.NewAmount(10000) // $100 daily limit

	// Win max daily limit
	err := wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000), maxDaily)
	assert.NoError(err)

	// Next prize should fail
	err = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(100), maxDaily)
	assert.Error(err)
	assert.Contains(err.Error(), "daily prize limit")

	// Simulate next day by manipulating LastPrizeWinDate
	wallet.LastPrizeWinDate = time.Now().AddDate(0, 0, -1)

	// Now prize should work (daily counter resets)
	err = wallet.AddPrize(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(5000), maxDaily)
	assert.NoError(err)

	// Daily winnings should be reset to just the new prize
	assert.Equal(wallet_vo.NewAmount(5000), wallet.DailyPrizeWinnings)
}

func TestScenario_MultiCurrencySupport(t *testing.T) {
	assert := assert.New(t)
	rxn := testWalletResourceOwner()

	wallet, _ := NewUserWallet(rxn, validEVMAddress())

	// Deposit different currencies
	_ = wallet.Deposit(wallet_vo.CurrencyUSD, wallet_vo.NewAmount(10000))
	_ = wallet.Deposit(wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(5000))
	_ = wallet.Deposit(wallet_vo.CurrencyUSDT, wallet_vo.NewAmount(2500))

	// Verify each currency balance
	assert.Equal(wallet_vo.NewAmount(10000), wallet.GetBalance(wallet_vo.CurrencyUSD))
	assert.Equal(wallet_vo.NewAmount(5000), wallet.GetBalance(wallet_vo.CurrencyUSDC))
	assert.Equal(wallet_vo.NewAmount(2500), wallet.GetBalance(wallet_vo.CurrencyUSDT))

	// Withdraw from specific currency
	_ = wallet.Withdraw(wallet_vo.CurrencyUSDC, wallet_vo.NewAmount(3000))

	assert.Equal(wallet_vo.NewAmount(10000), wallet.GetBalance(wallet_vo.CurrencyUSD))
	assert.Equal(wallet_vo.NewAmount(2000), wallet.GetBalance(wallet_vo.CurrencyUSDC))
	assert.Equal(wallet_vo.NewAmount(2500), wallet.GetBalance(wallet_vo.CurrencyUSDT))
}

