package iam_in

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockCreateRIDTokenCommand is a mock implementation of CreateRIDTokenCommand
type MockCreateRIDTokenCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockCreateRIDTokenCommand) Exec(ctx context.Context, reso common.ResourceOwner, source iam_entities.RIDSourceKey, aud common.IntendedAudienceKey) (*iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, reso, source, aud)

	var r0 *iam_entities.RIDToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.ResourceOwner, iam_entities.RIDSourceKey, common.IntendedAudienceKey) (*iam_entities.RIDToken, error)); ok {
		return rf(ctx, reso, source, aud)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.ResourceOwner, iam_entities.RIDSourceKey, common.IntendedAudienceKey) *iam_entities.RIDToken); ok {
		r0 = rf(ctx, reso, source, aud)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.RIDToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockCreateRIDTokenCommand creates a new instance of MockCreateRIDTokenCommand
func NewMockCreateRIDTokenCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCreateRIDTokenCommand {
	mock := &MockCreateRIDTokenCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
