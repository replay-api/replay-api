package billing_usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_usecases "github.com/replay-api/replay-api/pkg/domain/billing/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionReader implements billing_out.SubscriptionReader
type MockSubscriptionReader struct {
	mock.Mock
}

func (m *MockSubscriptionReader) GetCurrentSubscription(ctx context.Context, owner shared.ResourceOwner) (*billing_entities.Subscription, error) {
	args := m.Called(ctx, owner)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billing_entities.Subscription), args.Error(1)
}

// MockSubscriptionWriter implements billing_out.SubscriptionWriter
type MockSubscriptionWriter struct {
	mock.Mock
}

func (m *MockSubscriptionWriter) Create(ctx context.Context, sub *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	args := m.Called(ctx, sub)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billing_entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionWriter) Update(ctx context.Context, sub *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	args := m.Called(ctx, sub)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billing_entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionWriter) Cancel(ctx context.Context, sub *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	args := m.Called(ctx, sub)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billing_entities.Subscription), args.Error(1)
}

// MockPlanReader implements billing_out.PlanReader
type MockPlanReader struct {
	mock.Mock
}

func (m *MockPlanReader) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Plan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billing_entities.Plan), args.Error(1)
}

func (m *MockPlanReader) GetDefaultFreePlan(ctx context.Context) (*billing_entities.Plan, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billing_entities.Plan), args.Error(1)
}

func (m *MockPlanReader) GetAvailablePlans(ctx context.Context) ([]*billing_entities.Plan, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*billing_entities.Plan), args.Error(1)
}

func (m *MockPlanReader) Search(ctx context.Context, s shared.Search) ([]billing_entities.Plan, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]billing_entities.Plan), args.Error(1)
}

func (m *MockPlanReader) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

// Helper function to create authenticated context with user
func createAuthenticatedContext(userID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	return ctx
}

func TestUpgradeSubscriptionUseCase_Exec_Success(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)
	resourceOwner := shared.GetResourceOwner(ctx)

	currentPlanID := uuid.New()
	targetPlanID := uuid.New()

	currentSub := &billing_entities.Subscription{
		BaseEntity:    shared.NewEntity(resourceOwner),
		PlanID:        currentPlanID,
		BillingPeriod: billing_entities.BillingPeriodMonthly,
		Status:        billing_entities.SubscriptionStatusActive,
		IsFree:        false,
		History:       []billing_entities.SubscriptionHistory{},
	}

	currentPlan := &billing_entities.Plan{
		BaseEntity: shared.NewEntity(resourceOwner),
		Kind:       billing_entities.PlanKindTypeStarter,
	}
	targetPlan := &billing_entities.Plan{
		BaseEntity:  shared.NewEntity(resourceOwner),
		Kind:        billing_entities.PlanKindTypePro,
		IsAvailable: true,
		IsActive:    true,
	}

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(currentSub, nil)
	mockPlanReader.On("GetByID", mock.Anything, currentPlanID).Return(currentPlan, nil)
	mockPlanReader.On("GetByID", mock.Anything, targetPlanID).Return(targetPlan, nil)
	mockSubWriter.On("Update", mock.Anything, mock.AnythingOfType("*billing_entities.Subscription")).Return(currentSub, nil)

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: targetPlanID,
	}

	err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	mockSubReader.AssertExpectations(t)
	mockPlanReader.AssertExpectations(t)
	mockSubWriter.AssertExpectations(t)
	assert.Equal(t, targetPlanID, currentSub.PlanID)
	assert.Len(t, currentSub.History, 1)
	assert.Contains(t, currentSub.History[0].Reason, "Upgraded from starter to pro")
}

func TestUpgradeSubscriptionUseCase_Exec_Unauthenticated(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: uuid.New(),
		PlanID: uuid.New(),
	}

	// Context has tenant but no authenticated user
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	mockSubReader.AssertNotCalled(t, "GetCurrentSubscription", mock.Anything, mock.Anything)
}

