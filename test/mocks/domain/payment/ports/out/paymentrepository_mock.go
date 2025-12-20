package payment_out

import (
	"context"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockPaymentRepository is a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

// Save provides a mock function
func (_m *MockPaymentRepository) Save(ctx context.Context, payment *payment_entities.Payment) error {
	ret := _m.Called(ctx, payment)

	return ret.Error(0)
}

// FindByID provides a mock function
func (_m *MockPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*payment_entities.Payment, error) {
	ret := _m.Called(ctx, id)

	var r0 *payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*payment_entities.Payment, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *payment_entities.Payment); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByProviderPaymentID provides a mock function
func (_m *MockPaymentRepository) FindByProviderPaymentID(ctx context.Context, providerPaymentID string) (*payment_entities.Payment, error) {
	ret := _m.Called(ctx, providerPaymentID)

	var r0 *payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*payment_entities.Payment, error)); ok {
		return rf(ctx, providerPaymentID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *payment_entities.Payment); ok {
		r0 = rf(ctx, providerPaymentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByIdempotencyKey provides a mock function
func (_m *MockPaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*payment_entities.Payment, error) {
	ret := _m.Called(ctx, key)

	var r0 *payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*payment_entities.Payment, error)); ok {
		return rf(ctx, key)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) *payment_entities.Payment); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByUserID provides a mock function
func (_m *MockPaymentRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	ret := _m.Called(ctx, userID, filters)

	var r0 []*payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, payment_out.PaymentFilters) ([]*payment_entities.Payment, error)); ok {
		return rf(ctx, userID, filters)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, payment_out.PaymentFilters) []*payment_entities.Payment); ok {
		r0 = rf(ctx, userID, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByWalletID provides a mock function
func (_m *MockPaymentRepository) FindByWalletID(ctx context.Context, walletID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	ret := _m.Called(ctx, walletID, filters)

	var r0 []*payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, payment_out.PaymentFilters) ([]*payment_entities.Payment, error)); ok {
		return rf(ctx, walletID, filters)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, payment_out.PaymentFilters) []*payment_entities.Payment); ok {
		r0 = rf(ctx, walletID, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockPaymentRepository) Update(ctx context.Context, payment *payment_entities.Payment) error {
	ret := _m.Called(ctx, payment)

	return ret.Error(0)
}

// GetPendingPayments provides a mock function
func (_m *MockPaymentRepository) GetPendingPayments(ctx context.Context, olderThanSeconds int) ([]*payment_entities.Payment, error) {
	ret := _m.Called(ctx, olderThanSeconds)

	var r0 []*payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int) ([]*payment_entities.Payment, error)); ok {
		return rf(ctx, olderThanSeconds)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int) []*payment_entities.Payment); ok {
		r0 = rf(ctx, olderThanSeconds)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPaymentRepository creates a new instance of MockPaymentRepository
func NewMockPaymentRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPaymentRepository {
	mock := &MockPaymentRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
