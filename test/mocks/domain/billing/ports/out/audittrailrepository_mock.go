package billing_out

import (
	"context"

	time "time"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockAuditTrailRepository is a mock implementation of AuditTrailRepository
type MockAuditTrailRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockAuditTrailRepository) Create(ctx context.Context, entry *billing_entities.AuditTrailEntry) error {
	ret := _m.Called(ctx, entry)

	return ret.Error(0)
}

// GetByID provides a mock function
func (_m *MockAuditTrailRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, id)

	var r0 *billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetLatestForTarget provides a mock function
func (_m *MockAuditTrailRepository) GetLatestForTarget(ctx context.Context, targetType string, targetID uuid.UUID) (*billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, targetType, targetID)

	var r0 *billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) (*billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, targetType, targetID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) *billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, targetType, targetID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByUser provides a mock function
func (_m *MockAuditTrailRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, userID, limit, offset)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, userID, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, int, int) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, userID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByTarget provides a mock function
func (_m *MockAuditTrailRepository) GetByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit int, offset int) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, targetType, targetID, limit, offset)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, int, int) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, targetType, targetID, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, int, int) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, targetType, targetID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByEventType provides a mock function
func (_m *MockAuditTrailRepository) GetByEventType(ctx context.Context, eventType billing_entities.AuditEventType, from time.Time, to time.Time, limit int, offset int) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, eventType, from, to, limit, offset)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.AuditEventType, time.Time, time.Time, int, int) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, eventType, from, to, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.AuditEventType, time.Time, time.Time, int, int) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, eventType, from, to, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetBySeverity provides a mock function
func (_m *MockAuditTrailRepository) GetBySeverity(ctx context.Context, severity billing_entities.AuditSeverity, from time.Time, to time.Time, limit int, offset int) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, severity, from, to, limit, offset)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.AuditSeverity, time.Time, time.Time, int, int) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, severity, from, to, limit, offset)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.AuditSeverity, time.Time, time.Time, int, int) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, severity, from, to, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetChainForVerification provides a mock function
func (_m *MockAuditTrailRepository) GetChainForVerification(ctx context.Context, targetType string, targetID uuid.UUID, from time.Time, to time.Time) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, targetType, targetID, from, to)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, time.Time, time.Time) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, targetType, targetID, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, time.Time, time.Time) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, targetType, targetID, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetForComplianceReport provides a mock function
func (_m *MockAuditTrailRepository) GetForComplianceReport(ctx context.Context, from time.Time, to time.Time) ([]billing_entities.AuditTrailEntry, error) {
	ret := _m.Called(ctx, from, to)

	var r0 []billing_entities.AuditTrailEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time) ([]billing_entities.AuditTrailEntry, error)); ok {
		return rf(ctx, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time) []billing_entities.AuditTrailEntry); ok {
		r0 = rf(ctx, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]billing_entities.AuditTrailEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CountByType provides a mock function
func (_m *MockAuditTrailRepository) CountByType(ctx context.Context, eventType billing_entities.AuditEventType, from time.Time, to time.Time) (int64, error) {
	ret := _m.Called(ctx, eventType, from, to)

	var r0 int64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.AuditEventType, time.Time, time.Time) (int64, error)); ok {
		return rf(ctx, eventType, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, billing_entities.AuditEventType, time.Time, time.Time) int64); ok {
		r0 = rf(ctx, eventType, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetFinancialSummary provides a mock function
func (_m *MockAuditTrailRepository) GetFinancialSummary(ctx context.Context, userID uuid.UUID, from time.Time, to time.Time) (*billing_entities.AuditSummary, error) {
	ret := _m.Called(ctx, userID, from, to)

	var r0 *billing_entities.AuditSummary
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time, time.Time) (*billing_entities.AuditSummary, error)); ok {
		return rf(ctx, userID, from, to)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Time, time.Time) *billing_entities.AuditSummary); ok {
		r0 = rf(ctx, userID, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.AuditSummary)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// ArchiveOldEntries provides a mock function
func (_m *MockAuditTrailRepository) ArchiveOldEntries(ctx context.Context, before time.Time) (int64, error) {
	ret := _m.Called(ctx, before)

	var r0 int64
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, time.Time) (int64, error)); ok {
		return rf(ctx, before)
	}

	if rf, ok := ret.Get(0).(func(context.Context, time.Time) int64); ok {
		r0 = rf(ctx, before)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int64)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockAuditTrailRepository creates a new instance of MockAuditTrailRepository
func NewMockAuditTrailRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuditTrailRepository {
	mock := &MockAuditTrailRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
