package squad_out

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockSquadWriter is a mock implementation of SquadWriter
type MockSquadWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockSquadWriter) CreateMany(createCtx context.Context, events []*squad_entities.Squad) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockSquadWriter) Create(createCtx context.Context, events *squad_entities.Squad) (*squad_entities.Squad, error) {
	ret := _m.Called(createCtx, events)

	var r0 *squad_entities.Squad
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.Squad) (*squad_entities.Squad, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.Squad) *squad_entities.Squad); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.Squad)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockSquadWriter) Update(ctx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error) {
	ret := _m.Called(ctx, squad)

	var r0 *squad_entities.Squad
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.Squad) (*squad_entities.Squad, error)); ok {
		return rf(ctx, squad)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.Squad) *squad_entities.Squad); ok {
		r0 = rf(ctx, squad)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.Squad)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Delete provides a mock function
func (_m *MockSquadWriter) Delete(ctx context.Context, squadID uuid.UUID) error {
	ret := _m.Called(ctx, squadID)

	return ret.Error(0)
}

// NewMockSquadWriter creates a new instance of MockSquadWriter
func NewMockSquadWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSquadWriter {
	mock := &MockSquadWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
