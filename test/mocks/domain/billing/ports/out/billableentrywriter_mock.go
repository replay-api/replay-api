package billing_out

import (
	"context"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockBillableEntryWriter is a mock implementation of BillableEntryWriter
type MockBillableEntryWriter struct {
	mock.Mock
}

// Create provides a mock function
func (_m *MockBillableEntryWriter) Create(ctx context.Context, billableOperation *billing_entities.BillableEntry) (*billing_entities.BillableEntry, error) {
	ret := _m.Called(ctx, billableOperation)

	var r0 *billing_entities.BillableEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.BillableEntry) (*billing_entities.BillableEntry, error)); ok {
		return rf(ctx, billableOperation)
	}

	if rf, ok := ret.Get(0).(func(context.Context, *billing_entities.BillableEntry) *billing_entities.BillableEntry); ok {
		r0 = rf(ctx, billableOperation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*billing_entities.BillableEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockBillableEntryWriter creates a new instance of MockBillableEntryWriter
func NewMockBillableEntryWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBillableEntryWriter {
	mock := &MockBillableEntryWriter{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
