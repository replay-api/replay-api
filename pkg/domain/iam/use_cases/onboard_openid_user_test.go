package iam_use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_use_cases "github.com/replay-api/replay-api/pkg/domain/iam/use_cases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Mock Implementations - Complete Interface Coverage
// =============================================================================

// MockUserReader implements iam_out.UserReader
type MockUserReader struct {
	mock.Mock
}

func (m *MockUserReader) Search(ctx context.Context, s shared.Search) ([]iam_entities.User, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]iam_entities.User), args.Error(1)
}

// MockUserWriter implements iam_out.UserWriter
type MockUserWriter struct {
	mock.Mock
}

func (m *MockUserWriter) Create(ctx context.Context, user *iam_entities.User) (*iam_entities.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.User), args.Error(1)
}

func (m *MockUserWriter) CreateMany(ctx context.Context, users []*iam_entities.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

// MockProfileReader implements iam_out.ProfileReader (shared.Searchable[Profile])
type MockProfileReader struct {
	mock.Mock
}

func (m *MockProfileReader) Search(ctx context.Context, s shared.Search) ([]iam_entities.Profile, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]iam_entities.Profile), args.Error(1)
}

func (m *MockProfileReader) GetByID(ctx context.Context, id uuid.UUID) (*iam_entities.Profile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.Profile), args.Error(1)
}

func (m *MockProfileReader) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

// MockProfileWriter implements iam_out.ProfileWriter
type MockProfileWriter struct {
	mock.Mock
}

func (m *MockProfileWriter) Create(ctx context.Context, profile *iam_entities.Profile) (*iam_entities.Profile, error) {
	args := m.Called(ctx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.Profile), args.Error(1)
}

func (m *MockProfileWriter) CreateMany(ctx context.Context, profiles []*iam_entities.Profile) error {
	args := m.Called(ctx, profiles)
	return args.Error(0)
}

// MockGroupWriter implements iam_out.GroupWriter
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

// MockMembershipWriter implements iam_out.MembershipWriter
type MockMembershipWriter struct {
	mock.Mock
}

func (m *MockMembershipWriter) Create(ctx context.Context, membership *iam_entities.Membership) (*iam_entities.Membership, error) {
	args := m.Called(ctx, membership)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.Membership), args.Error(1)
}

func (m *MockMembershipWriter) CreateMany(ctx context.Context, memberships []*iam_entities.Membership) error {
	args := m.Called(ctx, memberships)
	return args.Error(0)
}

// MockCreateRIDTokenCommand implements iam_in.CreateRIDTokenCommand
type MockCreateRIDTokenCommand struct {
	mock.Mock
}

func (m *MockCreateRIDTokenCommand) Exec(ctx context.Context, owner shared.ResourceOwner, source iam_entities.RIDSourceKey, audience shared.IntendedAudienceKey) (*iam_entities.RIDToken, error) {
	args := m.Called(ctx, owner, source, audience)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.RIDToken), args.Error(1)
}

// =============================================================================
// Test Helpers
// =============================================================================

func createContextWithResourceOwner(userID, groupID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, groupID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	return ctx
}

func createTestProfile(userID, groupID uuid.UUID, source iam_entities.RIDSourceKey, key string) iam_entities.Profile {
	rxn := shared.ResourceOwner{UserID: userID, GroupID: groupID}
	return iam_entities.Profile{
		ID:            uuid.New(),
		RIDSource:     source,
		SourceKey:     key,
		ResourceOwner: rxn,
	}
}

func createTestUser(id uuid.UUID, name string, rxn shared.ResourceOwner) *iam_entities.User {
	return iam_entities.NewUser(id, name, rxn)
}

func createTestGroup(id uuid.UUID, name string, rxn shared.ResourceOwner) *iam_entities.Group {
	return iam_entities.NewGroup(id, name, iam_entities.GroupTypeSystem, rxn)
}

func createTestMembership(rxn shared.ResourceOwner) *iam_entities.Membership {
	return iam_entities.NewMembership(iam_entities.MembershipTypeOwner, iam_entities.MembershipStatusActive, rxn)
}

func createTestRIDToken(rxn shared.ResourceOwner, source iam_entities.RIDSourceKey) *iam_entities.RIDToken {
	return &iam_entities.RIDToken{
		ID:               uuid.New(),
		Key:              uuid.New(),
		Source:           source,
		ResourceOwner:    rxn,
		IntendedAudience: iam_entities.DefaultTokenAudience,
		GrantType:        "authorization_code",
	}
}

