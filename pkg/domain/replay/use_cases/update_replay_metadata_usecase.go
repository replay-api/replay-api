package use_cases

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

var (
	ErrReplayNotFound     = errors.New("replay file not found")
	ErrNotAuthorized      = errors.New("not authorized to update this replay")
	ErrInvalidVisibility  = errors.New("invalid visibility type")
)

// UpdateReplayMetadataUseCase handles updating user-editable metadata for replay files
type UpdateReplayMetadataUseCase struct {
	MetadataReader replay_in.ReplayFileReader
	MetadataWriter replay_out.ReplayFileMetadataWriter
}

// NewUpdateReplayMetadataUseCase creates a new use case for updating replay metadata
func NewUpdateReplayMetadataUseCase(
	metadataReader replay_in.ReplayFileReader,
	metadataWriter replay_out.ReplayFileMetadataWriter,
) *UpdateReplayMetadataUseCase {
	return &UpdateReplayMetadataUseCase{
		MetadataReader: metadataReader,
		MetadataWriter: metadataWriter,
	}
}

// Exec updates the replay file metadata. Only the owner can update their replays.
// 
// Updates can include:
//   - Title: Display name for the replay
//   - Description: User-provided description
//   - Tags: Searchable tags
//   - Visibility: Public, Private, or Restricted
//
// Authorization rules:
//   - Owner (UserID matches) can always update
//   - Group members can update if visibility allows
//   - Admins (TenantID/ClientID level) can update any replay in their scope
func (usecase *UpdateReplayMetadataUseCase) Exec(
	ctx context.Context,
	replayFileID uuid.UUID,
	updates *replay_entity.ReplayFileOptions,
) (*replay_entity.ReplayFile, error) {
	if updates == nil {
		return nil, errors.New("no updates provided")
	}

	// Get current resource owner from context
	currentOwner := shared.GetResourceOwner(ctx)
	isAdmin := shared.IsAdmin(ctx)
	
	slog.InfoContext(ctx, "updating replay metadata",
		"replayFileID", replayFileID,
		"currentOwner", currentOwner,
		"isAdmin", isAdmin,
	)

	// Fetch the existing replay file
	replayFile, err := usecase.MetadataReader.GetByID(ctx, replayFileID)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching replay file", "replayFileID", replayFileID, "err", err)
		return nil, ErrReplayNotFound
	}
	if replayFile == nil {
		return nil, ErrReplayNotFound
	}

	// Authorization check
	if !canUserUpdateReplay(ctx, currentOwner, replayFile, isAdmin) {
		slog.WarnContext(ctx, "unauthorized update attempt",
			"replayFileID", replayFileID,
			"resourceOwner", replayFile.ResourceOwner,
			"requestOwner", currentOwner,
		)
		return nil, ErrNotAuthorized
	}

	// Apply updates
	updated := false
	
	if updates.Title != "" && updates.Title != replayFile.Title {
		replayFile.Title = updates.Title
		updated = true
	}
	
	if updates.Description != "" && updates.Description != replayFile.Description {
		replayFile.Description = updates.Description
		updated = true
	}
	
	if updates.Tags != nil {
		replayFile.Tags = updates.Tags
		updated = true
	}
	
	// Update visibility if provided and valid
	if updates.Visibility != 0 {
		if err := validateVisibility(updates.Visibility); err != nil {
			return nil, err
		}
		
		// Update visibility type and level
		switch updates.Visibility {
		case shared.PublicVisibilityTypeKey:
			replayFile.VisibilityType = shared.PublicVisibilityTypeKey
			replayFile.VisibilityLevel = shared.TenantAudienceIDKey
		case shared.PrivateVisibilityTypeKey:
			replayFile.VisibilityType = shared.PrivateVisibilityTypeKey
			replayFile.VisibilityLevel = shared.UserAudienceIDKey
		case shared.RestrictedVisibilityTypeKey:
			replayFile.VisibilityType = shared.RestrictedVisibilityTypeKey
			replayFile.VisibilityLevel = shared.GroupAudienceIDKey
		}
		updated = true
	}

	if !updated {
		slog.InfoContext(ctx, "no changes detected", "replayFileID", replayFileID)
		return replayFile, nil
	}

	// Update timestamp
	replayFile.UpdatedAt = time.Now()

	// Persist changes
	result, err := usecase.MetadataWriter.Update(ctx, replayFile)
	if err != nil {
		slog.ErrorContext(ctx, "error updating replay metadata", "replayFileID", replayFileID, "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "successfully updated replay metadata",
		"replayFileID", replayFileID,
		"visibility", result.VisibilityType,
		"title", result.Title,
	)

	return result, nil
}

// canUserUpdateReplay checks if the current user is authorized to update the replay
func canUserUpdateReplay(ctx context.Context, currentOwner shared.ResourceOwner, replayFile *replay_entity.ReplayFile, isAdmin bool) bool {
	// Admins can update anything in their scope
	if isAdmin {
		return true
	}

	// Owner can always update their own replays
	if currentOwner.UserID != uuid.Nil && currentOwner.UserID == replayFile.ResourceOwner.UserID {
		return true
	}

	// Same group members can update if visibility is restricted to group
	if currentOwner.GroupID != uuid.Nil && 
	   currentOwner.GroupID == replayFile.ResourceOwner.GroupID &&
	   replayFile.VisibilityType == shared.RestrictedVisibilityTypeKey {
		return true
	}

	return false
}

// validateVisibility ensures the visibility type is valid
func validateVisibility(v shared.VisibilityTypeKey) error {
	switch v {
	case shared.PublicVisibilityTypeKey,
		shared.PrivateVisibilityTypeKey,
		shared.RestrictedVisibilityTypeKey:
		return nil
	default:
		return ErrInvalidVisibility
	}
}
