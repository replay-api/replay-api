package billing_out

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockPlanReader is a mock implementation of PlanReader
type MockPlanReader struct {
	mock.Mock
}

// GetDefaultFreePlan provides a mock function
func (_m *MockPlanReader) GetDefaultFreePlan(ctx context.Context) (*billing_entities.Plan, error) {
	ret := _m.Called(ctx)

	var r0 *billing_entities.Plan
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) (*billing_entities.Plan, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) *billing_entities.Plan); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Plan)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetByID provides a mock function
func (_m *MockPlanReader) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Plan, error) {
	ret := _m.Called(ctx, id)

	var r0 *billing_entities.Plan
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*billing_entities.Plan, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *billing_entities.Plan); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.Plan)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetAvailablePlans provides a mock function
func (_m *MockPlanReader) GetAvailablePlans(ctx context.Context) ([]*billing_entities.Plan, error) {
	ret := _m.Called(ctx)

	var r0 []*billing_entities.Plan
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context) ([]*billing_entities.Plan, error)); ok {
		return rf(ctx)
	}

	if rf, ok := ret.Get(0).(func(context.Context) []*billing_entities.Plan); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*billing_entities.Plan)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPlanReader creates a new instance of MockPlanReader
func NewMockPlanReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlanReader {
	mock := &MockPlanReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
