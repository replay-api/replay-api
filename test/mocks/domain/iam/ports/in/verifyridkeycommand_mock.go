package iam_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	"github.com/stretchr/testify/mock"
)

// MockVerifyRIDKeyCommand is a mock implementation of VerifyRIDKeyCommand
type MockVerifyRIDKeyCommand struct {
	mock.Mock
}

// Exec provides a mock function
func (_m *MockVerifyRIDKeyCommand) Exec(ctx context.Context, key uuid.UUID) (common.ResourceOwner, common.IntendedAudienceKey, error) {
	ret := _m.Called(ctx, key)

	var r0 common.ResourceOwner
	var r1 common.IntendedAudienceKey
	var r2 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (common.ResourceOwner, common.IntendedAudienceKey, error)); ok {
		return rf(ctx, key)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) common.ResourceOwner); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.ResourceOwner)
		}
	}
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) common.IntendedAudienceKey); ok {
		r1 = rf(ctx, key)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(common.IntendedAudienceKey)
		}
	}
	r2 = ret.Error(2)

	return r0, r1, r2
}

// NewMockVerifyRIDKeyCommand creates a new instance of MockVerifyRIDKeyCommand
func NewMockVerifyRIDKeyCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockVerifyRIDKeyCommand {
	mock := &MockVerifyRIDKeyCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
