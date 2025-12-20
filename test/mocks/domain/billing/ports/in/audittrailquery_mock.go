package billing_in

import (
	"context"

	time "time"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockAuditTrailQuery is a mock implementation of AuditTrailQuery
type MockAuditTrailQuery struct {
	mock.Mock
}

// GetUserAuditHistory provides a mock function
func (_m *MockAuditTrailQuery) GetUserAuditHistory(ctx context.Context, userID uuid.UUID, filters billing_in.AuditFilters) (*billing_in.AuditHistoryResult, error) {
	ret := _m.Called(ctx, userID, filters)

	var r0 *billing_in.AuditHistoryResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, billing_in.AuditFilters) (*billing_in.AuditHistoryResult, error)); ok {
		return rf(ctx, userID, filters)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, billing_in.AuditFilters) *billing_in.AuditHistoryResult); ok {
		r0 = rf(ctx, userID, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_in.AuditHistoryResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetTransactionAudit provides a mock function
func (_m *MockAuditTrailQuery) GetTransactionAudit(ctx context.Context, transactionID uuid.UUID) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, transactionID)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, transactionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, transactionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GenerateComplianceReport provides a mock function
func (_m *MockAuditTrailQuery) GenerateComplianceReport(ctx context.Context, reportType string, from time.Time, to time.Time) (*billing_entities.ComplianceReport, error) {
	ret := _m.Called(ctx, reportType, from, to)

	var r0 *billing_entities.ComplianceReport
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, time.Time) (*billing_entities.ComplianceReport, error)); ok {
		return rf(ctx, reportType, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, time.Time) *billing_entities.ComplianceReport); ok {
		r0 = rf(ctx, reportType, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.ComplianceReport)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAuditSummary provides a mock function
func (_m *MockAuditTrailQuery) GetAuditSummary(ctx context.Context, userID uuid.UUID, period string) (*billing_entities.AuditSummary, error) {
	ret := _m.Called(ctx, userID, period)

	var r0 *billing_entities.AuditSummary
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) (*billing_entities.AuditSummary, error)); ok {
		return rf(ctx, userID, period)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *billing_entities.AuditSummary); ok {
		r0 = rf(ctx, userID, period)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.AuditSummary)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// SearchAudit provides a mock function
func (_m *MockAuditTrailQuery) SearchAudit(ctx context.Context, query billing_in.AuditSearchQuery) (*billing_in.AuditSearchResult, error) {
	ret := _m.Called(ctx, query)

	var r0 *billing_in.AuditSearchResult
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_in.AuditSearchQuery) (*billing_in.AuditSearchResult, error)); ok {
		return rf(ctx, query)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_in.AuditSearchQuery) *billing_in.AuditSearchResult); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_in.AuditSearchResult)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ExportAudit provides a mock function
func (_m *MockAuditTrailQuery) ExportAudit(ctx context.Context, from time.Time, to time.Time, format string) ([]byte, error) {
	ret := _m.Called(ctx, from, to, format)

	var r0 []byte
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, string) ([]byte, error)); ok {
		return rf(ctx, from, to, format)
	}

	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, string) []byte); ok {
		r0 = rf(ctx, from, to, format)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockAuditTrailQuery creates a new instance of MockAuditTrailQuery
func NewMockAuditTrailQuery(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuditTrailQuery {
	mock := &MockAuditTrailQuery{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
