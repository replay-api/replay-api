package squad_in

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDeleteSquadCommandHandler is a mock implementation of DeleteSquadCommandHandler
type MockDeleteSquadCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockDeleteSquadCommandHandler) Exec(c context.Context, squadID uuid.UUID) error {
	ret := _m.Called(c, squadID)

	return ret.Error(0)
}

// NewMockDeleteSquadCommandHandler creates a new instance of MockDeleteSquadCommandHandler
func NewMockDeleteSquadCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDeleteSquadCommandHandler {
	mock := &MockDeleteSquadCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
