package replay_out

import (
	"context"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/stretchr/testify/mock"
)

// MockEventsByGameReader is a mock implementation of EventsByGameReader
type MockEventsByGameReader struct {
	mock.Mock
}

// GetByGameIDAndMatchID provides a mock function
func (_m *MockEventsByGameReader) GetByGameIDAndMatchID(ctx context.Context, gameID string, matchID string) ([]replay_entity.GameEvent, error) {
	ret := _m.Called(ctx, gameID, matchID)

	var r0 []replay_entity.GameEvent
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]replay_entity.GameEvent, error)); ok {
		return rf(ctx, gameID, matchID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, string, string) []replay_entity.GameEvent); ok {
		r0 = rf(ctx, gameID, matchID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]replay_entity.GameEvent)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockEventsByGameReader creates a new instance of MockEventsByGameReader
func NewMockEventsByGameReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEventsByGameReader {
	mock := &MockEventsByGameReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
