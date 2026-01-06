package use_cases_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	use_cases "github.com/replay-api/replay-api/pkg/domain/replay/use_cases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Mock Implementations
// =============================================================================

// MockReplayFileMetadataWriter implements replay_out.ReplayFileMetadataWriter
type MockReplayFileMetadataWriter struct {
	mock.Mock
}

func (m *MockReplayFileMetadataWriter) Create(ctx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error) {
	args := m.Called(ctx, replayFile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*replay_entity.ReplayFile), args.Error(1)
}

func (m *MockReplayFileMetadataWriter) Update(ctx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error) {
	args := m.Called(ctx, replayFile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*replay_entity.ReplayFile), args.Error(1)
}

// MockReplayFileContentWriter implements replay_out.ReplayFileContentWriter
type MockReplayFileContentWriter struct {
	mock.Mock
}

func (m *MockReplayFileContentWriter) Put(ctx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error) {
	args := m.Called(ctx, replayFileID, reader)
	return args.String(0), args.Error(1)
}

// =============================================================================
// Test Helpers
// =============================================================================

func createAuthenticatedContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	return ctx
}

func createUnauthenticatedContext() context.Context {
	return context.Background()
}

func createTestReplayFile(size int) *replay_entity.ReplayFile {
	return &replay_entity.ReplayFile{
		ID:          uuid.New(),
		GameID:      "cs",
		NetworkID:   "steam",
		Size:        size,
		Status:      replay_entity.ReplayFileStatusPending,
		InternalURI: "",
	}
}

// =============================================================================
// Business Scenario Tests
// =============================================================================

