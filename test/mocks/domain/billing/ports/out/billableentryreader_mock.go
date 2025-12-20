package billing_out

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	"github.com/stretchr/testify/mock"
)

// MockBillableEntryReader is a mock implementation of BillableEntryReader
type MockBillableEntryReader struct {
	mock.Mock
}

// GetEntriesBySubscriptionID provides a mock function
func (_m *MockBillableEntryReader) GetEntriesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry, error) {
	ret := _m.Called(ctx, subscriptionID)

	var r0 map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry, error)); ok {
		return rf(ctx, subscriptionID)
	}

	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry); ok {
		r0 = rf(ctx, subscriptionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// NewMockBillableEntryReader creates a new instance of MockBillableEntryReader
func NewMockBillableEntryReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBillableEntryReader {
	mock := &MockBillableEntryReader{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
