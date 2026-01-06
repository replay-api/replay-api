package payment_usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	payment_usecases "github.com/replay-api/replay-api/pkg/domain/payment/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserPayments_Success(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	walletID := uuid.New()

	testPayments := []*payment_entities.Payment{
		createTestPayment(userID, walletID),
		createTestPayment(userID, walletID),
		createTestPayment(userID, walletID),
	}

	mockRepo.On("FindByUserID", mock.Anything, userID, mock.AnythingOfType("payment_out.PaymentFilters")).Return(testPayments, nil)

	query := payment_in.GetUserPaymentsQuery{
		UserID: userID,
		Filters: payment_in.PaymentQueryFilters{
			Limit:  20,
			Offset: 0,
		},
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Payments, 3)
	assert.Equal(t, 20, result.Limit)
	assert.Equal(t, 0, result.Offset)
	mockRepo.AssertExpectations(t)
}

func TestGetUserPayments_Unauthenticated(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	// No authentication context

	query := payment_in.GetUserPaymentsQuery{
		UserID: uuid.New(),
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGetUserPayments_InvalidQuery_NoUserID(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	query := payment_in.GetUserPaymentsQuery{
		UserID: uuid.Nil,
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetUserPayments_EmptyResult(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()

	mockRepo.On("FindByUserID", mock.Anything, userID, mock.AnythingOfType("payment_out.PaymentFilters")).Return([]*payment_entities.Payment{}, nil)

	query := payment_in.GetUserPaymentsQuery{
		UserID: userID,
		Filters: payment_in.PaymentQueryFilters{
			Limit:  20,
			Offset: 0,
		},
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Payments, 0)
	mockRepo.AssertExpectations(t)
}

func TestGetUserPayments_WithFilters(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	walletID := uuid.New()

	statusFilter := payment_entities.PaymentStatusSucceeded
	typeFilter := payment_entities.PaymentTypeDeposit

	testPayments := []*payment_entities.Payment{
		createTestPayment(userID, walletID),
	}

	mockRepo.On("FindByUserID", mock.Anything, userID, mock.MatchedBy(func(f payment_out.PaymentFilters) bool {
		return f.Status != nil && *f.Status == statusFilter &&
			f.Type != nil && *f.Type == typeFilter
	})).Return(testPayments, nil)

	query := payment_in.GetUserPaymentsQuery{
		UserID: userID,
		Filters: payment_in.PaymentQueryFilters{
			Status: &statusFilter,
			Type:   &typeFilter,
			Limit:  20,
			Offset: 0,
		},
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Payments, 1)
	mockRepo.AssertExpectations(t)
}

func TestGetUserPayments_RepositoryError(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()

	mockRepo.On("FindByUserID", mock.Anything, userID, mock.AnythingOfType("payment_out.PaymentFilters")).Return(nil, errors.New("database error"))

	query := payment_in.GetUserPaymentsQuery{
		UserID: userID,
		Filters: payment_in.PaymentQueryFilters{
			Limit:  20,
			Offset: 0,
		},
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetUserPayments_DefaultLimit(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()

	mockRepo.On("FindByUserID", mock.Anything, userID, mock.MatchedBy(func(f payment_out.PaymentFilters) bool {
		return f.Limit == 20 // Default limit applied
	})).Return([]*payment_entities.Payment{}, nil)

	// Query with no limit set (should default to 20)
	query := payment_in.GetUserPaymentsQuery{
		UserID:  userID,
		Filters: payment_in.PaymentQueryFilters{},
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 20, result.Limit)
	mockRepo.AssertExpectations(t)
}

func TestGetUserPayments_MaxLimit(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	usecase := payment_usecases.NewGetUserPaymentsUseCase(mockRepo)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()

	mockRepo.On("FindByUserID", mock.Anything, userID, mock.MatchedBy(func(f payment_out.PaymentFilters) bool {
		return f.Limit == 100 // Max limit enforced
	})).Return([]*payment_entities.Payment{}, nil)

	// Query with limit exceeding max (should be capped at 100)
	query := payment_in.GetUserPaymentsQuery{
		UserID: userID,
		Filters: payment_in.PaymentQueryFilters{
			Limit: 500,
		},
	}

	result, err := usecase.GetUserPayments(ctx, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, result.Limit)
	mockRepo.AssertExpectations(t)
}