// =============================================================================
// Business Scenario Tests
// =============================================================================

// TestScenario_NewUserOnboardingViaSteam tests the full onboarding flow for a new Steam user
func TestScenario_NewUserOnboardingViaSteam(t *testing.T) {
	// Given: A new user authenticating via Steam for the first time
	userID := uuid.New()
	groupID := uuid.New()
	steamID := "76561198012345678"
	ctx := createContextWithResourceOwner(userID, groupID)
	rxn := shared.ResourceOwner{UserID: userID, GroupID: groupID}

	mockUserReader := new(MockUserReader)
	mockUserWriter := new(MockUserWriter)
	mockProfileReader := new(MockProfileReader)
	mockProfileWriter := new(MockProfileWriter)
	mockGroupWriter := new(MockGroupWriter)
	mockMembershipWriter := new(MockMembershipWriter)
	mockCreateRIDToken := new(MockCreateRIDTokenCommand)

	// Profile search returns empty (new user)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)

	// User creation succeeds
	expectedUser := createTestUser(userID, "SteamGamer", rxn)
	mockUserWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.User")).Return(expectedUser, nil)

	// Group creation succeeds
	expectedGroup := createTestGroup(groupID, iam_entities.DefaultUserGroupName, rxn)
	mockGroupWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Group")).Return(expectedGroup, nil)

	// Membership creation succeeds
	expectedMembership := createTestMembership(rxn)
	mockMembershipWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Membership")).Return(expectedMembership, nil)

	// Profile creation succeeds
	expectedProfile := &iam_entities.Profile{
		ID:            uuid.New(),
		RIDSource:     iam_entities.RIDSource_Steam,
		SourceKey:     steamID,
		ResourceOwner: rxn,
	}
	mockProfileWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Profile")).Return(expectedProfile, nil)

	// RID Token creation succeeds
	expectedToken := createTestRIDToken(rxn, iam_entities.RIDSource_Steam)
	mockCreateRIDToken.On("Exec", mock.Anything, mock.Anything, iam_entities.RIDSource_Steam, iam_entities.DefaultTokenAudience).Return(expectedToken, nil)

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		mockUserReader,
		mockUserWriter,
		mockProfileReader,
		mockProfileWriter,
		mockGroupWriter,
		mockMembershipWriter,
		mockCreateRIDToken,
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:           "SteamGamer",
		Source:         iam_entities.RIDSource_Steam,
		Key:            steamID,
		ProfileDetails: map[string]interface{}{"avatar": "https://steamcdn.com/avatar.jpg"},
	}

	// When: The user completes Steam authentication
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: A complete user profile and session token are created
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.NotNil(t, token)
	assert.Equal(t, iam_entities.RIDSource_Steam, profile.RIDSource)
	assert.Equal(t, steamID, profile.SourceKey)

	// Verify all creation calls were made
	mockUserWriter.AssertExpectations(t)
	mockGroupWriter.AssertExpectations(t)
	mockMembershipWriter.AssertExpectations(t)
	mockProfileWriter.AssertExpectations(t)
	mockCreateRIDToken.AssertExpectations(t)
}

// TestScenario_ReturningUserLoginViaSteam tests that existing users get a new token without recreation
func TestScenario_ReturningUserLoginViaSteam(t *testing.T) {
	// Given: An existing user who has previously onboarded via Steam
	userID := uuid.New()
	groupID := uuid.New()
	steamID := "76561198012345678"
	ctx := createContextWithResourceOwner(userID, groupID)
	rxn := shared.ResourceOwner{UserID: userID, GroupID: groupID}

	mockUserReader := new(MockUserReader)
	mockUserWriter := new(MockUserWriter)
	mockProfileReader := new(MockProfileReader)
	mockProfileWriter := new(MockProfileWriter)
	mockGroupWriter := new(MockGroupWriter)
	mockMembershipWriter := new(MockMembershipWriter)
	mockCreateRIDToken := new(MockCreateRIDTokenCommand)

	// Profile search returns existing user
	existingProfile := createTestProfile(userID, groupID, iam_entities.RIDSource_Steam, steamID)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{existingProfile}, nil)

	// Only RID Token creation should be called (reusing profile)
	expectedToken := createTestRIDToken(rxn, iam_entities.RIDSource_Steam)
	mockCreateRIDToken.On("Exec", mock.Anything, mock.Anything, iam_entities.RIDSource_Steam, iam_entities.DefaultTokenAudience).Return(expectedToken, nil)

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		mockUserReader,
		mockUserWriter,
		mockProfileReader,
		mockProfileWriter,
		mockGroupWriter,
		mockMembershipWriter,
		mockCreateRIDToken,
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "SteamGamer",
		Source: iam_entities.RIDSource_Steam,
		Key:    steamID,
	}

	// When: The existing user logs in via Steam
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: The existing profile is returned with a fresh token
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.NotNil(t, token)
	assert.Equal(t, existingProfile.ID, profile.ID)

	// No new entities should be created
	mockUserWriter.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockGroupWriter.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockMembershipWriter.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockProfileWriter.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestScenario_GoogleOAuthOnboarding tests onboarding via Google OAuth