func TestUpgradeSubscriptionUseCase_Exec_NoActiveSubscription(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(nil, nil) // No active subscription

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: uuid.New(),
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active subscription found")
	mockSubReader.AssertExpectations(t)
	mockPlanReader.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestUpgradeSubscriptionUseCase_Exec_InvalidUpgradePath(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)
	resourceOwner := shared.GetResourceOwner(ctx)

	currentPlanID := uuid.New()
	targetPlanID := uuid.New()

	currentSub := &billing_entities.Subscription{
		BaseEntity:    shared.NewEntity(resourceOwner),
		PlanID:        currentPlanID,
		BillingPeriod: billing_entities.BillingPeriodMonthly,
		Status:        billing_entities.SubscriptionStatusActive,
		IsFree:        false,
		History:       []billing_entities.SubscriptionHistory{},
	}

	// Attempting to "upgrade" to a lower or same tier
	currentPlan := &billing_entities.Plan{
		BaseEntity: shared.NewEntity(resourceOwner),
		Kind:       billing_entities.PlanKindTypePro,
	}
	targetPlan := &billing_entities.Plan{
		BaseEntity:  shared.NewEntity(resourceOwner),
		Kind:        billing_entities.PlanKindTypeStarter,
		IsAvailable: true,
		IsActive:    true,
	}

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(currentSub, nil)
	mockPlanReader.On("GetByID", mock.Anything, currentPlanID).Return(currentPlan, nil)
	mockPlanReader.On("GetByID", mock.Anything, targetPlanID).Return(targetPlan, nil)

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: targetPlanID,
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not an upgrade")
	mockSubReader.AssertExpectations(t)
	mockPlanReader.AssertExpectations(t)
	mockSubWriter.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpgradeSubscriptionUseCase_Exec_TargetPlanNotAvailable(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)
	resourceOwner := shared.GetResourceOwner(ctx)

	currentPlanID := uuid.New()
	targetPlanID := uuid.New()

	currentSub := &billing_entities.Subscription{
		BaseEntity:    shared.NewEntity(resourceOwner),
		PlanID:        currentPlanID,
		BillingPeriod: billing_entities.BillingPeriodMonthly,
		Status:        billing_entities.SubscriptionStatusActive,
		IsFree:        false,
		History:       []billing_entities.SubscriptionHistory{},
	}

	currentPlan := &billing_entities.Plan{
		BaseEntity: shared.NewEntity(resourceOwner),
		Kind:       billing_entities.PlanKindTypeStarter,
	}
	targetPlan := &billing_entities.Plan{
		BaseEntity:  shared.NewEntity(resourceOwner),
		Kind:        billing_entities.PlanKindTypePro,
		IsAvailable: false, // Not available
		IsActive:    true,
	}

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(currentSub, nil)
	mockPlanReader.On("GetByID", mock.Anything, currentPlanID).Return(currentPlan, nil)
	mockPlanReader.On("GetByID", mock.Anything, targetPlanID).Return(targetPlan, nil)

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: targetPlanID,
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target plan is not available")
	mockSubReader.AssertExpectations(t)
	mockPlanReader.AssertExpectations(t)
	mockSubWriter.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpgradeSubscriptionUseCase_Exec_SubscriptionReaderError(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: uuid.New(),
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current subscription")
	mockSubReader.AssertExpectations(t)
}

func TestUpgradeSubscriptionUseCase_Exec_UpdateError(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)
	resourceOwner := shared.GetResourceOwner(ctx)

	currentPlanID := uuid.New()
	targetPlanID := uuid.New()

	currentSub := &billing_entities.Subscription{
		BaseEntity:    shared.NewEntity(resourceOwner),
		PlanID:        currentPlanID,
		BillingPeriod: billing_entities.BillingPeriodMonthly,
		Status:        billing_entities.SubscriptionStatusActive,
		IsFree:        false,
		History:       []billing_entities.SubscriptionHistory{},
	}

	currentPlan := &billing_entities.Plan{
		BaseEntity: shared.NewEntity(resourceOwner),
		Kind:       billing_entities.PlanKindTypeStarter,
	}
	targetPlan := &billing_entities.Plan{
		BaseEntity:  shared.NewEntity(resourceOwner),
		Kind:        billing_entities.PlanKindTypePro,
		IsAvailable: true,
		IsActive:    true,
	}

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(currentSub, nil)
	mockPlanReader.On("GetByID", mock.Anything, currentPlanID).Return(currentPlan, nil)
	mockPlanReader.On("GetByID", mock.Anything, targetPlanID).Return(targetPlan, nil)
	mockSubWriter.On("Update", mock.Anything, mock.AnythingOfType("*billing_entities.Subscription")).Return(nil, errors.New("update failed"))

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: targetPlanID,
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update subscription")
	mockSubWriter.AssertExpectations(t)
}

