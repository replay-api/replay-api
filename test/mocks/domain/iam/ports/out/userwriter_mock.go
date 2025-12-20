package iam_out

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockUserWriter is a mock implementation of UserWriter
type MockUserWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockUserWriter) CreateMany(createCtx context.Context, events []*iam_entities.User) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockUserWriter) Create(createCtx context.Context, events *iam_entities.User) (*iam_entities.User, error) {
	ret := _m.Called(createCtx, events)

	var r0 *iam_entities.User
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.User) (*iam_entities.User, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.User) *iam_entities.User); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.User)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockUserWriter creates a new instance of MockUserWriter
func NewMockUserWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserWriter {
	mock := &MockUserWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
