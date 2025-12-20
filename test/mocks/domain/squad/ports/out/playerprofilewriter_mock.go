package squad_out

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockPlayerProfileWriter is a mock implementation of PlayerProfileWriter
type MockPlayerProfileWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockPlayerProfileWriter) CreateMany(createCtx context.Context, events []*squad_entities.PlayerProfile) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockPlayerProfileWriter) Create(createCtx context.Context, events *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error) {
	ret := _m.Called(createCtx, events)

	var r0 *squad_entities.PlayerProfile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.PlayerProfile) *squad_entities.PlayerProfile); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.PlayerProfile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockPlayerProfileWriter) Update(ctx context.Context, profile *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error) {
	ret := _m.Called(ctx, profile)

	var r0 *squad_entities.PlayerProfile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)); ok {
		return rf(ctx, profile)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.PlayerProfile) *squad_entities.PlayerProfile); ok {
		r0 = rf(ctx, profile)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.PlayerProfile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Delete provides a mock function
func (_m *MockPlayerProfileWriter) Delete(ctx context.Context, profileID uuid.UUID) error {
	ret := _m.Called(ctx, profileID)

	return ret.Error(0)
}

// NewMockPlayerProfileWriter creates a new instance of MockPlayerProfileWriter
func NewMockPlayerProfileWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerProfileWriter {
	mock := &MockPlayerProfileWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
