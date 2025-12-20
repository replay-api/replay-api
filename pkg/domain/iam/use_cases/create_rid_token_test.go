package iam_use_cases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_use_cases "github.com/replay-api/replay-api/pkg/domain/iam/use_cases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRIDTokenWriter is a mock implementation of iam_out.RIDTokenWriter
type MockRIDTokenWriter struct {
	mock.Mock
}

func (m *MockRIDTokenWriter) Create(ctx context.Context, token *iam_entities.RIDToken) (*iam_entities.RIDToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.RIDToken), args.Error(1)
}

func (m *MockRIDTokenWriter) Revoke(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockRIDTokenWriter) Update(ctx context.Context, token *iam_entities.RIDToken) (*iam_entities.RIDToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.RIDToken), args.Error(1)
}

func (m *MockRIDTokenWriter) Delete(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

// MockRIDTokenReader is a mock implementation of iam_out.RIDTokenReader
type MockRIDTokenReader struct {
	mock.Mock
}

func (m *MockRIDTokenReader) Search(ctx context.Context, search common.Search) ([]iam_entities.RIDToken, error) {
	args := m.Called(ctx, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]iam_entities.RIDToken), args.Error(1)
}

func (m *MockRIDTokenReader) FindByID(ctx context.Context, id uuid.UUID) (*iam_entities.RIDToken, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.RIDToken), args.Error(1)
}

func (m *MockRIDTokenReader) FindByKey(ctx context.Context, key uuid.UUID) (*iam_entities.RIDToken, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam_entities.RIDToken), args.Error(1)
}

func TestCreateRIDToken_UserAudience_Success(t *testing.T) {
	mockWriter := new(MockRIDTokenWriter)
	mockReader := new(MockRIDTokenReader)

	usecase := iam_use_cases.NewCreateRIDTokenUseCase(mockWriter, mockReader)

	ctx := context.Background()
	userID := uuid.New()
	groupID := uuid.New()
	tenantID := uuid.New()

	resourceOwner := common.ResourceOwner{
		UserID:   userID,
		GroupID:  groupID,
		TenantID: tenantID,
	}

	source := iam_entities.RIDSource_Steam

	// Mock successful token creation
	mockWriter.On("Create", mock.Anything, mock.MatchedBy(func(token *iam_entities.RIDToken) bool {
		return token.ResourceOwner.UserID == userID &&
			token.Source == source &&
			token.IntendedAudience == common.UserAudienceIDKey &&
			token.GrantType == "authorization_code" &&
			token.ExpiresAt.After(time.Now())
	})).Return(&iam_entities.RIDToken{
		ID:               uuid.New(),
		Key:              uuid.New(),
		Source:           source,
		ResourceOwner:    resourceOwner,
		IntendedAudience: common.UserAudienceIDKey,
		GrantType:        "authorization_code",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
		CreatedAt:        time.Now(),
	}, nil)

	token, err := usecase.Exec(ctx, resourceOwner, source, common.UserAudienceIDKey)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "authorization_code", token.GrantType)
	assert.Equal(t, common.UserAudienceIDKey, token.IntendedAudience)
	assert.Equal(t, userID, token.ResourceOwner.UserID)
	mockWriter.AssertExpectations(t)
}

func TestCreateRIDToken_ClientAudience_Success(t *testing.T) {
	mockWriter := new(MockRIDTokenWriter)
	mockReader := new(MockRIDTokenReader)

	usecase := iam_use_cases.NewCreateRIDTokenUseCase(mockWriter, mockReader)

	ctx := context.Background()
	userID := uuid.New()
	groupID := uuid.New()

	resourceOwner := common.ResourceOwner{
		UserID:  userID,
		GroupID: groupID,
	}

	source := iam_entities.RIDSource_Google

	// Mock successful token creation for client application
	mockWriter.On("Create", mock.Anything, mock.MatchedBy(func(token *iam_entities.RIDToken) bool {
		return token.GrantType == "client_credentials" &&
			token.IntendedAudience == common.ClientApplicationAudienceIDKey
	})).Return(&iam_entities.RIDToken{
		ID:               uuid.New(),
		Key:              uuid.New(),
		Source:           source,
		ResourceOwner:    resourceOwner,
		IntendedAudience: common.ClientApplicationAudienceIDKey,
		GrantType:        "client_credentials",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
		CreatedAt:        time.Now(),
	}, nil)

	token, err := usecase.Exec(ctx, resourceOwner, source, common.ClientApplicationAudienceIDKey)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "client_credentials", token.GrantType)
	assert.Equal(t, common.ClientApplicationAudienceIDKey, token.IntendedAudience)
	mockWriter.AssertExpectations(t)
}

