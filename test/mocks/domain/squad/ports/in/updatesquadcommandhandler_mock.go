package squad_in

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockUpdateSquadCommandHandler is a mock implementation of UpdateSquadCommandHandler
type MockUpdateSquadCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockUpdateSquadCommandHandler) Exec(c context.Context, squadID uuid.UUID, cmd squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error) {
	ret := _m.Called(c, squadID, cmd)

	var r0 *squad_entities.Squad
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error)); ok {
		return rf(c, squadID, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, squad_in.CreateOrUpdatedSquadCommand) *squad_entities.Squad); ok {
		r0 = rf(c, squadID, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.Squad)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUpdateSquadCommandHandler creates a new instance of MockUpdateSquadCommandHandler
func NewMockUpdateSquadCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUpdateSquadCommandHandler {
	mock := &MockUpdateSquadCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
