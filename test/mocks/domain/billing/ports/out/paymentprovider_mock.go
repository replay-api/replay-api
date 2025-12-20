package billing_out

import (
	"context"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockPaymentProvider is a mock implementation of PaymentProvider
type MockPaymentProvider struct {
	mock.Mock
}

// ProcessWithdrawal provides a mock function
func (_m *MockPaymentProvider) ProcessWithdrawal(ctx context.Context, withdrawal *billing_entities.Withdrawal) (string, error) {
	ret := _m.Called(ctx, withdrawal)

	var r0 string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Withdrawal) (string, error)); ok {
		return rf(ctx, withdrawal)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.Withdrawal) string); ok {
		r0 = rf(ctx, withdrawal)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetWithdrawalStatus provides a mock function
func (_m *MockPaymentProvider) GetWithdrawalStatus(ctx context.Context, providerRef string) (string, error) {
	ret := _m.Called(ctx, providerRef)

	var r0 string
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, providerRef)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, providerRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(string)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPaymentProvider creates a new instance of MockPaymentProvider
func NewMockPaymentProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPaymentProvider {
	mock := &MockPaymentProvider{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