// TestScenario_AuthenticatedUserUploadsReplay tests the happy path for replay upload
func TestScenario_AuthenticatedUserUploadsReplay(t *testing.T) {
	// Given: An authenticated user with a valid replay file
	ctx := createAuthenticatedContext()
	replayContent := []byte("demo file content simulation")
	reader := bytes.NewReader(replayContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	// Metadata creation succeeds
	createdFile := createTestReplayFile(len(replayContent))
	mockMetadataWriter.On("Create", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(createdFile, nil)

	// Content upload succeeds
	expectedURI := "s3://replays/bucket/" + createdFile.ID.String()
	mockContentWriter.On("Put", mock.Anything, createdFile.ID, mock.Anything).Return(expectedURI, nil)

	// Metadata update succeeds
	updatedFile := createdFile
	updatedFile.InternalURI = expectedURI
	updatedFile.Status = replay_entity.ReplayFileStatusProcessing
	mockMetadataWriter.On("Update", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(updatedFile, nil)

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: The user uploads the replay
	result, err := usecase.Exec(ctx, reader)

	// Then: The replay is successfully uploaded and metadata is created
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedURI, result.InternalURI)
	assert.Equal(t, replay_entity.ReplayFileStatusProcessing, result.Status)

	mockMetadataWriter.AssertExpectations(t)
	mockContentWriter.AssertExpectations(t)
}

// TestScenario_UnauthenticatedUserCannotUpload tests access control
func TestScenario_UnauthenticatedUserCannotUpload(t *testing.T) {
	// Given: An unauthenticated user trying to upload
	ctx := createUnauthenticatedContext()
	replayContent := []byte("demo file content")
	reader := bytes.NewReader(replayContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: The user attempts to upload
	result, err := usecase.Exec(ctx, reader)

	// Then: Access is denied with ErrUnauthorized
	assert.Error(t, err)
	assert.Nil(t, result)

	// No storage operations should occur
	mockMetadataWriter.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockContentWriter.AssertNotCalled(t, "Put", mock.Anything, mock.Anything, mock.Anything)
}

// TestScenario_LargeReplayFileUpload tests handling of larger files
func TestScenario_LargeReplayFileUpload(t *testing.T) {
	// Given: An authenticated user with a large replay file (10MB simulated)
	ctx := createAuthenticatedContext()
	largeContent := make([]byte, 10*1024*1024) // 10MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	reader := bytes.NewReader(largeContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	createdFile := createTestReplayFile(len(largeContent))
	mockMetadataWriter.On("Create", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(createdFile, nil)

	expectedURI := "s3://replays/bucket/" + createdFile.ID.String()
	mockContentWriter.On("Put", mock.Anything, createdFile.ID, mock.Anything).Return(expectedURI, nil)

	updatedFile := createdFile
	updatedFile.InternalURI = expectedURI
	mockMetadataWriter.On("Update", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(updatedFile, nil)

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: The user uploads the large file
	result, err := usecase.Exec(ctx, reader)

	// Then: The upload succeeds despite file size
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedURI, result.InternalURI)
}

// =============================================================================
// Error Path Tests
// =============================================================================

func TestUploadReplayFile_MetadataCreationFails(t *testing.T) {
	// Given: Metadata creation fails
	ctx := createAuthenticatedContext()
	replayContent := []byte("demo content")
	reader := bytes.NewReader(replayContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	mockMetadataWriter.On("Create", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(nil, errors.New("database error"))

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: Metadata creation fails
	result, err := usecase.Exec(ctx, reader)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, result)

	// Content should not be uploaded if metadata fails
	mockContentWriter.AssertNotCalled(t, "Put", mock.Anything, mock.Anything, mock.Anything)
}

func TestUploadReplayFile_ContentUploadFails(t *testing.T) {
	// Given: Blob storage upload fails
	ctx := createAuthenticatedContext()
	replayContent := []byte("demo content")
	reader := bytes.NewReader(replayContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	createdFile := createTestReplayFile(len(replayContent))
	mockMetadataWriter.On("Create", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(createdFile, nil)

	// Content upload fails
	mockContentWriter.On("Put", mock.Anything, createdFile.ID, mock.Anything).Return("", errors.New("storage unavailable"))

	// Metadata should be updated to failed status
	failedFile := createdFile
	failedFile.Status = replay_entity.ReplayFileStatusFailed
	mockMetadataWriter.On("Update", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(failedFile, nil)

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: Content upload fails
	result, err := usecase.Exec(ctx, reader)

	// Then: Error is propagated and metadata is marked as failed
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage unavailable")
	assert.Nil(t, result)

	// Verify metadata was updated to failed status
	mockMetadataWriter.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(rf *replay_entity.ReplayFile) bool {
		return rf.Status == replay_entity.ReplayFileStatusFailed
	}))
}

func TestUploadReplayFile_MetadataUpdateFails(t *testing.T) {
	// Given: Final metadata update fails
	ctx := createAuthenticatedContext()
	replayContent := []byte("demo content")
	reader := bytes.NewReader(replayContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	createdFile := createTestReplayFile(len(replayContent))
	mockMetadataWriter.On("Create", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(createdFile, nil)

	expectedURI := "s3://replays/bucket/" + createdFile.ID.String()
	mockContentWriter.On("Put", mock.Anything, createdFile.ID, mock.Anything).Return(expectedURI, nil)

	// Final update fails
	mockMetadataWriter.On("Update", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(nil, errors.New("update failed"))

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: Metadata update fails
	result, err := usecase.Exec(ctx, reader)

	// Then: Error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	assert.Nil(t, result)
}

func TestUploadReplayFile_AuthenticatedKeyFalse(t *testing.T) {
	// Given: A context with AuthenticatedKey explicitly set to false
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, false)

	replayContent := []byte("demo content")
	reader := bytes.NewReader(replayContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: Upload is attempted with authenticated=false
	result, err := usecase.Exec(ctx, reader)

	// Then: Access is denied
	assert.Error(t, err)
	assert.Nil(t, result)
	mockMetadataWriter.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestUploadReplayFile_EmptyFile tests handling of empty replay files
func TestUploadReplayFile_EmptyFile(t *testing.T) {
	// Given: An authenticated user with an empty file
	ctx := createAuthenticatedContext()
	emptyContent := []byte{}
	reader := bytes.NewReader(emptyContent)

	mockMetadataWriter := new(MockReplayFileMetadataWriter)
	mockContentWriter := new(MockReplayFileContentWriter)

	createdFile := createTestReplayFile(0)
	mockMetadataWriter.On("Create", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(createdFile, nil)

	expectedURI := "s3://replays/bucket/" + createdFile.ID.String()
	mockContentWriter.On("Put", mock.Anything, createdFile.ID, mock.Anything).Return(expectedURI, nil)

	updatedFile := createdFile
	updatedFile.InternalURI = expectedURI
	mockMetadataWriter.On("Update", mock.Anything, mock.AnythingOfType("*entities.ReplayFile")).Return(updatedFile, nil)

	usecase := use_cases.NewUploadReplayFileUseCase(mockMetadataWriter, mockContentWriter)

	// When: An empty file is uploaded
	result, err := usecase.Exec(ctx, reader)

	// Then: The upload succeeds (validation happens elsewhere)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