// Verify that the upgrade function correctly orders plans
func TestIsUpgrade_PlanOrdering(t *testing.T) {
	testCases := []struct {
		current  billing_entities.PlanKindType
		target   billing_entities.PlanKindType
		expected bool
	}{
		{billing_entities.PlanKindTypeFree, billing_entities.PlanKindTypeStarter, true},
		{billing_entities.PlanKindTypeStarter, billing_entities.PlanKindTypePro, true},
		{billing_entities.PlanKindTypePro, billing_entities.PlanKindTypeTeam, true},
		{billing_entities.PlanKindTypeTeam, billing_entities.PlanKindTypeBusiness, true},
		{billing_entities.PlanKindTypeBusiness, billing_entities.PlanKindTypeCustom, true},
		// Not upgrades (same level or downgrade)
		{billing_entities.PlanKindTypePro, billing_entities.PlanKindTypePro, false},
		{billing_entities.PlanKindTypePro, billing_entities.PlanKindTypeStarter, false},
		{billing_entities.PlanKindTypeCustom, billing_entities.PlanKindTypeFree, false},
	}

	// Use the use case to test ordering indirectly through error messages
	for _, tc := range testCases {
		mockSubReader := new(MockSubscriptionReader)
		mockSubWriter := new(MockSubscriptionWriter)
		mockPlanReader := new(MockPlanReader)

		usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

		userID := uuid.New()
		ctx := createAuthenticatedContext(userID)
		resourceOwner := shared.GetResourceOwner(ctx)

		currentPlanID := uuid.New()
		targetPlanID := uuid.New()

		currentSub := &billing_entities.Subscription{
			BaseEntity:    shared.NewEntity(resourceOwner),
			PlanID:        currentPlanID,
			BillingPeriod: billing_entities.BillingPeriodMonthly,
			Status:        billing_entities.SubscriptionStatusActive,
			History:       []billing_entities.SubscriptionHistory{},
		}

		currentPlan := &billing_entities.Plan{
			BaseEntity: shared.NewEntity(resourceOwner),
			Kind:       tc.current,
		}
		targetPlan := &billing_entities.Plan{
			BaseEntity:  shared.NewEntity(resourceOwner),
			Kind:        tc.target,
			IsAvailable: true,
			IsActive:    true,
		}

		mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(currentSub, nil)
		mockPlanReader.On("GetByID", mock.Anything, currentPlanID).Return(currentPlan, nil)
		mockPlanReader.On("GetByID", mock.Anything, targetPlanID).Return(targetPlan, nil)
		if tc.expected {
			mockSubWriter.On("Update", mock.Anything, mock.AnythingOfType("*billing_entities.Subscription")).Return(currentSub, nil)
		}

		cmd := billing_in.UpgradeSubscriptionCommand{
			UserID: userID,
			PlanID: targetPlanID,
		}

		err := usecase.Exec(ctx, cmd)

		if tc.expected {
			assert.NoError(t, err, "Expected %s -> %s to be a valid upgrade", tc.current, tc.target)
		} else {
			assert.Error(t, err, "Expected %s -> %s to be rejected as not an upgrade", tc.current, tc.target)
			assert.Contains(t, err.Error(), "is not an upgrade")
		}
	}
}

// TestUpgradeSubscriptionUseCase_Exec_FreeToProUpgrade tests upgrading from free to pro tier
func TestUpgradeSubscriptionUseCase_Exec_FreeToProUpgrade(t *testing.T) {
	mockSubReader := new(MockSubscriptionReader)
	mockSubWriter := new(MockSubscriptionWriter)
	mockPlanReader := new(MockPlanReader)

	usecase := billing_usecases.NewUpgradeSubscriptionUseCase(mockSubReader, mockSubWriter, mockPlanReader)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)
	resourceOwner := shared.GetResourceOwner(ctx)

	currentPlanID := uuid.New()
	targetPlanID := uuid.New()

	currentSub := &billing_entities.Subscription{
		BaseEntity:    shared.NewEntity(resourceOwner),
		PlanID:        currentPlanID,
		BillingPeriod: billing_entities.BillingPeriodMonthly,
		Status:        billing_entities.SubscriptionStatusActive,
		IsFree:        true,
		History:       []billing_entities.SubscriptionHistory{},
	}

	currentPlan := &billing_entities.Plan{
		BaseEntity: shared.NewEntity(resourceOwner),
		Kind:       billing_entities.PlanKindTypeFree,
		IsFree:     true,
	}
	targetPlan := &billing_entities.Plan{
		BaseEntity:  shared.NewEntity(resourceOwner),
		Kind:        billing_entities.PlanKindTypePro,
		IsFree:      false,
		IsAvailable: true,
		IsActive:    true,
	}

	mockSubReader.On("GetCurrentSubscription", mock.Anything, mock.Anything).Return(currentSub, nil)
	mockPlanReader.On("GetByID", mock.Anything, currentPlanID).Return(currentPlan, nil)
	mockPlanReader.On("GetByID", mock.Anything, targetPlanID).Return(targetPlan, nil)
	mockSubWriter.On("Update", mock.Anything, mock.AnythingOfType("*billing_entities.Subscription")).Return(currentSub, nil)

	cmd := billing_in.UpgradeSubscriptionCommand{
		UserID: userID,
		PlanID: targetPlanID,
	}

	err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	assert.False(t, currentSub.IsFree, "Subscription should no longer be free after upgrade")
	assert.Equal(t, targetPlanID, currentSub.PlanID)
}

// Compile test to ensure the module is valid
var _ = time.Now // suppress unused import warning