func TestScenario_GoogleOAuthOnboarding(t *testing.T) {
	// Given: A new user authenticating via Google OAuth
	userID := uuid.New()
	groupID := uuid.New()
	googleEmail := "gamer@gmail.com"
	ctx := createContextWithResourceOwner(userID, groupID)
	rxn := shared.ResourceOwner{UserID: userID, GroupID: groupID}

	mockUserReader := new(MockUserReader)
	mockUserWriter := new(MockUserWriter)
	mockProfileReader := new(MockProfileReader)
	mockProfileWriter := new(MockProfileWriter)
	mockGroupWriter := new(MockGroupWriter)
	mockMembershipWriter := new(MockMembershipWriter)
	mockCreateRIDToken := new(MockCreateRIDTokenCommand)

	// Setup mocks for new user flow
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)
	mockUserWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.User")).Return(createTestUser(userID, "GoogleGamer", rxn), nil)
	mockGroupWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Group")).Return(createTestGroup(groupID, iam_entities.DefaultUserGroupName, rxn), nil)
	mockMembershipWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Membership")).Return(createTestMembership(rxn), nil)

	expectedProfile := &iam_entities.Profile{
		ID:            uuid.New(),
		RIDSource:     iam_entities.RIDSource_Google,
		SourceKey:     googleEmail,
		ResourceOwner: rxn,
	}
	mockProfileWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Profile")).Return(expectedProfile, nil)
	mockCreateRIDToken.On("Exec", mock.Anything, mock.Anything, iam_entities.RIDSource_Google, iam_entities.DefaultTokenAudience).Return(createTestRIDToken(rxn, iam_entities.RIDSource_Google), nil)

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		mockUserReader,
		mockUserWriter,
		mockProfileReader,
		mockProfileWriter,
		mockGroupWriter,
		mockMembershipWriter,
		mockCreateRIDToken,
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "GoogleGamer",
		Source: iam_entities.RIDSource_Google,
		Key:    googleEmail,
		ProfileDetails: map[string]interface{}{
			"picture": "https://lh3.googleusercontent.com/avatar",
			"locale":  "en-US",
		},
	}

	// When: The user completes Google OAuth
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Profile is created with Google as source
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.NotNil(t, token)
	assert.Equal(t, iam_entities.RIDSource_Google, profile.RIDSource)
	assert.Equal(t, googleEmail, profile.SourceKey)
}

// =============================================================================
// Error Path Tests
// =============================================================================

