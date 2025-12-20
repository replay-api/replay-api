package iam_in

import (
	"context"

	"github.com/google/uuid"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockRefreshRIDTokenCommand is a mock implementation of RefreshRIDTokenCommand
type MockRefreshRIDTokenCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockRefreshRIDTokenCommand) Exec(ctx context.Context, tokenID uuid.UUID) (*iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, tokenID)

	var r0 *iam_entities.RIDToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*iam_entities.RIDToken, error)); ok {
		return rf(ctx, tokenID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *iam_entities.RIDToken); ok {
		r0 = rf(ctx, tokenID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.RIDToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockRefreshRIDTokenCommand creates a new instance of MockRefreshRIDTokenCommand
func NewMockRefreshRIDTokenCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRefreshRIDTokenCommand {
	mock := &MockRefreshRIDTokenCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
