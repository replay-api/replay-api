package iam_out

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockMembershipWriter is a mock implementation of MembershipWriter
type MockMembershipWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockMembershipWriter) CreateMany(createCtx context.Context, events []*iam_entities.Membership) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockMembershipWriter) Create(createCtx context.Context, events *iam_entities.Membership) (*iam_entities.Membership, error) {
	ret := _m.Called(createCtx, events)

	var r0 *iam_entities.Membership
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.Membership) (*iam_entities.Membership, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.Membership) *iam_entities.Membership); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.Membership)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMembershipWriter creates a new instance of MockMembershipWriter
func NewMockMembershipWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMembershipWriter {
	mock := &MockMembershipWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
