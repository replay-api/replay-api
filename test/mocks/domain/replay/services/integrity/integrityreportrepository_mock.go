package integrity

import (
	"context"

	"github.com/google/uuid"
	integrity_service "github.com/replay-api/replay-api/pkg/domain/replay/services/integrity"
	"github.com/stretchr/testify/mock"
)

// MockIntegrityReportRepository is a mock implementation of IntegrityReportRepository
type MockIntegrityReportRepository struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockIntegrityReportRepository) Create(ctx context.Context, report *integrity_service.ReplayIntegrityReport) error {
	ret := _m.Called(ctx, report)

	return ret.Error(0)
}

// GetByReplayID provides a mock function
func (_m *MockIntegrityReportRepository) GetByReplayID(ctx context.Context, replayID uuid.UUID) (*integrity_service.ReplayIntegrityReport, error) {
	ret := _m.Called(ctx, replayID)

	var r0 *integrity_service.ReplayIntegrityReport
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*integrity_service.ReplayIntegrityReport, error)); ok {
		return rf(ctx, replayID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *integrity_service.ReplayIntegrityReport); ok {
		r0 = rf(ctx, replayID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*integrity_service.ReplayIntegrityReport)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockIntegrityReportRepository) Update(ctx context.Context, report *integrity_service.ReplayIntegrityReport) error {
	ret := _m.Called(ctx, report)

	return ret.Error(0)
}

// GetPendingReviews provides a mock function
func (_m *MockIntegrityReportRepository) GetPendingReviews(ctx context.Context, limit int) ([]integrity_service.ReplayIntegrityReport, error) {
	ret := _m.Called(ctx, limit)

	var r0 []integrity_service.ReplayIntegrityReport
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, int) ([]integrity_service.ReplayIntegrityReport, error)); ok {
		return rf(ctx, limit)
	}

	if rf, ok := ret.Get(0).(func(context.Context, int) []integrity_service.ReplayIntegrityReport); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]integrity_service.ReplayIntegrityReport)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockIntegrityReportRepository creates a new instance of MockIntegrityReportRepository
func NewMockIntegrityReportRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIntegrityReportRepository {
	mock := &MockIntegrityReportRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
