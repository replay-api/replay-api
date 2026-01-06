//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_usecases "github.com/replay-api/replay-api/pkg/domain/billing/usecases"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
)

// mockWalletAdapter provides a mock wallet for testing
type mockWalletAdapter struct {
	balance float64
	debits  []mockDebit
	credits []mockCredit
}

type mockDebit struct {
	walletID  uuid.UUID
	amount    float64
	reference string
}

type mockCredit struct {
	walletID  uuid.UUID
	amount    float64
	reference string
}

func newMockWalletAdapter(balance float64) *mockWalletAdapter {
	return &mockWalletAdapter{
		balance: balance,
		debits:  make([]mockDebit, 0),
		credits: make([]mockCredit, 0),
	}
}

func (m *mockWalletAdapter) GetBalance(ctx context.Context, userID uuid.UUID, currency string) (float64, error) {
	return m.balance, nil
}

func (m *mockWalletAdapter) GetByID(ctx context.Context, walletID uuid.UUID) (interface{}, error) {
	return map[string]interface{}{
		"id":       walletID,
		"balance":  m.balance,
		"currency": "USD",
	}, nil
}

func (m *mockWalletAdapter) Debit(ctx context.Context, walletID uuid.UUID, amount float64, reference string) error {
	if m.balance < amount {
		return fmt.Errorf("insufficient balance: have %.2f, need %.2f", m.balance, amount)
	}
	m.balance -= amount
	m.debits = append(m.debits, mockDebit{walletID, amount, reference})
	return nil
}

func (m *mockWalletAdapter) Credit(ctx context.Context, walletID uuid.UUID, amount float64, reference string) error {
	m.balance += amount
	m.credits = append(m.credits, mockCredit{walletID, amount, reference})
	return nil
}

