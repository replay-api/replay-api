package iam_out

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entity "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/stretchr/testify/mock"
)

// MockRIDTokenReader is a mock implementation of RIDTokenReader
type MockRIDTokenReader struct {
	mock.Mock
}

// Search provides a mock function
func (_m *MockRIDTokenReader) Search(ctx context.Context, s common.Search) ([]iam_entity.RIDToken, error) {
	ret := _m.Called(ctx, s)

	var r0 []iam_entity.RIDToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) ([]iam_entity.RIDToken, error)); ok {
		return rf(ctx, s)
	}

	if rf, ok := ret.Get(0).(func(context.Context, common.Search) []iam_entity.RIDToken); ok {
		r0 = rf(ctx, s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iam_entity.RIDToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// FindByID provides a mock function
func (_m *MockRIDTokenReader) FindByID(ctx context.Context, tokenID uuid.UUID) (*iam_entity.RIDToken, error) {
	ret := _m.Called(ctx, tokenID)

	var r0 *iam_entity.RIDToken
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*iam_entity.RIDToken, error)); ok {
		return rf(ctx, tokenID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *iam_entity.RIDToken); ok {
		r0 = rf(ctx, tokenID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iam_entity.RIDToken)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockRIDTokenReader creates a new instance of MockRIDTokenReader
func NewMockRIDTokenReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRIDTokenReader {
	mock := &MockRIDTokenReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
