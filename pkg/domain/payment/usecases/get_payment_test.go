package payment_usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	payment_usecases "github.com/replay-api/replay-api/pkg/domain/payment/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPaymentRepository is a mock implementation of payment_out.PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Save(ctx context.Context, payment *payment_entities.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*payment_entities.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment_entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByProviderPaymentID(ctx context.Context, providerPaymentID string) (*payment_entities.Payment, error) {
	args := m.Called(ctx, providerPaymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment_entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*payment_entities.Payment, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment_entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	args := m.Called(ctx, userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*payment_entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByWalletID(ctx context.Context, walletID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	args := m.Called(ctx, walletID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*payment_entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *payment_entities.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPendingPayments(ctx context.Context, olderThanSeconds int) ([]*payment_entities.Payment, error) {
	args := m.Called(ctx, olderThanSeconds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*payment_entities.Payment), args.Error(1)
}

func createTestPayment(userID, walletID uuid.UUID) *payment_entities.Payment {
	now := time.Now().UTC()
	return &payment_entities.Payment{
		ID:                uuid.New(),
		UserID:            userID,
		WalletID:          walletID,
		Type:              payment_entities.PaymentTypeDeposit,
		Provider:          payment_entities.PaymentProviderStripe,
		Status:            payment_entities.PaymentStatusSucceeded,
		Amount:            10000, // $100.00 in cents
		Currency:          "USD",
		Fee:               100, // $1.00
		ProviderFee:       50,  // $0.50
		NetAmount:         9850,
		ProviderPaymentID: "pi_test123",
		Description:       "Test deposit",
		CreatedAt:         now,
		UpdatedAt:         now,
		IdempotencyKey:    uuid.New().String(),
	}
}

func TestGetPayment_Success(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetPaymentUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	walletID := uuid.New()

	testPayment := createTestPayment(userID, walletID)

	mockRepo.On("FindByID", mock.Anything, testPayment.ID).Return(testPayment, nil)

	query := payment_in.GetPaymentQuery{
		PaymentID: testPayment.ID,
		UserID:    userID,
	}

	result, err := usecase.GetPayment(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testPayment.ID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, walletID, result.WalletID)
	assert.Equal(t, payment_entities.PaymentTypeDeposit, result.Type)
	assert.Equal(t, int64(10000), result.Amount)
	mockRepo.AssertExpectations(t)
}

func TestGetPayment_Unauthenticated(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetPaymentUseCase(mockRepo)

	ctx := context.Background()
	// No authentication context

	query := payment_in.GetPaymentQuery{
		PaymentID: uuid.New(),
		UserID:    uuid.New(),
	}

	result, err := usecase.GetPayment(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGetPayment_InvalidQuery_NoPaymentID(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetPaymentUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)

	query := payment_in.GetPaymentQuery{
		PaymentID: uuid.Nil,
		UserID:    uuid.New(),
	}

	result, err := usecase.GetPayment(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetPayment_InvalidQuery_NoUserID(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetPaymentUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)

	query := payment_in.GetPaymentQuery{
		PaymentID: uuid.New(),
		UserID:    uuid.Nil,
	}

	result, err := usecase.GetPayment(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetPayment_NotFound(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetPaymentUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)

	paymentID := uuid.New()
	userID := uuid.New()

	mockRepo.On("FindByID", mock.Anything, paymentID).Return(nil, errors.New("not found"))

	query := payment_in.GetPaymentQuery{
		PaymentID: paymentID,
		UserID:    userID,
	}

	result, err := usecase.GetPayment(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestGetPayment_UnauthorizedAccess(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetPaymentUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)

	ownerID := uuid.New()
	attackerID := uuid.New()
	walletID := uuid.New()

	testPayment := createTestPayment(ownerID, walletID)

	mockRepo.On("FindByID", mock.Anything, testPayment.ID).Return(testPayment, nil)

	// Attacker tries to access owner's payment
	query := payment_in.GetPaymentQuery{
		PaymentID: testPayment.ID,
		UserID:    attackerID,
	}

	result, err := usecase.GetPayment(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unauthorized")
	mockRepo.AssertExpectations(t)
}