func TestCreateRIDToken_WriterError(t *testing.T) {
	mockWriter := new(MockRIDTokenWriter)
	mockReader := new(MockRIDTokenReader)

	usecase := iam_use_cases.NewCreateRIDTokenUseCase(mockWriter, mockReader)

	ctx := context.Background()
	resourceOwner := common.ResourceOwner{
		UserID:  uuid.New(),
		GroupID: uuid.New(),
	}

	// Mock database error
	mockWriter.On("Create", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	token, err := usecase.Exec(ctx, resourceOwner, iam_entities.RIDSource_Email, common.UserAudienceIDKey)

	assert.Error(t, err)
	assert.Nil(t, token)
	mockWriter.AssertExpectations(t)
}

func TestCreateRIDToken_ExpirationIsOneHour(t *testing.T) {
	mockWriter := new(MockRIDTokenWriter)
	mockReader := new(MockRIDTokenReader)

	usecase := iam_use_cases.NewCreateRIDTokenUseCase(mockWriter, mockReader)

	ctx := context.Background()
	resourceOwner := common.ResourceOwner{
		UserID:  uuid.New(),
		GroupID: uuid.New(),
	}

	var capturedToken *iam_entities.RIDToken
	mockWriter.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedToken = args.Get(1).(*iam_entities.RIDToken)
	}).Return(&iam_entities.RIDToken{
		ID:        uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	usecase.Exec(ctx, resourceOwner, iam_entities.RIDSource_Steam, common.UserAudienceIDKey)

	// Verify expiration is approximately 1 hour from now
	expectedExpiry := time.Now().Add(1 * time.Hour)
	assert.NotNil(t, capturedToken)
	assert.WithinDuration(t, expectedExpiry, capturedToken.ExpiresAt, 5*time.Second)
	mockWriter.AssertExpectations(t)
}

func TestCreateRIDToken_UniqueIDsGenerated(t *testing.T) {
	mockWriter := new(MockRIDTokenWriter)
	mockReader := new(MockRIDTokenReader)

	usecase := iam_use_cases.NewCreateRIDTokenUseCase(mockWriter, mockReader)

	ctx := context.Background()
	resourceOwner := common.ResourceOwner{
		UserID:  uuid.New(),
		GroupID: uuid.New(),
	}

	var tokenIDs []uuid.UUID
	var tokenKeys []uuid.UUID

	mockWriter.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		token := args.Get(1).(*iam_entities.RIDToken)
		tokenIDs = append(tokenIDs, token.ID)
		tokenKeys = append(tokenKeys, token.Key)
	}).Return(&iam_entities.RIDToken{
		ID:  uuid.New(),
		Key: uuid.New(),
	}, nil)

	// Create multiple tokens
	for i := 0; i < 3; i++ {
		usecase.Exec(ctx, resourceOwner, iam_entities.RIDSource_Steam, common.UserAudienceIDKey)
	}

	// Verify all IDs are unique
	assert.Len(t, tokenIDs, 3)
	assert.NotEqual(t, tokenIDs[0], tokenIDs[1])
	assert.NotEqual(t, tokenIDs[1], tokenIDs[2])
	assert.NotEqual(t, tokenIDs[0], tokenIDs[2])

	// Verify all Keys are unique
	assert.Len(t, tokenKeys, 3)
	assert.NotEqual(t, tokenKeys[0], tokenKeys[1])
	assert.NotEqual(t, tokenKeys[1], tokenKeys[2])
	assert.NotEqual(t, tokenKeys[0], tokenKeys[2])
}

