package iam_out

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockGroupWriter is a mock implementation of GroupWriter
type MockGroupWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockGroupWriter) CreateMany(createCtx context.Context, events []*iam_entities.Group) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockGroupWriter) Create(createCtx context.Context, events *iam_entities.Group) (*iam_entities.Group, error) {
	ret := _m.Called(createCtx, events)

	var r0 *iam_entities.Group
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.Group) (*iam_entities.Group, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.Group) *iam_entities.Group); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.Group)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockGroupWriter creates a new instance of MockGroupWriter
func NewMockGroupWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGroupWriter {
	mock := &MockGroupWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
