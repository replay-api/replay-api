package squad_out

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockSquadInvitationReader is a mock implementation of SquadInvitationReader
type MockSquadInvitationReader struct {
	mock.Mock
}

// GetByID provides a mock function
func (_m *MockSquadInvitationReader) GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, id)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, id)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// GetBySquadID provides a mock function
func (_m *MockSquadInvitationReader) GetBySquadID(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
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

// GetByPlayerID provides a mock function
func (_m *MockSquadInvitationReader) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
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

// GetPendingBySquadAndPlayer provides a mock function
func (_m *MockSquadInvitationReader) GetPendingBySquadAndPlayer(ctx context.Context, squadID uuid.UUID, playerID uuid.UUID) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, squadID, playerID)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, squadID, playerID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, squadID, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockSquadInvitationReader creates a new instance of MockSquadInvitationReader
func NewMockSquadInvitationReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSquadInvitationReader {
	mock := &MockSquadInvitationReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
