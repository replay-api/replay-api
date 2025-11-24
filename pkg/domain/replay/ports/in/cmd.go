package replay_in

import (
	"context"
	"io"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

// UploadReplayFileCommand is an interface that defines the contract for executing a command to upload a replay file.
type UploadReplayFileCommand interface {
	// Exec executes the UploadReplayFileCommand with the given user context and file.
	// It returns the UUID of the uploaded replay file and any error encountered.
	Exec(c context.Context, file io.Reader) (*replay_entity.ReplayFile, error)
}

type UploadAndProcessReplayFileCommand interface {
	// Exec executes the UploadAndProcessReplayFileCommand with the given user context and file.
	// It returns the processed MatchID and any error encountered.
	Exec(c context.Context, file io.Reader) (*replay_entity.Match, error)
}

// ProcessReplayFileCommand is an interface that defines the contract for executing a command to process a replay file.
type ProcessReplayFileCommand interface {
	// Exec executes the command to process a replay file.
	// It takes a user context and a replayFileID as input parameters.
	// It returns the processed MatchID and an error if any.
	Exec(c context.Context, replayFileID uuid.UUID) (*replay_entity.Match, error)
}

// UpdateReplayFileHeaderCommand is an interface that defines the contract for updating a replay file header based on processed game events
type UpdateReplayFileHeaderCommand interface {
	Exec(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error)
}

// ShareTokenCommand is an interface for share token management operations
type ShareTokenCommand interface {
	// CreateToken creates a new share token for a resource
	Create(ctx context.Context, token *replay_entity.ShareToken) error
	// RevokeToken revokes (deletes) a share token
	Revoke(ctx context.Context, tokenID uuid.UUID) error
	// UpdateToken updates share token properties
	Update(ctx context.Context, token *replay_entity.ShareToken) error
}
