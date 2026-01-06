package squad_usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_usecases "github.com/replay-api/replay-api/pkg/domain/squad/usecases"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations

type MockSquadWriter struct {
	mock.Mock
}

func (m *MockSquadWriter) Create(ctx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error) {
	args := m.Called(ctx, squad)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*squad_entities.Squad), args.Error(1)
}

func (m *MockSquadWriter) CreateMany(ctx context.Context, squads []*squad_entities.Squad) error {
	args := m.Called(ctx, squads)
	return args.Error(0)
}

func (m *MockSquadWriter) Update(ctx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error) {
	args := m.Called(ctx, squad)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*squad_entities.Squad), args.Error(1)
}

func (m *MockSquadWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSquadReader struct {
	mock.Mock
}

func (m *MockSquadReader) Search(ctx context.Context, search shared.Search) ([]squad_entities.Squad, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]squad_entities.Squad), args.Error(1)
}

func (m *MockSquadReader) GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.Squad, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*squad_entities.Squad), args.Error(1)
}

func (m *MockSquadReader) Compile(ctx context.Context, aggregations []shared.SearchAggregation, options shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, aggregations, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

type MockGroupWriter struct {
	mock.Mock
}

func (m *MockGroupWriter) Create(ctx context.Context, group *iam_entities.Group) (*iam_entities.Group, error) {
	args := m.Called(ctx, group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.Group), args.Error(1)
}

func (m *MockGroupWriter) CreateMany(ctx context.Context, groups []*iam_entities.Group) error {
	args := m.Called(ctx, groups)
	return args.Error(0)
}

func (m *MockGroupWriter) Update(ctx context.Context, group *iam_entities.Group) (*iam_entities.Group, error) {
	args := m.Called(ctx, group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.Group), args.Error(1)
}

func (m *MockGroupWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockGroupReader struct {
	mock.Mock
}

func (m *MockGroupReader) Search(ctx context.Context, search shared.Search) ([]iam_entities.Group, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]iam_entities.Group), args.Error(1)
}

func (m *MockGroupReader) GetByID(ctx context.Context, id uuid.UUID) (*iam_entities.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.Group), args.Error(1)
}

func (m *MockGroupReader) Compile(ctx context.Context, aggregations []shared.SearchAggregation, options shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, aggregations, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

type MockMediaWriter struct {
	mock.Mock
}

func (m *MockMediaWriter) Create(ctx context.Context, data []byte, name string, extension string) (string, error) {
	args := m.Called(ctx, data, name, extension)
	return args.String(0), args.Error(1)
}

type MockBillableOperationHandler struct {
	mock.Mock
}

func (m *MockBillableOperationHandler) Validate(ctx context.Context, cmd billing_in.BillableOperationCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockBillableOperationHandler) Exec(ctx context.Context, cmd billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	args := m.Called(ctx, cmd)
	var billable *billing_entities.BillableEntry
	var subscription *billing_entities.Subscription
	if args.Get(0) != nil {
		billable = args.Get(0).(*billing_entities.BillableEntry)
	}
	if args.Get(1) != nil {
		subscription = args.Get(1).(*billing_entities.Subscription)
	}
	return billable, subscription, args.Error(2)
}

// Helper to create authenticated context
func createAuthenticatedContext(userID, tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, tenantID)
	return ctx
}

// Tests

