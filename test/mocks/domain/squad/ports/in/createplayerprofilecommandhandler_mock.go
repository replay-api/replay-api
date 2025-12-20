package squad_in

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockCreatePlayerProfileCommandHandler is a mock implementation of CreatePlayerProfileCommandHandler
type MockCreatePlayerProfileCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockCreatePlayerProfileCommandHandler) Exec(c context.Context, cmd squad_in.CreatePlayerProfileCommand) (*squad_entities.PlayerProfile, error) {
	ret := _m.Called(c, cmd)

	var r0 *squad_entities.PlayerProfile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.CreatePlayerProfileCommand) (*squad_entities.PlayerProfile, error)); ok {
		return rf(c, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.CreatePlayerProfileCommand) *squad_entities.PlayerProfile); ok {
		r0 = rf(c, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.PlayerProfile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockCreatePlayerProfileCommandHandler creates a new instance of MockCreatePlayerProfileCommandHandler
func NewMockCreatePlayerProfileCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCreatePlayerProfileCommandHandler {
	mock := &MockCreatePlayerProfileCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
