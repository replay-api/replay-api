package squad_in

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockUpdatePlayerProfileCommandHandler is a mock implementation of UpdatePlayerProfileCommandHandler
type MockUpdatePlayerProfileCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUpdatePlayerProfileCommandHandler) Exec(c context.Context, cmd squad_in.UpdatePlayerCommand) (*squad_entities.PlayerProfile, error) {
	ret := _m.Called(c, cmd)

	var r0 *squad_entities.PlayerProfile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.UpdatePlayerCommand) (*squad_entities.PlayerProfile, error)); ok {
		return rf(c, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.UpdatePlayerCommand) *squad_entities.PlayerProfile); ok {
		r0 = rf(c, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.PlayerProfile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUpdatePlayerProfileCommandHandler creates a new instance of MockUpdatePlayerProfileCommandHandler
func NewMockUpdatePlayerProfileCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUpdatePlayerProfileCommandHandler {
	mock := &MockUpdatePlayerProfileCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
