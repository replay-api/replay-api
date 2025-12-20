package billing_out

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockWalletReader is a mock implementation of WalletReader
type MockWalletReader struct {
	mock.Mock
}

// GetBalance provides a mock function
func (_m *MockWalletReader) GetBalance(ctx context.Context, userID uuid.UUID, currency string) (float64, error) {
	ret := _m.Called(ctx, userID, currency)

	var r0 float64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (float64, error)); ok {
		return rf(ctx, userID, currency)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) float64); ok {
		r0 = rf(ctx, userID, currency)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(float64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByID provides a mock function
func (_m *MockWalletReader) GetByID(ctx context.Context, walletID uuid.UUID) (interface{}, error) {
	ret := _m.Called(ctx, walletID)

	var r0 interface{}
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (interface{}, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) interface{}); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockWalletReader creates a new instance of MockWalletReader
func NewMockWalletReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWalletReader {
	mock := &MockWalletReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
