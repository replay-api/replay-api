package entities

import (
	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type ReplayFileStatus string

const (
	ReplayFileStatusPending     ReplayFileStatus = "Pending"
	ReplayFileStatusProcessing  ReplayFileStatus = "Processing"
	ReplayFileStatusParsed      ReplayFileStatus = "Parsed"      // Demo parsed, events written, awaiting aggregation
	ReplayFileStatusAggregating ReplayFileStatus = "Aggregating"  // Computing match/player stats
	ReplayFileStatusFailed      ReplayFileStatus = "Failed"
	ReplayFileStatusCompleted   ReplayFileStatus = "Completed"
	ReplayFileStatusDeleted     ReplayFileStatus = "Deleted"
)

// ReplayFileOptions contains optional parameters for creating a replay file
type ReplayFileOptions struct {
	Title            string                    `json:"title,omitempty"`
	Description      string                    `json:"description,omitempty"`
	Tags             []string                  `json:"tags,omitempty"`
	Visibility       shared.VisibilityTypeKey  `json:"visibility,omitempty"`
	ContentHash      string                    `json:"content_hash,omitempty"`
	OriginalReplayID *uuid.UUID                `json:"original_replay_id,omitempty"`
}

// NewReplayFile creates a new replay file entity with default public visibility
func NewReplayFile(gameID replay_common.GameIDKey, networkID replay_common.NetworkIDKey, size int, uri string, resourceOwner shared.ResourceOwner) *ReplayFile {
	return NewReplayFileWithOptions(gameID, networkID, size, uri, resourceOwner, nil)
}

// NewReplayFileWithOptions creates a new replay file entity with custom options
func NewReplayFileWithOptions(gameID replay_common.GameIDKey, networkID replay_common.NetworkIDKey, size int, uri string, resourceOwner shared.ResourceOwner, opts *ReplayFileOptions) *ReplayFile {
	var entity shared.BaseEntity
	
	// Determine visibility based on options
	if opts != nil && opts.Visibility != 0 {
		switch opts.Visibility {
		case shared.PrivateVisibilityTypeKey:
			entity = shared.NewPrivateEntity(resourceOwner)
		case shared.RestrictedVisibilityTypeKey:
			entity = shared.NewRestrictedEntity(resourceOwner)
		default:
			entity = shared.NewUnrestrictedEntity(resourceOwner)
		}
	} else {
		// Default to public visibility for guest uploads
		entity = shared.NewUnrestrictedEntity(resourceOwner)
	}
	
	rf := &ReplayFile{
		BaseEntity:  entity,
		GameID:      gameID,
		NetworkID:   networkID,
		Size:        size,
		InternalURI: uri,
		Status:      ReplayFileStatusPending,
		Error:       "",
		Header:      nil,
	}
	
	// Apply optional metadata
	if opts != nil {
		rf.Title = opts.Title
		rf.Description = opts.Description
		rf.Tags = opts.Tags
		rf.ContentHash = opts.ContentHash
		rf.OriginalReplayID = opts.OriginalReplayID
	}
	
	return rf
}

type ReplayFile struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	GameID            replay_common.GameIDKey    `json:"game_id" bson:"game_id"`
	NetworkID         replay_common.NetworkIDKey `json:"network_id" bson:"network_id"`
	Size              int                        `json:"size" bson:"size"`
	InternalURI       string                     `json:"uri" bson:"uri"`
	Status            ReplayFileStatus           `json:"status" bson:"status"`
	Error             string                     `json:"error" bson:"error"`
	Header            interface{}                `json:"header" bson:"header"`
	// User-editable metadata
	Title       string   `json:"title,omitempty" bson:"title,omitempty"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"`
	Tags        []string `json:"tags,omitempty" bson:"tags,omitempty"`
	// Deduplication support - SHA256 hash of file content
	ContentHash string `json:"content_hash,omitempty" bson:"content_hash,omitempty"`
	// Reference to original replay if this is a duplicate from different user
	OriginalReplayID *uuid.UUID `json:"original_replay_id,omitempty" bson:"original_replay_id,omitempty"`
}

func (r ReplayFile) GetID() uuid.UUID {
	return r.ID
}