// TestE2E_WithdrawalLifecycle tests the complete withdrawal lifecycle with real MongoDB
func TestE2E_WithdrawalLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup: Connect to test MongoDB
	mongoURI := getMongoTestURI()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err, "Failed to connect to MongoDB")
	defer func() { _ = client.Disconnect(ctx) }()

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("Skipping withdrawal E2E test: MongoDB not available")
	}

	// Create test database
	dbName := "replay_withdrawal_test_" + uuid.New().String()[:8]
	defer func() {
		// Cleanup: Drop test database
		_ = client.Database(dbName).Drop(ctx)
	}()

	// Initialize repository
	withdrawalRepo := db.NewWithdrawalMongoDBRepository(client, dbName)

	// Initialize mock wallet adapter with $1000 balance
	walletAdapter := newMockWalletAdapter(1000.00)

	// Initialize use case
	withdrawalUseCase := billing_usecases.NewWithdrawalUseCase(withdrawalRepo, walletAdapter, walletAdapter)

	// Create test user context
	userID := uuid.New()
	walletID := uuid.New()
	groupID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, groupID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	t.Log("✓ Test environment initialized")

	var createdWithdrawalID uuid.UUID

	// Test 1: Create Withdrawal Successfully
	t.Run("CreateWithdrawal_Success", func(t *testing.T) {
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   100.00,
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodPIX,
			BankDetails: billing_entities.BankDetails{
				PIXKey:     "test@email.com",
				PIXKeyType: "email",
			},
		}

		withdrawal, err := withdrawalUseCase.Create(ctx, cmd)
		require.NoError(t, err, "Withdrawal creation should succeed")
		require.NotNil(t, withdrawal)

		// Verify withdrawal properties
		assert.Equal(t, userID, withdrawal.UserID)
		assert.Equal(t, walletID, withdrawal.WalletID)
		assert.Equal(t, 100.00, withdrawal.Amount)
		assert.Equal(t, "USD", withdrawal.Currency)
		assert.Equal(t, billing_entities.WithdrawalMethodPIX, withdrawal.Method)
		assert.Equal(t, billing_entities.WithdrawalStatusPending, withdrawal.Status)
		assert.Greater(t, withdrawal.Fee, 0.0, "Fee should be calculated")
		assert.Equal(t, withdrawal.Amount-withdrawal.Fee, withdrawal.NetAmount)

		createdWithdrawalID = withdrawal.ID

		// Verify wallet was debited
		assert.Equal(t, 1, len(walletAdapter.debits), "Wallet should have 1 debit")
		assert.Equal(t, 100.00, walletAdapter.debits[0].amount)

		t.Logf("✓ Withdrawal created: ID=%s, Amount=$%.2f, Fee=$%.2f, Net=$%.2f",
			withdrawal.ID, withdrawal.Amount, withdrawal.Fee, withdrawal.NetAmount)
	})

	// Test 2: Get Withdrawal by ID
	t.Run("GetWithdrawal_ByID", func(t *testing.T) {
		withdrawal, err := withdrawalUseCase.GetByID(ctx, createdWithdrawalID)
		require.NoError(t, err)
		require.NotNil(t, withdrawal)

		assert.Equal(t, createdWithdrawalID, withdrawal.ID)
		assert.Equal(t, userID, withdrawal.UserID)

		t.Log("✓ Withdrawal retrieved successfully")
	})

	// Test 3: List User Withdrawals
	t.Run("ListWithdrawals_ByUser", func(t *testing.T) {
		withdrawals, err := withdrawalUseCase.GetByUserID(ctx, userID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(withdrawals), 1, "Should have at least 1 withdrawal")

		t.Logf("✓ Listed %d withdrawals for user", len(withdrawals))
	})

	// Test 4: Cancel Withdrawal
	t.Run("CancelWithdrawal_Success", func(t *testing.T) {
		// Create another withdrawal to cancel
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   50.00,
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodPayPal,
			BankDetails: billing_entities.BankDetails{
				PayPalEmail: "test@paypal.com",
			},
		}

		withdrawal, err := withdrawalUseCase.Create(ctx, cmd)
		require.NoError(t, err)

		initialCredits := len(walletAdapter.credits)

		// Cancel the withdrawal
		canceled, err := withdrawalUseCase.Cancel(ctx, withdrawal.ID)
		require.NoError(t, err)
		require.NotNil(t, canceled)

		assert.Equal(t, billing_entities.WithdrawalStatusCanceled, canceled.Status)

		// Verify wallet was credited (refund)
		assert.Equal(t, initialCredits+1, len(walletAdapter.credits), "Wallet should have 1 new credit")
		assert.Equal(t, 50.00, walletAdapter.credits[len(walletAdapter.credits)-1].amount)

		t.Log("✓ Withdrawal canceled and funds refunded")
	})

	// Test 5: Insufficient Balance
	t.Run("CreateWithdrawal_InsufficientBalance", func(t *testing.T) {
		// Try to withdraw more than available balance
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   99999.00, // More than wallet balance
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodBankTransfer,
			BankDetails: billing_entities.BankDetails{
				AccountNumber: "123456789",
				AccountHolder: "Test User",
			},
		}

		_, err := withdrawalUseCase.Create(ctx, cmd)
		require.Error(t, err, "Should fail due to insufficient balance")
		assert.Contains(t, err.Error(), "insufficient balance")

		t.Log("✓ Insufficient balance correctly rejected")
	})

	// Test 6: Minimum Amount Validation
	t.Run("CreateWithdrawal_BelowMinimum", func(t *testing.T) {
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   1.00, // Below minimum ($10)
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodPIX,
			BankDetails: billing_entities.BankDetails{
				PIXKey: "test@email.com",
			},
		}

		_, err := withdrawalUseCase.Create(ctx, cmd)
		require.Error(t, err, "Should fail due to below minimum")
		assert.Contains(t, err.Error(), "minimum withdrawal amount")

		t.Log("✓ Below minimum amount correctly rejected")
	})

	// Test 7: Invalid Bank Details
	t.Run("CreateWithdrawal_MissingBankDetails", func(t *testing.T) {
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   50.00,
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodPIX,
			BankDetails: billing_entities.BankDetails{
				// Missing PIXKey
			},
		}

		_, err := withdrawalUseCase.Create(ctx, cmd)
		require.Error(t, err, "Should fail due to missing PIX key")
		assert.Contains(t, err.Error(), "PIX key")

		t.Log("✓ Missing bank details correctly rejected")
	})

	// Test 8: Unauthorized User Cannot Cancel Others' Withdrawal
	t.Run("CancelWithdrawal_Unauthorized", func(t *testing.T) {
		// Create withdrawal with original user
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   25.00,
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodPIX,
			BankDetails: billing_entities.BankDetails{
				PIXKey: "test@email.com",
			},
		}

		withdrawal, err := withdrawalUseCase.Create(ctx, cmd)
		require.NoError(t, err)

		// Create context with different user
		otherUserID := uuid.New()
		otherGroupID := uuid.New()
		otherCtx := context.WithValue(ctx, shared.UserIDKey, otherUserID)
		otherCtx = context.WithValue(otherCtx, shared.GroupIDKey, otherGroupID)
		otherCtx = context.WithValue(otherCtx, shared.TenantIDKey, replay_common.TeamPROTenantID)
		otherCtx = context.WithValue(otherCtx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
		otherCtx = context.WithValue(otherCtx, shared.AuthenticatedKey, true)

		// Try to cancel with different user
		_, err = withdrawalUseCase.Cancel(otherCtx, withdrawal.ID)
		require.Error(t, err, "Should fail - cannot cancel another user's withdrawal")

		t.Log("✓ Unauthorized cancellation correctly rejected")
	})

	// Test 9: Multiple Withdrawals Fee Calculation
	t.Run("WithdrawalFee_Calculation", func(t *testing.T) {
		testCases := []struct {
			amount      float64
			expectedFee float64 // Based on DefaultWithdrawalFeeConfig
		}{
			{10.00, 0.70},   // $0.50 + 2% = $0.70
			{100.00, 2.50},  // $0.50 + 2% = $2.50
			{500.00, 10.50}, // $0.50 + 2% = $10.50
		}

		for _, tc := range testCases {
			cmd := billing_in.CreateWithdrawalCommand{
				UserID:   userID,
				WalletID: walletID,
				Amount:   tc.amount,
				Currency: "USD",
				Method:   billing_entities.WithdrawalMethodPIX,
				BankDetails: billing_entities.BankDetails{
					PIXKey: "test@email.com",
				},
			}

			withdrawal, err := withdrawalUseCase.Create(ctx, cmd)
			require.NoError(t, err)

			assert.InDelta(t, tc.expectedFee, withdrawal.Fee, 0.01,
				"Fee for $%.2f should be ~$%.2f, got $%.2f",
				tc.amount, tc.expectedFee, withdrawal.Fee)
		}

		t.Log("✓ Fee calculation verified")
	})

	// Test 10: Withdrawal History Tracking
	t.Run("Withdrawal_HistoryTracking", func(t *testing.T) {
		cmd := billing_in.CreateWithdrawalCommand{
			UserID:   userID,
			WalletID: walletID,
			Amount:   15.00,
			Currency: "USD",
			Method:   billing_entities.WithdrawalMethodPIX,
			BankDetails: billing_entities.BankDetails{
				PIXKey: "test@email.com",
			},
		}

		withdrawal, err := withdrawalUseCase.Create(ctx, cmd)
		require.NoError(t, err)

		// Check history has initial entry
		assert.Len(t, withdrawal.History, 1, "Should have 1 history entry on creation")
		assert.Equal(t, billing_entities.WithdrawalStatusPending, withdrawal.History[0].Status)
		assert.False(t, withdrawal.History[0].Timestamp.IsZero())

		t.Log("✓ Withdrawal history tracking verified")
	})

	t.Log("✓ All withdrawal E2E tests passed!")
}

