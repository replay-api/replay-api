package billing_in

import (
	"context"

	time "time"

	"github.com/google/uuid"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockAuditTrailCommand is a mock implementation of AuditTrailCommand
type MockAuditTrailCommand struct {
	mock.Mock
}

// RecordFinancialEvent provides a mock function
func (_m *MockAuditTrailCommand) RecordFinancialEvent(ctx context.Context, req billing_in.RecordFinancialEventRequest) error {
	ret := _m.Called(ctx, req)

	return ret.Error(0)
}

// RecordSecurityEvent provides a mock function
func (_m *MockAuditTrailCommand) RecordSecurityEvent(ctx context.Context, req billing_in.RecordSecurityEventRequest) error {
	ret := _m.Called(ctx, req)

	return ret.Error(0)
}

// RecordAdminAction provides a mock function
func (_m *MockAuditTrailCommand) RecordAdminAction(ctx context.Context, req billing_in.RecordAdminActionRequest) error {
	ret := _m.Called(ctx, req)

	return ret.Error(0)
}

// VerifyChainIntegrity provides a mock function
func (_m *MockAuditTrailCommand) VerifyChainIntegrity(ctx context.Context, targetType string, targetID uuid.UUID, from time.Time, to time.Time) (*billing_in.ChainIntegrityResult, error) {
	ret := _m.Called(ctx, targetType, targetID, from, to)

	var r0 *billing_in.ChainIntegrityResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, time.Time, time.Time) (*billing_in.ChainIntegrityResult, error)); ok {
		return rf(ctx, targetType, targetID, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, time.Time, time.Time) *billing_in.ChainIntegrityResult); ok {
		r0 = rf(ctx, targetType, targetID, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_in.ChainIntegrityResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockAuditTrailCommand creates a new instance of MockAuditTrailCommand
func NewMockAuditTrailCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuditTrailCommand {
	mock := &MockAuditTrailCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