func TestOnboardOpenIDUser_FailsWithoutUserID(t *testing.T) {
	// Given: A context without UserID
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		new(MockUserWriter),
		mockProfileReader,
		new(MockProfileWriter),
		new(MockGroupWriter),
		new(MockMembershipWriter),
		new(MockCreateRIDTokenCommand),
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: Attempting to onboard without UserID
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is returned indicating missing UserID
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no user id")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}

func TestOnboardOpenIDUser_FailsWithoutGroupID(t *testing.T) {
	// Given: A context without GroupID
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		new(MockUserWriter),
		mockProfileReader,
		new(MockProfileWriter),
		new(MockGroupWriter),
		new(MockMembershipWriter),
		new(MockCreateRIDTokenCommand),
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: Attempting to onboard without GroupID
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is returned indicating missing GroupID
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no group id")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}

func TestOnboardOpenIDUser_ProfileSearchError(t *testing.T) {
	// Given: A database error when searching for profiles
	ctx := createContextWithResourceOwner(uuid.New(), uuid.New())

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return(nil, errors.New("database connection error"))

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		new(MockUserWriter),
		mockProfileReader,
		new(MockProfileWriter),
		new(MockGroupWriter),
		new(MockMembershipWriter),
		new(MockCreateRIDTokenCommand),
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: Profile search fails
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection error")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}

func TestOnboardOpenIDUser_UserCreationError(t *testing.T) {
	// Given: User creation fails
	userID := uuid.New()
	groupID := uuid.New()
	ctx := createContextWithResourceOwner(userID, groupID)

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)

	mockUserWriter := new(MockUserWriter)
	mockUserWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.User")).Return(nil, errors.New("duplicate user error"))

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		mockUserWriter,
		mockProfileReader,
		new(MockProfileWriter),
		new(MockGroupWriter),
		new(MockMembershipWriter),
		new(MockCreateRIDTokenCommand),
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: User creation fails
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate user error")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}

func TestOnboardOpenIDUser_GroupCreationError(t *testing.T) {
	// Given: Group creation fails
	userID := uuid.New()
	groupID := uuid.New()
	ctx := createContextWithResourceOwner(userID, groupID)
	rxn := shared.ResourceOwner{UserID: userID, GroupID: groupID}

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)

	mockUserWriter := new(MockUserWriter)
	mockUserWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.User")).Return(createTestUser(userID, "TestUser", rxn), nil)

	mockGroupWriter := new(MockGroupWriter)
	mockGroupWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Group")).Return(nil, errors.New("group creation failed"))

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		mockUserWriter,
		mockProfileReader,
		new(MockProfileWriter),
		mockGroupWriter,
		new(MockMembershipWriter),
		new(MockCreateRIDTokenCommand),
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: Group creation fails
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group creation failed")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}

func TestOnboardOpenIDUser_TokenCreationError(t *testing.T) {
	// Given: Token creation fails for new user
	userID := uuid.New()
	groupID := uuid.New()
	ctx := createContextWithResourceOwner(userID, groupID)
	rxn := shared.ResourceOwner{UserID: userID, GroupID: groupID}

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{}, nil)

	mockUserWriter := new(MockUserWriter)
	mockUserWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.User")).Return(createTestUser(userID, "TestUser", rxn), nil)

	mockGroupWriter := new(MockGroupWriter)
	mockGroupWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Group")).Return(createTestGroup(groupID, iam_entities.DefaultUserGroupName, rxn), nil)

	mockMembershipWriter := new(MockMembershipWriter)
	mockMembershipWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Membership")).Return(createTestMembership(rxn), nil)

	expectedProfile := &iam_entities.Profile{
		ID:            uuid.New(),
		RIDSource:     iam_entities.RIDSource_Steam,
		SourceKey:     "12345",
		ResourceOwner: rxn,
	}
	mockProfileWriter := new(MockProfileWriter)
	mockProfileWriter.On("Create", mock.Anything, mock.AnythingOfType("*iam_entities.Profile")).Return(expectedProfile, nil)

	mockCreateRIDToken := new(MockCreateRIDTokenCommand)
	mockCreateRIDToken.On("Exec", mock.Anything, mock.Anything, iam_entities.RIDSource_Steam, iam_entities.DefaultTokenAudience).Return(nil, errors.New("token creation failed"))

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		mockUserWriter,
		mockProfileReader,
		mockProfileWriter,
		mockGroupWriter,
		mockMembershipWriter,
		mockCreateRIDToken,
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: Token creation fails
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token creation failed")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}

func TestOnboardOpenIDUser_ExistingUserTokenCreationError(t *testing.T) {
	// Given: Token creation fails for existing user
	userID := uuid.New()
	groupID := uuid.New()
	ctx := createContextWithResourceOwner(userID, groupID)

	existingProfile := createTestProfile(userID, groupID, iam_entities.RIDSource_Steam, "12345")

	mockProfileReader := new(MockProfileReader)
	mockProfileReader.On("Search", mock.Anything, mock.Anything).Return([]iam_entities.Profile{existingProfile}, nil)

	mockCreateRIDToken := new(MockCreateRIDTokenCommand)
	mockCreateRIDToken.On("Exec", mock.Anything, mock.Anything, iam_entities.RIDSource_Steam, iam_entities.DefaultTokenAudience).Return(nil, errors.New("token creation failed"))

	usecase := iam_use_cases.NewOnboardOpenIDUserUseCase(
		new(MockUserReader),
		new(MockUserWriter),
		mockProfileReader,
		new(MockProfileWriter),
		new(MockGroupWriter),
		new(MockMembershipWriter),
		mockCreateRIDToken,
	)

	cmd := iam_in.OnboardOpenIDUserCommand{
		Name:   "TestUser",
		Source: iam_entities.RIDSource_Steam,
		Key:    "12345",
	}

	// When: Token creation fails for existing user
	profile, token, err := usecase.Exec(ctx, cmd)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token creation failed")
	assert.Nil(t, profile)
	assert.Nil(t, token)
}
