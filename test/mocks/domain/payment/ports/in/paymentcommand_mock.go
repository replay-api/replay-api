package payment_in

import (
	"context"

	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockPaymentCommand is a mock implementation of PaymentCommand
type MockPaymentCommand struct {
	mock.Mock
}

// CreatePaymentIntent provides a mock function
func (_m *MockPaymentCommand) CreatePaymentIntent(ctx context.Context, cmd payment_in.CreatePaymentIntentCommand) (*payment_in.PaymentIntentResult, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *payment_in.PaymentIntentResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.CreatePaymentIntentCommand) (*payment_in.PaymentIntentResult, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.CreatePaymentIntentCommand) *payment_in.PaymentIntentResult); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_in.PaymentIntentResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ConfirmPayment provides a mock function
func (_m *MockPaymentCommand) ConfirmPayment(ctx context.Context, cmd payment_in.ConfirmPaymentCommand) (*payment_entities.Payment, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.ConfirmPaymentCommand) (*payment_entities.Payment, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.ConfirmPaymentCommand) *payment_entities.Payment); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RefundPayment provides a mock function
func (_m *MockPaymentCommand) RefundPayment(ctx context.Context, cmd payment_in.RefundPaymentCommand) (*payment_entities.Payment, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.RefundPaymentCommand) (*payment_entities.Payment, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.RefundPaymentCommand) *payment_entities.Payment); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelPayment provides a mock function
func (_m *MockPaymentCommand) CancelPayment(ctx context.Context, cmd payment_in.CancelPaymentCommand) (*payment_entities.Payment, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *payment_entities.Payment
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.CancelPaymentCommand) (*payment_entities.Payment, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, payment_in.CancelPaymentCommand) *payment_entities.Payment); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payment_entities.Payment)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ProcessWebhook provides a mock function
func (_m *MockPaymentCommand) ProcessWebhook(ctx context.Context, cmd payment_in.ProcessWebhookCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// NewMockPaymentCommand creates a new instance of MockPaymentCommand
func NewMockPaymentCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPaymentCommand {
	mock := &MockPaymentCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
