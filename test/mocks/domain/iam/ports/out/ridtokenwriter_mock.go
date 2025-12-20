package iam_out

import (
	"context"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockRIDTokenWriter is a mock implementation of RIDTokenWriter
type MockRIDTokenWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockRIDTokenWriter) Create(ctx context.Context, rid *iam_entities.RIDToken) (*iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, rid)

	var r0 *iam_entities.RIDToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.RIDToken) (*iam_entities.RIDToken, error)); ok {
		return rf(ctx, rid)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.RIDToken) *iam_entities.RIDToken); ok {
		r0 = rf(ctx, rid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.RIDToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Update provides a mock function
func (_m *MockRIDTokenWriter) Update(ctx context.Context, rid *iam_entities.RIDToken) (*iam_entities.RIDToken, error) {
	ret := _m.Called(ctx, rid)

	var r0 *iam_entities.RIDToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.RIDToken) (*iam_entities.RIDToken, error)); ok {
		return rf(ctx, rid)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *iam_entities.RIDToken) *iam_entities.RIDToken); ok {
		r0 = rf(ctx, rid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entities.RIDToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// Delete provides a mock function
func (_m *MockRIDTokenWriter) Delete(ctx context.Context, tokenID string) error {
	ret := _m.Called(ctx, tokenID)

	return ret.Error(0)
}

// Revoke provides a mock function
func (_m *MockRIDTokenWriter) Revoke(ctx context.Context, tokenID string) error {
	ret := _m.Called(ctx, tokenID)

	return ret.Error(0)
}

// NewMockRIDTokenWriter creates a new instance of MockRIDTokenWriter
func NewMockRIDTokenWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRIDTokenWriter {
	mock := &MockRIDTokenWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
