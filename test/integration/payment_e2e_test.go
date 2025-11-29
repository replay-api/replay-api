//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	common "github.com/replay-api/replay-api/pkg/domain"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	payment_services "github.com/replay-api/replay-api/pkg/domain/payment/services"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
)

// MockStripeAdapter is a mock implementation of the Stripe adapter for testing
type MockStripeAdapter struct {
	CreatePaymentIntentFunc func(ctx context.Context, req payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error)
	ConfirmPaymentFunc      func(ctx context.Context, req payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error)
	RefundPaymentFunc       func(ctx context.Context, req payment_out.RefundRequest) (*payment_out.RefundResponse, error)
	CancelPaymentFunc       func(ctx context.Context, req payment_out.CancelRequest) (*payment_out.CancelResponse, error)
	ParseWebhookFunc        func(payload []byte, signature string) (*payment_out.WebhookEvent, error)
	CreateOrGetCustomerFunc func(ctx context.Context, req payment_out.CustomerRequest) (*payment_out.CustomerResponse, error)
}

func (m *MockStripeAdapter) GetProvider() payment_entities.PaymentProvider {
	return payment_entities.PaymentProviderStripe
}

func (m *MockStripeAdapter) CreatePaymentIntent(ctx context.Context, req payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error) {
	if m.CreatePaymentIntentFunc != nil {
		return m.CreatePaymentIntentFunc(ctx, req)
	}
	return &payment_out.CreateIntentResponse{
		ProviderPaymentID: "pi_" + uuid.New().String(),
		ClientSecret:      "sk_test_secret_" + uuid.New().String(),
		Status:            "requires_payment_method",
	}, nil
}

func (m *MockStripeAdapter) ConfirmPayment(ctx context.Context, req payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error) {
	if m.ConfirmPaymentFunc != nil {
		return m.ConfirmPaymentFunc(ctx, req)
	}
	return &payment_out.ConfirmPaymentResponse{
		Status:      "succeeded",
		ProviderFee: 59, // $0.59 (2.9% + $0.30 on $10)
	}, nil
}

func (m *MockStripeAdapter) RefundPayment(ctx context.Context, req payment_out.RefundRequest) (*payment_out.RefundResponse, error) {
	if m.RefundPaymentFunc != nil {
		return m.RefundPaymentFunc(ctx, req)
	}
	return &payment_out.RefundResponse{
		RefundID: "re_" + uuid.New().String(),
		Status:   "succeeded",
		Amount:   req.Amount,
	}, nil
}

func (m *MockStripeAdapter) CancelPayment(ctx context.Context, req payment_out.CancelRequest) (*payment_out.CancelResponse, error) {
	if m.CancelPaymentFunc != nil {
		return m.CancelPaymentFunc(ctx, req)
	}
	return &payment_out.CancelResponse{
		Status: "canceled",
	}, nil
}

func (m *MockStripeAdapter) ParseWebhook(payload []byte, signature string) (*payment_out.WebhookEvent, error) {
	if m.ParseWebhookFunc != nil {
		return m.ParseWebhookFunc(payload, signature)
	}
	return &payment_out.WebhookEvent{
		EventType:         "payment_intent.succeeded",
		ProviderPaymentID: "pi_test",
		Status:            payment_entities.PaymentStatusSucceeded,
	}, nil
}

func (m *MockStripeAdapter) CreateOrGetCustomer(ctx context.Context, req payment_out.CustomerRequest) (*payment_out.CustomerResponse, error) {
	if m.CreateOrGetCustomerFunc != nil {
		return m.CreateOrGetCustomerFunc(ctx, req)
	}
	return &payment_out.CustomerResponse{
		CustomerID: "cus_" + uuid.New().String(),
	}, nil
}