func TestValidateSlugURL(t *testing.T) {
	tests := []struct {
		name      string
		slugURL   string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid slug with lowercase letters",
			slugURL:   "my-squad",
			expectErr: false,
		},
		{
			name:      "Valid slug with numbers",
			slugURL:   "squad123",
			expectErr: false,
		},
		{
			name:      "Valid slug with underscore",
			slugURL:   "my_squad",
			expectErr: false,
		},
		{
			name:      "Valid slug with slash",
			slugURL:   "team/squad",
			expectErr: false,
		},
		{
			name:      "Too short",
			slugURL:   "ab",
			expectErr: true,
			errMsg:    "slugURI must be at least 3 characters long",
		},
		{
			name:      "Invalid characters - uppercase",
			slugURL:   "MySquad",
			expectErr: true,
			errMsg:    "slugURI contains invalid characters",
		},
		{
			name:      "Invalid characters - space",
			slugURL:   "my squad",
			expectErr: true,
			errMsg:    "slugURI contains invalid characters",
		},
		{
			name:      "Invalid characters - special",
			slugURL:   "my@squad",
			expectErr: true,
			errMsg:    "slugURI contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := squad_usecases.ValidateSlugURL(tt.slugURL)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMembershipUUIDs(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name      string
		members   map[string]squad_in.CreateSquadMembershipInput
		expectErr bool
	}{
		{
			name: "Valid UUIDs",
			members: map[string]squad_in.CreateSquadMembershipInput{
				validUUID: {
					Type:   squad_value_objects.SquadMembershipTypeMember,
					Roles:  []string{"player"},
					Status: squad_value_objects.SquadMembershipStatusActive,
				},
			},
			expectErr: false,
		},
		{
			name:      "Empty members",
			members:   map[string]squad_in.CreateSquadMembershipInput{},
			expectErr: false,
		},
		{
			name: "Invalid UUID",
			members: map[string]squad_in.CreateSquadMembershipInput{
				"not-a-uuid": {
					Type:   squad_value_objects.SquadMembershipTypeMember,
					Roles:  []string{"player"},
					Status: squad_value_objects.SquadMembershipStatusActive,
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := squad_usecases.ValidateMembershipUUIDs(tt.members)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateSquadUseCase_Exec_Unauthorized(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	// Unauthenticated context
	ctx := context.Background()

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "test-squad",
		Members:     map[string]squad_in.CreateSquadMembershipInput{},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.Nil(t, squad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestCreateSquadUseCase_Exec_InvalidSlugURI(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	userID := uuid.New()
	tenantID := uuid.New()
	ctx := createAuthenticatedContext(userID, tenantID)

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "ab", // Too short
		Members:     map[string]squad_in.CreateSquadMembershipInput{},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.Nil(t, squad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 3 characters")
}

func TestCreateSquadUseCase_Exec_DuplicateSlugURI(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	userID := uuid.New()
	tenantID := uuid.New()
	ctx := createAuthenticatedContext(userID, tenantID)

	// Mock existing squad with same slug
	existingSquad := squad_entities.Squad{
		BaseEntity: shared.BaseEntity{
			ID: uuid.New(),
		},
		SlugURI: "test-squad",
	}

	mockSquadReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.Squad{existingSquad}, nil).Once()

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "test-squad",
		Members:     map[string]squad_in.CreateSquadMembershipInput{},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.Nil(t, squad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCreateSquadUseCase_Exec_Success(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	userID := uuid.New()
	tenantID := uuid.New()
	groupID := uuid.New()
	ctx := createAuthenticatedContext(userID, tenantID)

	// No existing squads
	mockSquadReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.Squad{}, nil)

	// No existing group, will create one
	mockGroupReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Group{}, nil)

	newGroup := &iam_entities.Group{
		ID: groupID,
	}
	mockGroupWriter.On("Create", mock.Anything, mock.Anything).Return(newGroup, nil)

	// Billing operations
	mockBillableHandler.On("Validate", mock.Anything, mock.Anything).Return(nil)
	mockBillableHandler.On("Exec", mock.Anything, mock.Anything).Return(&billing_entities.BillableEntry{}, &billing_entities.Subscription{}, nil)

	// Squad creation
	createdSquad := &squad_entities.Squad{
		BaseEntity: shared.BaseEntity{
			ID: uuid.New(),
		},
		Name:    "Test Squad",
		SlugURI: "test-squad",
	}
	mockSquadWriter.On("Create", mock.Anything, mock.Anything).Return(createdSquad, nil)

	// History creation
	mockSquadHistoryWriter.On("Create", mock.Anything, mock.Anything).Return(&squad_entities.SquadHistory{}, nil)

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "test-squad",
		Members:     map[string]squad_in.CreateSquadMembershipInput{},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, squad)
	assert.Equal(t, "Test Squad", squad.Name)
	mockSquadWriter.AssertExpectations(t)
	mockGroupWriter.AssertExpectations(t)
	mockBillableHandler.AssertExpectations(t)
}

func TestCreateSquadUseCase_Exec_BillingFailure(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	userID := uuid.New()
	tenantID := uuid.New()
	groupID := uuid.New()
	ctx := createAuthenticatedContext(userID, tenantID)

	// No existing squads
	mockSquadReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.Squad{}, nil)

	// Existing group
	existingGroup := iam_entities.Group{
		ID: groupID,
	}
	mockGroupReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Group{existingGroup}, nil)

	// Billing validation passes but execution fails
	mockBillableHandler.On("Validate", mock.Anything, mock.Anything).Return(nil)
	mockBillableHandler.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, errors.New("billing failed"))

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "test-squad",
		Members:     map[string]squad_in.CreateSquadMembershipInput{},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.Nil(t, squad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "billing failed")
}

func TestCreateSquadUseCase_Exec_WithMembers(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	userID := uuid.New()
	tenantID := uuid.New()
	groupID := uuid.New()
	playerProfileID := uuid.New()
	ctx := createAuthenticatedContext(userID, tenantID)

	// No existing squads
	mockSquadReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.Squad{}, nil)

	// Existing group
	existingGroup := iam_entities.Group{
		ID: groupID,
	}
	mockGroupReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Group{existingGroup}, nil)

	// Player profile for membership
	playerProfile := squad_entities.PlayerProfile{
		BaseEntity: shared.BaseEntity{
			ID: playerProfileID,
			ResourceOwner: shared.ResourceOwner{
				UserID:   userID,
				TenantID: tenantID,
			},
		},
	}
	mockPlayerProfileReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.PlayerProfile{playerProfile}, nil)

	// Billing operations
	mockBillableHandler.On("Validate", mock.Anything, mock.Anything).Return(nil)
	mockBillableHandler.On("Exec", mock.Anything, mock.Anything).Return(&billing_entities.BillableEntry{}, &billing_entities.Subscription{}, nil)

	// Squad creation
	createdSquad := &squad_entities.Squad{
		BaseEntity: shared.BaseEntity{
			ID: uuid.New(),
		},
		Name:    "Test Squad",
		SlugURI: "test-squad",
	}
	mockSquadWriter.On("Create", mock.Anything, mock.Anything).Return(createdSquad, nil)

	// History creation
	mockSquadHistoryWriter.On("Create", mock.Anything, mock.Anything).Return(&squad_entities.SquadHistory{}, nil)

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "test-squad",
		Members: map[string]squad_in.CreateSquadMembershipInput{
			playerProfileID.String(): {
				Type:   squad_value_objects.SquadMembershipTypeMember,
				Roles:  []string{"player"},
				Status: squad_value_objects.SquadMembershipStatusActive,
			},
		},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, squad)
	mockPlayerProfileReader.AssertExpectations(t)
}

