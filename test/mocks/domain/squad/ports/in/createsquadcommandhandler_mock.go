package squad_in

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockCreateSquadCommandHandler is a mock implementation of CreateSquadCommandHandler
type MockCreateSquadCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockCreateSquadCommandHandler) Exec(c context.Context, cmd squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error) {
	ret := _m.Called(c, cmd)

	var r0 *squad_entities.Squad
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error)); ok {
		return rf(c, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.CreateOrUpdatedSquadCommand) *squad_entities.Squad); ok {
		r0 = rf(c, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.Squad)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockCreateSquadCommandHandler creates a new instance of MockCreateSquadCommandHandler
func NewMockCreateSquadCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCreateSquadCommandHandler {
	mock := &MockCreateSquadCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