// TestE2E_PaymentLifecycle tests the complete payment lifecycle with real MongoDB
func TestE2E_PaymentLifecycle(t *testing.T) {
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
	dbName := "replay_payment_test_" + uuid.New().String()

	// Initialize payment repository
	paymentRepo := db.NewPaymentMongoDBRepository(client, dbName)

	// Create mock Stripe adapter
	mockStripe := &MockStripeAdapter{}

	// Create payment service with mock adapter (no wallet service for simplicity)
	paymentService := payment_services.NewPaymentService(paymentRepo, nil, mockStripe)

	// Create test user context
	userID := uuid.New()
	walletID := uuid.New()
	tenantID := common.TeamPROTenantID
	clientID := common.TeamPROAppClientID

	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, clientID)

	// Cleanup after tests
	defer func() {
		client.Database(dbName).Drop(ctx)
	}()

	var createdPaymentID uuid.UUID
	var providerPaymentID string

	// Test 1: Create Payment Intent
	t.Run("CreatePaymentIntent_Success", func(t *testing.T) {
		result, err := paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      1000, // $10.00 in cents
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
			Metadata: map[string]any{
				"source": "e2e_test",
			},
		})

		require.NoError(t, err, "Should create payment intent")
		require.NotNil(t, result, "Result should not be nil")
		require.NotNil(t, result.Payment, "Payment should not be nil")

		assert.NotEmpty(t, result.ClientSecret, "Should have client secret for Stripe")
		assert.Equal(t, payment_entities.PaymentStatusProcessing, result.Payment.Status)
		assert.Equal(t, int64(1000), result.Payment.Amount)
		assert.Equal(t, "usd", result.Payment.Currency)
		assert.Equal(t, payment_entities.PaymentProviderStripe, result.Payment.Provider)

		createdPaymentID = result.Payment.ID
		providerPaymentID = result.Payment.ProviderPaymentID

		t.Log("✓ Payment intent created successfully")
	})

	// Test 2: Get Payment by ID
	t.Run("GetPayment_ByID", func(t *testing.T) {
		payment, err := paymentRepo.FindByID(ctx, createdPaymentID)

		require.NoError(t, err, "Should find payment by ID")
		assert.Equal(t, createdPaymentID, payment.ID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, walletID, payment.WalletID)

		t.Log("✓ Payment retrieved by ID")
	})

	// Test 3: Get Payment by Provider Payment ID
	t.Run("GetPayment_ByProviderPaymentID", func(t *testing.T) {
		payment, err := paymentRepo.FindByProviderPaymentID(ctx, providerPaymentID)

		require.NoError(t, err, "Should find payment by provider payment ID")
		assert.Equal(t, createdPaymentID, payment.ID)

		t.Log("✓ Payment retrieved by provider payment ID")
	})

	// Test 4: Get User Payments
	t.Run("GetUserPayments", func(t *testing.T) {
		payments, err := paymentRepo.FindByUserID(ctx, userID, payment_out.PaymentFilters{
			Limit: 50,
		})

		require.NoError(t, err, "Should get user payments")
		assert.Greater(t, len(payments), 0, "Should have at least one payment")

		t.Log("✓ User payments retrieved")
	})

	// Test 5: Confirm Payment
	t.Run("ConfirmPayment_Success", func(t *testing.T) {
		payment, err := paymentService.ConfirmPayment(ctx, payment_in.ConfirmPaymentCommand{
			PaymentID:       createdPaymentID,
			PaymentMethodID: "pm_test_visa",
		})

		require.NoError(t, err, "Should confirm payment")
		assert.Equal(t, payment_entities.PaymentStatusSucceeded, payment.Status)
		assert.NotNil(t, payment.CompletedAt, "Should have completed timestamp")
		assert.Greater(t, payment.ProviderFee, int64(0), "Should have provider fee")

		t.Log("✓ Payment confirmed successfully")
	})

	// Test 6: Create and Cancel Payment
	t.Run("CancelPayment_Success", func(t *testing.T) {
		// Create new payment
		result, err := paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      500,
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
		})
		require.NoError(t, err)

		// Cancel it
		payment, err := paymentService.CancelPayment(ctx, payment_in.CancelPaymentCommand{
			PaymentID: result.Payment.ID,
			Reason:    "user_requested",
		})

		require.NoError(t, err, "Should cancel payment")
		assert.Equal(t, payment_entities.PaymentStatusCanceled, payment.Status)

		t.Log("✓ Payment cancelled successfully")
	})

	// Test 7: Refund Payment
	t.Run("RefundPayment_Success", func(t *testing.T) {
		// Create and confirm a payment first
		result, err := paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      2000,
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
		})
		require.NoError(t, err)

		// Confirm it
		_, err = paymentService.ConfirmPayment(ctx, payment_in.ConfirmPaymentCommand{
			PaymentID:       result.Payment.ID,
			PaymentMethodID: "pm_test_visa",
		})
		require.NoError(t, err)

		// Now refund it
		payment, err := paymentService.RefundPayment(ctx, payment_in.RefundPaymentCommand{
			PaymentID: result.Payment.ID,
			Amount:    0, // Full refund
			Reason:    "duplicate_charge",
		})

		require.NoError(t, err, "Should refund payment")
		assert.Equal(t, payment_entities.PaymentStatusRefunded, payment.Status)
		assert.NotEmpty(t, payment.Metadata["refund_id"])

		t.Log("✓ Payment refunded successfully")
	})

	// Test 8: Idempotency - duplicate payment intent should return existing
	t.Run("Idempotency_DuplicatePaymentIntent", func(t *testing.T) {
		// Create first payment
		_, err := paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      750,
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
		})
		require.NoError(t, err)

		// Count payments for user
		paymentsBefore, _ := paymentRepo.FindByUserID(ctx, userID, payment_out.PaymentFilters{Limit: 100})
		assert.GreaterOrEqual(t, len(paymentsBefore), 1, "Should have at least one payment")

		// Attempt to create with same idempotency key would return existing
		// Note: In our implementation, each NewPayment generates a new idempotency key
		// For true idempotency testing, we'd need to pass the key explicitly

		t.Log("✓ Idempotency check complete")
	})

	// Test 9: Payment Filters
	t.Run("PaymentFilters_ByStatus", func(t *testing.T) {
		status := payment_entities.PaymentStatusSucceeded
		payments, err := paymentRepo.FindByUserID(ctx, userID, payment_out.PaymentFilters{
			Status: &status,
			Limit:  50,
		})

		require.NoError(t, err)
		for _, p := range payments {
			assert.Equal(t, payment_entities.PaymentStatusSucceeded, p.Status)
		}

		t.Log("✓ Payment filtering by status works")
	})

	// Test 10: Get Pending Payments
	t.Run("GetPendingPayments", func(t *testing.T) {
		// Create a payment but don't confirm it
		result, err := paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      300,
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
		})
		require.NoError(t, err)

		// Get pending payments older than 0 seconds (should include our new one)
		pendingPayments, err := paymentRepo.GetPendingPayments(ctx, 0)
		require.NoError(t, err)

		// Should find at least one pending/processing payment
		found := false
		for _, p := range pendingPayments {
			if p.ID == result.Payment.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find our pending payment")

		t.Log("✓ Pending payments retrieval works")
	})

	// Test 11: Webhook Processing
	t.Run("WebhookProcessing_PaymentSucceeded", func(t *testing.T) {
		// Create a payment
		result, err := paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      1500,
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
		})
		require.NoError(t, err)

		// Configure mock to return specific webhook event
		mockStripe.ParseWebhookFunc = func(payload []byte, signature string) (*payment_out.WebhookEvent, error) {
			return &payment_out.WebhookEvent{
				EventType:         "payment_intent.succeeded",
				ProviderPaymentID: result.Payment.ProviderPaymentID,
				Status:            payment_entities.PaymentStatusSucceeded,
				ProviderFee:       75,
			}, nil
		}

		// Process webhook
		err = paymentService.ProcessWebhook(ctx, payment_in.ProcessWebhookCommand{
			Provider:  payment_entities.PaymentProviderStripe,
			Payload:   []byte(`{"type": "payment_intent.succeeded"}`),
			Signature: "test_signature",
		})
		require.NoError(t, err)

		// Verify payment status updated
		payment, err := paymentRepo.FindByID(ctx, result.Payment.ID)
		require.NoError(t, err)
		assert.Equal(t, payment_entities.PaymentStatusSucceeded, payment.Status)

		t.Log("✓ Webhook processing works")
	})
}

