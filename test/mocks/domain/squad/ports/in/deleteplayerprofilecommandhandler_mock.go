package squad_in

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDeletePlayerProfileCommandHandler is a mock implementation of DeletePlayerProfileCommandHandler
type MockDeletePlayerProfileCommandHandler struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockDeletePlayerProfileCommandHandler) Exec(c context.Context, playerID uuid.UUID) error {
	ret := _m.Called(c, playerID)

	return ret.Error(0)
}

// NewMockDeletePlayerProfileCommandHandler creates a new instance of MockDeletePlayerProfileCommandHandler
func NewMockDeletePlayerProfileCommandHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDeletePlayerProfileCommandHandler {
	mock := &MockDeletePlayerProfileCommandHandler{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
