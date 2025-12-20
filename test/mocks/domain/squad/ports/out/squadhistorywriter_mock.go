package squad_out

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockSquadHistoryWriter is a mock implementation of SquadHistoryWriter
type MockSquadHistoryWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockSquadHistoryWriter) CreateMany(createCtx context.Context, histories []*squad_entities.SquadHistory) error {
	ret := _m.Called(createCtx, histories)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockSquadHistoryWriter) Create(createCtx context.Context, history *squad_entities.SquadHistory) (*squad_entities.SquadHistory, error) {
	ret := _m.Called(createCtx, history)

	var r0 *squad_entities.SquadHistory
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.SquadHistory) (*squad_entities.SquadHistory, error)); ok {
		return rf(createCtx, history)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.SquadHistory) *squad_entities.SquadHistory); ok {
		r0 = rf(createCtx, history)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadHistory)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSquadHistoryWriter creates a new instance of MockSquadHistoryWriter
func NewMockSquadHistoryWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSquadHistoryWriter {
	mock := &MockSquadHistoryWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
