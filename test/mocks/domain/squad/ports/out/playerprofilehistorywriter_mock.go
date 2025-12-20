package squad_out

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockPlayerProfileHistoryWriter is a mock implementation of PlayerProfileHistoryWriter
type MockPlayerProfileHistoryWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockPlayerProfileHistoryWriter) CreateMany(createCtx context.Context, histories []*squad_entities.PlayerProfileHistory) error {
	ret := _m.Called(createCtx, histories)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockPlayerProfileHistoryWriter) Create(createCtx context.Context, history *squad_entities.PlayerProfileHistory) (*squad_entities.PlayerProfileHistory, error) {
	ret := _m.Called(createCtx, history)

	var r0 *squad_entities.PlayerProfileHistory
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.PlayerProfileHistory) (*squad_entities.PlayerProfileHistory, error)); ok {
		return rf(createCtx, history)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.PlayerProfileHistory) *squad_entities.PlayerProfileHistory); ok {
		r0 = rf(createCtx, history)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.PlayerProfileHistory)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPlayerProfileHistoryWriter creates a new instance of MockPlayerProfileHistoryWriter
func NewMockPlayerProfileHistoryWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerProfileHistoryWriter {
	mock := &MockPlayerProfileHistoryWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
