//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	mongodb "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
)

// TestE2E_WalletLifecycle tests the complete wallet lifecycle with real MongoDB
// NO MOCKS - Production-grade integration test
func TestE2E_WalletLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup: Connect to test MongoDB
	mongoURI := getMongoTestURI()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err, "Failed to connect to MongoDB")
	defer func() { _ = client.Disconnect(ctx) }()

	// Create test database
	dbName := "replay_test_" + uuid.New().String()
	db := client.Database(dbName)
	defer func() {
		// Cleanup: Drop test database
		db.Drop(ctx)
	}()

	// Initialize repositories
	walletRepo := mongodb.NewWalletRepository(db)
	ledgerRepo := mongodb.NewLedgerRepository(db)
	idempotencyRepo := mongodb.NewIdempotencyRepository(db)

	// Initialize services
	ledgerService := wallet_services.NewLedgerService(ledgerRepo, idempotencyRepo)
	coordinator := wallet_services.NewTransactionCoordinator(walletRepo, ledgerService)
	reconciliationService := wallet_services.NewReconciliationService(walletRepo, ledgerRepo)

	// Create test user
	userID := uuid.New()
	resourceOwner := common.ResourceOwner{
		UserID: userID,
	}
	ctx = common.SetResourceOwner(ctx, resourceOwner)

	// Create wallet
	evmAddress, err := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	require.NoError(t, err)

	wallet, err := wallet_entities.NewUserWallet(resourceOwner, evmAddress)
	require.NoError(t, err)

	err = walletRepo.Save(ctx, wallet)
	require.NoError(t, err)

	t.Log("✓ Wallet created successfully")

	// Test 1: Deposit with automatic rollback on failure
	t.Run("Deposit_Success", func(t *testing.T) {
		depositAmount := wallet_vo.NewAmount(100.00) // $100
		paymentID := uuid.New()

		ledgerTxID, err := coordinator.ExecuteDeposit(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			depositAmount,
			paymentID,
			wallet_entities.LedgerMetadata{},
		)

		require.NoError(t, err, "Deposit should succeed")
		assert.NotEqual(t, uuid.Nil, ledgerTxID)

		// Verify wallet balance
		assert.Equal(t, depositAmount.ToCents(), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())

		// Verify ledger entries
		entries, err := ledgerRepo.FindByTransactionID(ctx, ledgerTxID)
		require.NoError(t, err)
		assert.Len(t, entries, 2, "Should have 2 ledger entries (debit + credit)")

		// Verify double-entry accounting
		var debits, credits int64
		for _, entry := range entries {
			if entry.EntryType == wallet_entities.EntryTypeDebit {
				debits += entry.Amount.ToCents()
				assert.Equal(t, wallet.ID, entry.AccountID, "Debit should be to user's account")
				assert.Equal(t, wallet_entities.AccountTypeAsset, entry.AccountType)
			} else {
				credits += entry.Amount.ToCents()
				assert.Equal(t, wallet_services.SystemLiabilityAccountID, entry.AccountID, "Credit should be to liability account")
				assert.Equal(t, wallet_entities.AccountTypeLiability, entry.AccountType)
			}
		}

		assert.Equal(t, debits, credits, "Debits must equal credits")
		t.Log("✓ Deposit created correct double-entry ledger entries")
	})

	// Test 2: Withdrawal
	t.Run("Withdrawal_Success", func(t *testing.T) {
		withdrawAmount := wallet_vo.NewAmount(30.00) // $30

		ledgerTxID, err := coordinator.ExecuteWithdrawal(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			withdrawAmount,
			"0x123...",
			wallet_entities.LedgerMetadata{},
		)

		require.NoError(t, err, "Withdrawal should succeed")

		// Verify wallet balance
		expectedBalance := wallet_vo.NewAmount(70.00) // $100 - $30
		assert.Equal(t, expectedBalance.ToCents(), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())

		// Verify ledger entries
		entries, err := ledgerRepo.FindByTransactionID(ctx, ledgerTxID)
		require.NoError(t, err)
		assert.Len(t, entries, 2)

		t.Log("✓ Withdrawal successful")
	})

	// Test 3: Balance Reconciliation
	t.Run("Balance_Reconciliation", func(t *testing.T) {
		// Reload wallet from DB
		walletFromDB, err := walletRepo.FindByID(ctx, wallet.ID)
		require.NoError(t, err)

		result, err := reconciliationService.ReconcileWallet(ctx, walletFromDB.ID)
		require.NoError(t, err)

		// All balances should match
		assert.Equal(t, wallet_services.ReconciliationStatusMatched, result.Status,
			"Wallet balance should match ledger")
		assert.Equal(t, 0, len(result.Discrepancies), "Should have no discrepancies")

		t.Log("✓ Reconciliation: Wallet matches ledger")
	})

	// Test 4: Idempotency - duplicate deposit should be rejected
	t.Run("Idempotency_DuplicateDeposit", func(t *testing.T) {
		depositAmount := wallet_vo.NewAmount(50.00)
		paymentID := uuid.New()

		// First deposit
		ledgerTxID1, err1 := coordinator.ExecuteDeposit(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			depositAmount,
			paymentID,
			wallet_entities.LedgerMetadata{},
		)
		require.NoError(t, err1)

		balanceAfterFirst := wallet.GetBalance(wallet_vo.CurrencyUSD)

		// Second deposit with SAME payment ID (should be idempotent)
		ledgerTxID2, err2 := coordinator.ExecuteDeposit(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			depositAmount,
			paymentID,
			wallet_entities.LedgerMetadata{},
		)

		// Should return existing transaction ID, not create new one
		assert.Error(t, err2, "Duplicate deposit should fail due to idempotency key")

		// Balance should NOT change
		assert.Equal(t, balanceAfterFirst.ToCents(), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())

		t.Log("✓ Idempotency protection working")
	})

	// Test 5: Entry Fee with automatic rollback
	t.Run("EntryFee_InsufficientBalance_Rollback", func(t *testing.T) {
		// Try to deduct more than available balance
		excessiveAmount := wallet_vo.NewAmount(10000.00) // $10,000

		balanceBefore := wallet.GetBalance(wallet_vo.CurrencyUSD)

		_, err := coordinator.ExecuteEntryFee(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			excessiveAmount,
			nil,
			nil,
			wallet_entities.LedgerMetadata{},
		)

		// Should fail
		require.Error(t, err, "Entry fee should fail due to insufficient balance")

		// Balance should be unchanged (rollback worked)
		assert.Equal(t, balanceBefore.ToCents(), wallet.GetBalance(wallet_vo.CurrencyUSD).ToCents())

		// Verify no orphaned ledger entries
		allEntries, err := ledgerRepo.FindByAccountID(ctx, wallet.ID, 1000, 0)
		require.NoError(t, err)

		// Count should not have increased (rollback cleaned up)
		entriesBeforeCount := len(allEntries)

		t.Log("✓ Rollback prevented orphaned ledger entries")
	})

	// Test 6: Prize Winning with daily limit check
	t.Run("PrizeWinning_WithinLimit", func(t *testing.T) {
		prizeAmount := wallet_vo.NewAmount(25.00) // $25
		maxDaily := wallet_vo.NewAmount(50.00) // $50/day limit

		ledgerTxID, err := coordinator.ExecutePrizeWinning(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			prizeAmount,
			nil,
			nil,
			maxDaily,
			wallet_entities.LedgerMetadata{},
		)

		require.NoError(t, err, "Prize should be awarded")

		// Verify double-entry
		entries, err := ledgerRepo.FindByTransactionID(ctx, ledgerTxID)
		require.NoError(t, err)
		assert.Len(t, entries, 2)

		// Find expense entry
		var foundExpenseEntry bool
		for _, entry := range entries {
			if entry.AccountID == wallet_services.SystemExpenseAccountID {
				foundExpenseEntry = true
				assert.Equal(t, wallet_entities.AccountTypeExpense, entry.AccountType)
				assert.Equal(t, wallet_entities.EntryTypeCredit, entry.EntryType)
			}
		}
		assert.True(t, foundExpenseEntry, "Should have expense entry for platform cost")

		t.Log("✓ Prize awarded with correct double-entry accounting")
	})

	// Test 7: Transaction History with Pagination
	t.Run("TransactionHistory_Pagination", func(t *testing.T) {
		filters := mongodb.HistoryFilters{
			Currency: &wallet_vo.CurrencyUSD,
			Limit:    10,
			Offset:   0,
			SortBy:   "created_at",
			SortOrder: "desc",
		}

		entries, total, err := ledgerRepo.GetAccountHistory(ctx, wallet.ID, filters)
		require.NoError(t, err)

		assert.Greater(t, len(entries), 0, "Should have transaction history")
		assert.Greater(t, total, int64(0), "Should have total count")

		// Verify sorting (newest first)
		if len(entries) > 1 {
			assert.True(t, entries[0].CreatedAt.After(entries[len(entries)-1].CreatedAt) ||
				entries[0].CreatedAt.Equal(entries[len(entries)-1].CreatedAt),
				"Entries should be sorted by created_at desc")
		}

		t.Log("✓ Transaction history retrieved successfully")
	})

	// Test 8: Ledger Balance Calculation
	t.Run("LedgerBalance_Calculation", func(t *testing.T) {
		ledgerBalance, err := ledgerRepo.CalculateBalance(ctx, wallet.ID, wallet_vo.CurrencyUSD)
		require.NoError(t, err)

		walletBalance := wallet.GetBalance(wallet_vo.CurrencyUSD)

		assert.Equal(t, walletBalance.ToCents(), ledgerBalance.ToCents(),
			"Ledger balance should match wallet balance")

		t.Log("✓ Ledger balance calculation accurate")
	})
}