func TestCreateSquadUseCase_Exec_MemberNotFound(t *testing.T) {
	mockSquadWriter := new(MockSquadWriter)
	mockSquadReader := new(MockSquadReader)
	mockGroupWriter := new(MockGroupWriter)
	mockGroupReader := new(MockGroupReader)
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockMediaWriter := new(MockMediaWriter)
	mockBillableHandler := new(MockBillableOperationHandler)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	uc := squad_usecases.NewCreateSquadUseCase(
		mockSquadWriter,
		mockSquadHistoryWriter,
		mockSquadReader,
		mockGroupWriter,
		mockGroupReader,
		mockPlayerProfileReader,
		mockMediaWriter,
		mockBillableHandler,
	)

	userID := uuid.New()
	tenantID := uuid.New()
	groupID := uuid.New()
	playerProfileID := uuid.New()
	ctx := createAuthenticatedContext(userID, tenantID)

	// No existing squads
	mockSquadReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.Squad{}, nil)

	// Existing group
	existingGroup := iam_entities.Group{
		ID: groupID,
	}
	mockGroupReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Group{existingGroup}, nil)

	// Player profile NOT found
	mockPlayerProfileReader.On("Search", mock.Anything, mock.Anything).Return([]squad_entities.PlayerProfile{}, nil)

	// Billing validation
	mockBillableHandler.On("Validate", mock.Anything, mock.Anything).Return(nil)

	cmd := squad_in.CreateOrUpdatedSquadCommand{
		GameID:      "cs2",
		Name:        "Test Squad",
		Symbol:      "TS",
		Description: "A test squad",
		SlugURI:     "test-squad",
		Members: map[string]squad_in.CreateSquadMembershipInput{
			playerProfileID.String(): {
				Type:   squad_value_objects.SquadMembershipTypeMember,
				Roles:  []string{"player"},
				Status: squad_value_objects.SquadMembershipStatusActive,
			},
		},
	}

	squad, err := uc.Exec(ctx, cmd)

	assert.Nil(t, squad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