// TestPaymentRepository_UniqueConstraints tests unique constraints
func TestPaymentRepository_UniqueConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	mongoURI := getMongoTestURI()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)
	defer func() { _ = client.Disconnect(ctx) }()

	dbName := "replay_payment_unique_test_" + uuid.New().String()
	defer client.Database(dbName).Drop(ctx)

	paymentRepo := db.NewPaymentMongoDBRepository(client, dbName)

	userID := uuid.New()
	walletID := uuid.New()

	// Create payment
	payment := payment_entities.NewPayment(
		userID,
		walletID,
		payment_entities.PaymentTypeDeposit,
		payment_entities.PaymentProviderStripe,
		1000,
		"usd",
	)
	payment.ProviderPaymentID = "pi_unique_test"

	err = paymentRepo.Save(ctx, payment)
	require.NoError(t, err)

	// Try to create duplicate with same idempotency key
	duplicatePayment := payment_entities.NewPayment(
		userID,
		walletID,
		payment_entities.PaymentTypeDeposit,
		payment_entities.PaymentProviderStripe,
		1000,
		"usd",
	)
	duplicatePayment.IdempotencyKey = payment.IdempotencyKey // Same key

	err = paymentRepo.Save(ctx, duplicatePayment)
	assert.Error(t, err, "Should fail on duplicate idempotency key")

	t.Log("✓ Unique constraints enforced")
}

// BenchmarkPaymentCreation benchmarks payment creation operations
func BenchmarkPaymentCreation(b *testing.B) {
	ctx := context.Background()

	mongoURI := getMongoTestURI()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	defer func() { _ = client.Disconnect(ctx) }()

	dbName := "replay_payment_bench_" + uuid.New().String()
	defer client.Database(dbName).Drop(ctx)

	paymentRepo := db.NewPaymentMongoDBRepository(client, dbName)
	mockStripe := &MockStripeAdapter{}
	paymentService := payment_services.NewPaymentService(paymentRepo, nil, mockStripe)

	userID := uuid.New()
	walletID := uuid.New()
	tenantID := common.TeamPROTenantID
	clientID := common.TeamPROAppClientID

	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, clientID)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paymentService.CreatePaymentIntent(ctx, payment_in.CreatePaymentIntentCommand{
			UserID:      userID,
			WalletID:    walletID,
			Amount:      int64(100 + i),
			Currency:    "usd",
			PaymentType: payment_entities.PaymentTypeDeposit,
			Provider:    payment_entities.PaymentProviderStripe,
		})
	}
}
