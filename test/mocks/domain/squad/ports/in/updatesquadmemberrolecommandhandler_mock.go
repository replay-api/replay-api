package squad_in

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockUpdateSquadMemberRoleCommandHandler is a mock implementation of UpdateSquadMemberRoleCommandHandler
type MockUpdateSquadMemberRoleCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUpdateSquadMemberRoleCommandHandler) Exec(c context.Context, cmd squad_in.UpdateSquadMemberRoleCommand) (*squad_entities.Squad, error) {
	ret := _m.Called(c, cmd)

	var r0 *squad_entities.Squad
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.UpdateSquadMemberRoleCommand) (*squad_entities.Squad, error)); ok {
		return rf(c, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.UpdateSquadMemberRoleCommand) *squad_entities.Squad); ok {
		r0 = rf(c, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.Squad)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUpdateSquadMemberRoleCommandHandler creates a new instance of MockUpdateSquadMemberRoleCommandHandler
func NewMockUpdateSquadMemberRoleCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUpdateSquadMemberRoleCommandHandler {
	mock := &MockUpdateSquadMemberRoleCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