// TestE2E_WithdrawalPersistence verifies data persists correctly in MongoDB
func TestE2E_WithdrawalPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	mongoURI := getMongoTestURI()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)
	defer func() { _ = client.Disconnect(ctx) }()

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("Skipping: MongoDB not available")
	}

	dbName := "replay_withdrawal_persist_" + uuid.New().String()[:8]
	defer func() { _ = client.Database(dbName).Drop(ctx) }()

	withdrawalRepo := db.NewWithdrawalMongoDBRepository(client, dbName)

	userID := uuid.New()
	walletID := uuid.New()
	resourceOwner := shared.ResourceOwner{
		UserID:   userID,
		GroupID:  uuid.New(),
		TenantID: replay_common.TeamPROTenantID,
		ClientID: replay_common.TeamPROAppClientID,
	}

	// Create withdrawal directly via repository
	withdrawal := billing_entities.NewWithdrawal(
		userID,
		walletID,
		250.00,
		"USD",
		billing_entities.WithdrawalMethodCrypto,
		billing_entities.BankDetails{
			CryptoAddress: "0x1234567890abcdef",
			CryptoNetwork: "ethereum",
		},
		5.00,
		resourceOwner,
	)

	// Save
	created, err := withdrawalRepo.Create(ctx, withdrawal)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Retrieve
	retrieved, err := withdrawalRepo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Verify all fields persisted
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.UserID, retrieved.UserID)
	assert.Equal(t, created.WalletID, retrieved.WalletID)
	assert.Equal(t, created.Amount, retrieved.Amount)
	assert.Equal(t, created.Currency, retrieved.Currency)
	assert.Equal(t, created.Method, retrieved.Method)
	assert.Equal(t, created.BankDetails.CryptoAddress, retrieved.BankDetails.CryptoAddress)
	assert.Equal(t, created.BankDetails.CryptoNetwork, retrieved.BankDetails.CryptoNetwork)
	assert.Equal(t, created.Fee, retrieved.Fee)
	assert.Equal(t, created.NetAmount, retrieved.NetAmount)
	assert.Equal(t, created.Status, retrieved.Status)

	// Update
	retrieved.Status = billing_entities.WithdrawalStatusProcessing
	now := time.Now()
	retrieved.ProcessedAt = &now
	// Add history entry directly
	retrieved.History = append(retrieved.History, billing_entities.WithdrawalHistory{
		Status:    billing_entities.WithdrawalStatusProcessing,
		Reason:    "Processing started",
		Timestamp: now,
	})

	updated, err := withdrawalRepo.Update(ctx, retrieved)
	require.NoError(t, err)

	// Verify update
	finalRetrieve, err := withdrawalRepo.GetByID(ctx, updated.ID)
	require.NoError(t, err)
	assert.Equal(t, billing_entities.WithdrawalStatusProcessing, finalRetrieve.Status)
	assert.NotNil(t, finalRetrieve.ProcessedAt)
	assert.Len(t, finalRetrieve.History, 2) // Initial + Processing

	t.Log("✓ Withdrawal persistence verified")
}

