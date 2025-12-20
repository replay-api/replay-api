package squad_in

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockRemoveSquadMemberCommandHandler is a mock implementation of RemoveSquadMemberCommandHandler
type MockRemoveSquadMemberCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockRemoveSquadMemberCommandHandler) Exec(c context.Context, cmd squad_in.RemoveSquadMemberCommand) (*squad_entities.Squad, error) {
	ret := _m.Called(c, cmd)

	var r0 *squad_entities.Squad
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.RemoveSquadMemberCommand) (*squad_entities.Squad, error)); ok {
		return rf(c, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.RemoveSquadMemberCommand) *squad_entities.Squad); ok {
		r0 = rf(c, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.Squad)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockRemoveSquadMemberCommandHandler creates a new instance of MockRemoveSquadMemberCommandHandler
func NewMockRemoveSquadMemberCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRemoveSquadMemberCommandHandler {
	mock := &MockRemoveSquadMemberCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
