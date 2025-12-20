package custody_in

import (
	"context"

	"github.com/google/uuid"
	in_custody "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockEmergencyService is a mock implementation of EmergencyService
type MockEmergencyService struct {
	mock.Mock
}

// EmergencyFreeze provides a mock function
func (_m *MockEmergencyService) EmergencyFreeze(ctx context.Context, req *in_custody.EmergencyFreezeRequest) (*in_custody.EmergencyResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *in_custody.EmergencyResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.EmergencyFreezeRequest) (*in_custody.EmergencyResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.EmergencyFreezeRequest) *in_custody.EmergencyResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.EmergencyResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// EmergencyUnfreeze provides a mock function
func (_m *MockEmergencyService) EmergencyUnfreeze(ctx context.Context, req *in_custody.EmergencyUnfreezeRequest) (*in_custody.EmergencyResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *in_custody.EmergencyResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.EmergencyUnfreezeRequest) (*in_custody.EmergencyResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.EmergencyUnfreezeRequest) *in_custody.EmergencyResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.EmergencyResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// EmergencyTransfer provides a mock function
func (_m *MockEmergencyService) EmergencyTransfer(ctx context.Context, req *in_custody.EmergencyTransferRequest) (*in_custody.	EmergencyTransferResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *in_custody.EmergencyTransferResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.EmergencyTransferRequest) (*in_custody.EmergencyTransferResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *in_custody.EmergencyTransferRequest) *in_custody.EmergencyTransferResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.EmergencyTransferResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetEmergencyStatus provides a mock function
func (_m *MockEmergencyService) GetEmergencyStatus(ctx context.Context, walletID uuid.UUID) (*in_custody.EmergencyStatus, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *in_custody.EmergencyStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*in_custody.EmergencyStatus, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *in_custody.EmergencyStatus); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*in_custody.EmergencyStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockEmergencyService creates a new instance of MockEmergencyService
func NewMockEmergencyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEmergencyService {
	mock := &MockEmergencyService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
