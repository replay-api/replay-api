package iam_out

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockProfileWriter is a mock implementation of ProfileWriter
type MockProfileWriter struct {
	mock.Mock
}

// CreateMany provides a mock function
func (_m *MockProfileWriter) CreateMany(createCtx context.Context, events []*iam_entities.Profile) error {
	ret := _m.Called(createCtx, events)

	return ret.Error(0)
}

// Create provides a mock function
func (_m *MockProfileWriter) Create(createCtx context.Context, events *iam_entities.Profile) (*iam_entities.Profile, error) {
	ret := _m.Called(createCtx, events)

	var r0 *iam_entities.Profile
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.Profile) (*iam_entities.Profile, error)); ok {
		return rf(createCtx, events)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.Profile) *iam_entities.Profile); ok {
		r0 = rf(createCtx, events)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.Profile)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockProfileWriter creates a new instance of MockProfileWriter
func NewMockProfileWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProfileWriter {
	mock := &MockProfileWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
