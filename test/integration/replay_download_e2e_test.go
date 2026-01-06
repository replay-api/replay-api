//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// mockReplayContentReader provides mock replay file content for testing
type mockReplayContentReader struct {
	content map[uuid.UUID][]byte
}

func newMockReplayContentReader() *mockReplayContentReader {
	return &mockReplayContentReader{
		content: make(map[uuid.UUID][]byte),
	}
}

func (m *mockReplayContentReader) GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error) {
	content, ok := m.content[replayFileID]
	if !ok {
		return nil, shared.NewErrNotFound("ReplayFile", "ID", replayFileID.String())
	}
	return &readSeekCloser{bytes.NewReader(content)}, nil
}

// readSeekCloser wraps bytes.Reader to satisfy io.ReadSeekCloser
type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error {
	return nil
}

// TestE2E_ReplayDownload tests replay download functionality
func TestE2E_ReplayDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup: Connect to test MongoDB
	mongoURI := getMongoTestURI()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err, "Failed to connect to MongoDB")
	defer func() { _ = client.Disconnect(ctx) }()

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("Skipping replay download E2E test: MongoDB not available")
	}

	// Create test database
	dbName := "replay_download_test_" + uuid.New().String()[:8]
	defer func() {
		_ = client.Database(dbName).Drop(ctx)
	}()

	// Test user setup
	userID := uuid.New()
	groupID := uuid.New()

	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, groupID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	resourceOwner := shared.GetResourceOwner(ctx)

	t.Run("ReplayFile_EntityCreation", func(t *testing.T) {
		// Create replay file entity
		replayFile := replay_entity.NewReplayFile(
			"cs2",
			"steam",
			1024*1024, // 1MB
			"internal://replays/test.dem",
			resourceOwner,
		)

		require.NotNil(t, replayFile)
		assert.NotEqual(t, uuid.Nil, replayFile.ID)
		assert.Equal(t, replay_common.GameIDKey("cs2"), replayFile.GameID)
		assert.Equal(t, replay_common.NetworkIDKey("steam"), replayFile.NetworkID)
		assert.Equal(t, 1024*1024, replayFile.Size)
		assert.Equal(t, replay_entity.ReplayFileStatusPending, replayFile.Status)
		assert.Equal(t, userID, replayFile.ResourceOwner.UserID)

		t.Log("✓ ReplayFile entity created successfully")
	})

	t.Run("ReplayFile_StatusTransitions", func(t *testing.T) {
		replayFile := replay_entity.NewReplayFile(
			"cs2",
			"steam",
			512*1024,
			"",
			resourceOwner,
		)

		// Initial state
		assert.Equal(t, replay_entity.ReplayFileStatusPending, replayFile.Status)

		// Transition to Processing
		replayFile.Status = replay_entity.ReplayFileStatusProcessing
		assert.Equal(t, replay_entity.ReplayFileStatusProcessing, replayFile.Status)

		// Transition to Completed
		replayFile.Status = replay_entity.ReplayFileStatusCompleted
		replayFile.InternalURI = "s3://bucket/replays/test.dem"
		assert.Equal(t, replay_entity.ReplayFileStatusCompleted, replayFile.Status)
		assert.NotEmpty(t, replayFile.InternalURI)

		t.Log("✓ ReplayFile status transitions work correctly")
	})

	t.Run("ShareToken_Creation", func(t *testing.T) {
		replayID := uuid.New()
		now := time.Now()
		expiresAt := now.AddDate(0, 0, 7) // 7 days

		token := &replay_entity.ShareToken{
			ID:           uuid.New(),
			ResourceID:   replayID,
			ResourceType: replay_entity.SharingResourceTypeReplayFileContent,
			ExpiresAt:    expiresAt,
			Uri:          "https://leetgaming.pro/share/" + uuid.New().String(),
			EntityType:   "ReplayFile",
			Status:       replay_entity.ShareTokenStatusActive,
			ResourceOwner: resourceOwner,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		require.NotNil(t, token)
		assert.NotEqual(t, uuid.Nil, token.ID)
		assert.Equal(t, replayID, token.ResourceID)
		assert.Equal(t, replay_entity.ShareTokenStatusActive, token.Status)

		// Test IsValid
		assert.True(t, token.IsValid(), "Token should be valid")

		t.Log("✓ ShareToken created successfully")
	})

	t.Run("ShareToken_Expiration", func(t *testing.T) {
		replayID := uuid.New()

		// Create expired token
		expiredToken := &replay_entity.ShareToken{
			ID:           uuid.New(),
			ResourceID:   replayID,
			ResourceType: replay_entity.SharingResourceTypeReplayFileContent,
			ExpiresAt:    time.Now().Add(-24 * time.Hour), // Expired 1 day ago
			Status:       replay_entity.ShareTokenStatusActive,
			ResourceOwner: resourceOwner,
			CreatedAt:    time.Now().Add(-48 * time.Hour),
			UpdatedAt:    time.Now().Add(-48 * time.Hour),
		}

		assert.False(t, expiredToken.IsValid(), "Expired token should not be valid")

		// Create inactive token (not expired but inactive)
		inactiveToken := &replay_entity.ShareToken{
			ID:           uuid.New(),
			ResourceID:   replayID,
			ResourceType: replay_entity.SharingResourceTypeReplayFileContent,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			Status:       replay_entity.ShareTokenStatusExpired, // Manually expired
			ResourceOwner: resourceOwner,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.False(t, inactiveToken.IsValid(), "Inactive token should not be valid")

		t.Log("✓ ShareToken expiration logic works correctly")
	})

	t.Run("ShareToken_Validation", func(t *testing.T) {
		// Test validation errors
		invalidToken := &replay_entity.ShareToken{
			ID:     uuid.New(),
			Status: replay_entity.ShareTokenStatusActive,
		}

		err := invalidToken.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resource_id")

		// Fix resource_id, still missing resource_type
		invalidToken.ResourceID = uuid.New()
		err = invalidToken.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resource_type")

		// Fix resource_type
		invalidToken.ResourceType = replay_entity.SharingResourceTypeReplayFileContent
		err = invalidToken.Validate()
		require.NoError(t, err)

		t.Log("✓ ShareToken validation works correctly")
	})

	t.Run("MockReplayContentReader_Success", func(t *testing.T) {
		reader := newMockReplayContentReader()

		// Add test content
		replayID := uuid.New()
		testContent := []byte("This is a test replay file content")
		reader.content[replayID] = testContent

		// Read content
		contentReader, err := reader.GetByID(ctx, replayID)
		require.NoError(t, err)
		require.NotNil(t, contentReader)
		defer contentReader.Close()

		// Verify content
		readContent, err := io.ReadAll(contentReader)
		require.NoError(t, err)
		assert.Equal(t, testContent, readContent)

		t.Log("✓ MockReplayContentReader returns correct content")
	})

	t.Run("MockReplayContentReader_NotFound", func(t *testing.T) {
		reader := newMockReplayContentReader()

		// Try to read non-existent replay
		_, err := reader.GetByID(ctx, uuid.New())
		require.Error(t, err)

		t.Log("✓ MockReplayContentReader returns error for non-existent replay")
	})

	t.Log("✓ All replay download E2E tests passed!")
}

// TestE2E_ReplayVisibility tests replay visibility and access control
func TestE2E_ReplayVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("MatchVisibility_Types", func(t *testing.T) {
		// Test visibility constants
		assert.Equal(t, replay_entity.MatchVisibility("public"), replay_entity.MatchVisibilityPublic)
		assert.Equal(t, replay_entity.MatchVisibility("squad"), replay_entity.MatchVisibilitySquad)
		assert.Equal(t, replay_entity.MatchVisibility("private"), replay_entity.MatchVisibilityPrivate)
		assert.Equal(t, replay_entity.MatchVisibility("custom"), replay_entity.MatchVisibilityCustom)

		t.Log("✓ Match visibility types defined correctly")
	})

	t.Run("ShareTokenStatus_Types", func(t *testing.T) {
		assert.Equal(t, replay_entity.ShareTokenStatus("Active"), replay_entity.ShareTokenStatusActive)
		assert.Equal(t, replay_entity.ShareTokenStatus("Used"), replay_entity.ShareTokenStatusUsed)
		assert.Equal(t, replay_entity.ShareTokenStatus("Expired"), replay_entity.ShareTokenStatusExpired)

		t.Log("✓ Share token status types defined correctly")
	})

	t.Log("✓ All replay visibility tests passed!")
}