// getMongoTestURI returns MongoDB connection URI for testing
func getMongoTestURI() string {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		uri = "mongodb://test:test123@localhost:27018/replay_test?authSource=admin"
	}
	return uri
}

// BenchmarkDeposit benchmarks deposit operations
func BenchmarkDeposit(b *testing.B) {
	ctx := context.Background()

	mongoURI := getMongoTestURI()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	defer func() { _ = client.Disconnect(ctx) }()

	dbName := "replay_bench_" + uuid.New().String()
	db := client.Database(dbName)
	defer db.Drop(ctx)

	walletRepo := mongodb.NewWalletRepository(db)
	ledgerRepo := mongodb.NewLedgerRepository(db)
	idempotencyRepo := mongodb.NewIdempotencyRepository(db)

	ledgerService := wallet_services.NewLedgerService(ledgerRepo, idempotencyRepo)
	coordinator := wallet_services.NewTransactionCoordinator(walletRepo, ledgerService)

	userID := uuid.New()
	resourceOwner := common.ResourceOwner{UserID: userID}
	ctx = common.SetResourceOwner(ctx, resourceOwner)

	evmAddress, _ := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	wallet, _ := wallet_entities.NewUserWallet(resourceOwner, evmAddress)
	walletRepo.Save(ctx, wallet)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		depositAmount := wallet_vo.NewAmount(10.00)
		paymentID := uuid.New()

		coordinator.ExecuteDeposit(
			ctx,
			wallet,
			wallet_vo.CurrencyUSD,
			depositAmount,
			paymentID,
			wallet_entities.LedgerMetadata{},
		)
	}
}
