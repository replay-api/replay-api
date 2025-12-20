package custody_in

import (
	"context"

	"time"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_in "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockRecoveryService is a mock implementation of RecoveryService
type MockRecoveryService struct {
	mock.Mock
}

// AddGuardian provides a mock function
func (_m *MockRecoveryService) AddGuardian(ctx context.Context, req *custody_in.AddGuardianRequest) (*custody_in.GuardianResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.GuardianResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.AddGuardianRequest) (*custody_in.GuardianResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.AddGuardianRequest) *custody_in.GuardianResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.GuardianResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RemoveGuardian provides a mock function
func (_m *MockRecoveryService) RemoveGuardian(ctx context.Context, walletID uuid.UUID, guardianID uuid.UUID) error {
	ret := _m.Called(ctx, walletID, guardianID)

	return ret.Error(0)
}

// GetGuardians provides a mock function
func (_m *MockRecoveryService) GetGuardians(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.Guardian, error) {
	ret := _m.Called(ctx, walletID)

	var r0 []*custody_entities.Guardian
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]*custody_entities.Guardian, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*custody_entities.Guardian); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_entities.Guardian)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SetGuardianThreshold provides a mock function
func (_m *MockRecoveryService) SetGuardianThreshold(ctx context.Context, walletID uuid.UUID, threshold uint8) error {
	ret := _m.Called(ctx, walletID, threshold)

	return ret.Error(0)
}

// InitiateRecovery provides a mock function
func (_m *MockRecoveryService) InitiateRecovery(ctx context.Context, req *custody_in.InitiateRecoveryRequest) (*custody_in.RecoveryInitResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.RecoveryInitResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.InitiateRecoveryRequest) (*custody_in.RecoveryInitResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.InitiateRecoveryRequest) *custody_in.RecoveryInitResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.RecoveryInitResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ApproveRecovery provides a mock function
func (_m *MockRecoveryService) ApproveRecovery(ctx context.Context, req *custody_in.ApproveRecoveryRequest) (*custody_in.RecoveryApprovalResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *custody_in.RecoveryApprovalResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.ApproveRecoveryRequest) (*custody_in.RecoveryApprovalResult, error)); ok {
		return rf(ctx, req)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *custody_in.ApproveRecoveryRequest) *custody_in.RecoveryApprovalResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.RecoveryApprovalResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ExecuteRecovery provides a mock function
func (_m *MockRecoveryService) ExecuteRecovery(ctx context.Context, walletID uuid.UUID) (*custody_in.RecoveryExecutionResult, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *custody_in.RecoveryExecutionResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_in.RecoveryExecutionResult, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_in.RecoveryExecutionResult); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.RecoveryExecutionResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelRecovery provides a mock function
func (_m *MockRecoveryService) CancelRecovery(ctx context.Context, walletID uuid.UUID) error {
	ret := _m.Called(ctx, walletID)

	return ret.Error(0)
}

// GetRecoveryStatus provides a mock function
func (_m *MockRecoveryService) GetRecoveryStatus(ctx context.Context, walletID uuid.UUID) (*custody_in.RecoveryStatus, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *custody_in.RecoveryStatus
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_in.RecoveryStatus, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_in.RecoveryStatus); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.RecoveryStatus)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetPendingRecoveries provides a mock function
func (_m *MockRecoveryService) GetPendingRecoveries(ctx context.Context, guardianAddress string) ([]*custody_in.PendingRecoveryInfo, error) {
	ret := _m.Called(ctx, guardianAddress)

	var r0 []*custody_in.PendingRecoveryInfo
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*custody_in.PendingRecoveryInfo, error)); ok {
		return rf(ctx, guardianAddress)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string) []*custody_in.PendingRecoveryInfo); ok {
		r0 = rf(ctx, guardianAddress)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custody_in.PendingRecoveryInfo)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SetRecoveryDelay provides a mock function
func (_m *MockRecoveryService) SetRecoveryDelay(ctx context.Context, walletID uuid.UUID, delay time.Duration) error {
	ret := _m.Called(ctx, walletID, delay)

	return ret.Error(0)
}

// GetRecoveryConfig provides a mock function
func (_m *MockRecoveryService) GetRecoveryConfig(ctx context.Context, walletID uuid.UUID) (*custody_in.RecoveryConfig, error) {
	ret := _m.Called(ctx, walletID)

	var r0 *custody_in.RecoveryConfig
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*custody_in.RecoveryConfig, error)); ok {
		return rf(ctx, walletID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *custody_in.RecoveryConfig); ok {
		r0 = rf(ctx, walletID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custody_in.RecoveryConfig)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockRecoveryService creates a new instance of MockRecoveryService
func NewMockRecoveryService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRecoveryService {
	mock := &MockRecoveryService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
