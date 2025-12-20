package squad_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockPlayerStatisticsReader is a mock implementation of PlayerStatisticsReader
type MockPlayerStatisticsReader struct {
	mock.Mock
}

// GetPlayerStatistics provides a mock function
func (_m *MockPlayerStatisticsReader) GetPlayerStatistics(ctx context.Context, playerID uuid.UUID, gameID *common.GameIDKey) (*squad_entities.PlayerStatistics, error) {
	ret := _m.Called(ctx, playerID, gameID)

	var r0 *squad_entities.PlayerStatistics
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *common.GameIDKey) (*squad_entities.PlayerStatistics, error)); ok {
		return rf(ctx, playerID, gameID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *common.GameIDKey) *squad_entities.PlayerStatistics); ok {
		r0 = rf(ctx, playerID, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.PlayerStatistics)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockPlayerStatisticsReader creates a new instance of MockPlayerStatisticsReader
func NewMockPlayerStatisticsReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPlayerStatisticsReader {
	mock := &MockPlayerStatisticsReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
