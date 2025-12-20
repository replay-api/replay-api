package iam_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_dtos "github.com/replay-api/replay-api/pkg/domain/iam/dtos"
	"github.com/stretchr/testify/mock"
)

// MockMembershipReader is a mock implementation of MembershipReader
type MockMembershipReader struct {
	mock.Mock
}

// ListMemberGroups provides a mock function
func (_m *MockMembershipReader) ListMemberGroups(ctx context.Context, s *common.Search) (map[uuid.UUID]iam_dtos.GroupMembershipDTO, error) {
	ret := _m.Called(ctx, s)

	var r0 map[uuid.UUID]iam_dtos.GroupMembershipDTO
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *common.Search) (map[uuid.UUID]iam_dtos.GroupMembershipDTO, error)); ok {
		return rf(ctx, s)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *common.Search) map[uuid.UUID]iam_dtos.GroupMembershipDTO); ok {
		r0 = rf(ctx, s)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[uuid.UUID]iam_dtos.GroupMembershipDTO)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockMembershipReader creates a new instance of MockMembershipReader
func NewMockMembershipReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMembershipReader {
	mock := &MockMembershipReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
