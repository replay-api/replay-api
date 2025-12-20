package tournament_services

import (
	"context"

	big "math/big"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockWalletService is a mock implementation of WalletService
type MockWalletService struct {
	mock.Mock
}

// GetBalance provides a mock function
func (_m *MockWalletService) GetBalance(ctx context.Context, userID uuid.UUID, currency string) (*big.Float, error) {
	ret := _m.Called(ctx, userID, currency)

	var r0 *big.Float
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*big.Float, error)); ok {
		return rf(ctx, userID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *big.Float); ok {
		r0 = rf(ctx, userID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Float)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Hold provides a mock function
func (_m *MockWalletService) Hold(ctx context.Context, userID uuid.UUID, amount *big.Float, currency string, reference uuid.UUID, description string) error {
	ret := _m.Called(ctx, userID, amount, currency, reference, description)

	return ret.Error(0)
}

// Release provides a mock function
func (_m *MockWalletService) Release(ctx context.Context, userID uuid.UUID, amount *big.Float, currency string, reference uuid.UUID) error {
	ret := _m.Called(ctx, userID, amount, currency, reference)

	return ret.Error(0)
}

// Transfer provides a mock function
func (_m *MockWalletService) Transfer(ctx context.Context, fromUserID uuid.UUID, toUserID uuid.UUID, amount *big.Float, currency string, reference uuid.UUID, description string) error {
	ret := _m.Called(ctx, fromUserID, toUserID, amount, currency, reference, description)

	return ret.Error(0)
}

// NewMockWalletService creates a new instance of MockWalletService
func NewMockWalletService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWalletService {
	mock := &MockWalletService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
