package replay_out

import (
	"context"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockGameEventWriter is a mock implementation of GameEventWriter
type MockGameEventWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockGameEventWriter) CreateMany(createCtx context.Context, events []*replay_entity.GameEvent) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockGameEventWriter) Create(createCtx context.Context, events *replay_entity.GameEvent) (*replay_entity.GameEvent, error) {
	ret := _m.Called(createCtx, events)

	var r0 *replay_entity.GameEvent
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *replay_entity.GameEvent) (*replay_entity.GameEvent, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *replay_entity.GameEvent) *replay_entity.GameEvent); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*replay_entity.GameEvent)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockGameEventWriter creates a new instance of MockGameEventWriter
func NewMockGameEventWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGameEventWriter {
	mock := &MockGameEventWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
