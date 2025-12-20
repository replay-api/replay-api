package squad_in

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockSquadInvitationCommand is a mock implementation of SquadInvitationCommand
type MockSquadInvitationCommand struct {
	mock.Mock
}

// InvitePlayer provides a mock function
func (_m *MockSquadInvitationCommand) InvitePlayer(ctx context.Context, cmd squad_in.InvitePlayerCommand) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.InvitePlayerCommand) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.InvitePlayerCommand) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RequestJoin provides a mock function
func (_m *MockSquadInvitationCommand) RequestJoin(ctx context.Context, cmd squad_in.RequestJoinCommand) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.RequestJoinCommand) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.RequestJoinCommand) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// RespondToInvitation provides a mock function
func (_m *MockSquadInvitationCommand) RespondToInvitation(ctx context.Context, cmd squad_in.RespondToInvitationCommand) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.RespondToInvitationCommand) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, squad_in.RespondToInvitationCommand) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// CancelInvitation provides a mock function
func (_m *MockSquadInvitationCommand) CancelInvitation(ctx context.Context, invitationID uuid.UUID) error {
	ret := _m.Called(ctx, invitationID)

	return ret.Error(0)
}

// GetPendingInvitations provides a mock function
func (_m *MockSquadInvitationCommand) GetPendingInvitations(ctx context.Context, playerID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, playerID)

	var r0 []squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetSquadInvitations provides a mock function
func (_m *MockSquadInvitationCommand) GetSquadInvitations(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, squadID)

	var r0 []squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, squadID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, squadID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSquadInvitationCommand creates a new instance of MockSquadInvitationCommand
func NewMockSquadInvitationCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSquadInvitationCommand {
	mock := &MockSquadInvitationCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
