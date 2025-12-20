package payment_in

import (
	"context"

	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockPaymentQuery is a mock implementation of PaymentQuery
type MockPaymentQuery struct {
	mock.Mock
}

// GetPayment provides a mock function
func (_m *MockPaymentQuery) GetPayment(ctx context.Context, query payment_in.GetPaymentQuery) (*payment_in.PaymentDTO, error) {
	ret := _m.Called(ctx, query)

	var r0 *payment_in.PaymentDTO
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.GetPaymentQuery) (*payment_in.PaymentDTO, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.GetPaymentQuery) *payment_in.PaymentDTO); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_in.PaymentDTO)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetUserPayments provides a mock function
func (_m *MockPaymentQuery) GetUserPayments(ctx context.Context, query payment_in.GetUserPaymentsQuery) (*payment_in.PaymentsResult, error) {
	ret := _m.Called(ctx, query)

	var r0 *payment_in.PaymentsResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.GetUserPaymentsQuery) (*payment_in.PaymentsResult, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.GetUserPaymentsQuery) *payment_in.PaymentsResult); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_in.PaymentsResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPaymentQuery creates a new instance of MockPaymentQuery
func NewMockPaymentQuery(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPaymentQuery {
	mock := &MockPaymentQuery{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
