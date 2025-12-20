package squad_out

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	"github.com/stretchr/testify/mock"
)

// MockSquadInvitationWriter is a mock implementation of SquadInvitationWriter
type MockSquadInvitationWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockSquadInvitationWriter) Create(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, invitation)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, invitation)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.SquadInvitation) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, invitation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockSquadInvitationWriter) Update(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error) {
	ret := _m.Called(ctx, invitation)

	var r0 *squad_entities.SquadInvitation
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error)); ok {
		return rf(ctx, invitation)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *squad_entities.SquadInvitation) *squad_entities.SquadInvitation); ok {
		r0 = rf(ctx, invitation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*squad_entities.SquadInvitation)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Delete provides a mock function
func (_m *MockSquadInvitationWriter) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	return ret.Error(0)
}

// NewMockSquadInvitationWriter creates a new instance of MockSquadInvitationWriter
func NewMockSquadInvitationWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSquadInvitationWriter {
	mock := &MockSquadInvitationWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
