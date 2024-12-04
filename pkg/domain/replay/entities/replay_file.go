package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type ReplayFileStatus string

const (
	ReplayFileStatusPending    ReplayFileStatus = "Pending"
	ReplayFileStatusProcessing ReplayFileStatus = "Processing"
	ReplayFileStatusFailed     ReplayFileStatus = "Failed"
	ReplayFileStatusCompleted  ReplayFileStatus = "Completed"
)

func NewReplayFile(gameID common.GameIDKey, networkID common.NetworkIDKey, size int, uri string, resourceOwner common.ResourceOwner) ReplayFile {
	entity := common.NewEntity(resourceOwner)
	return ReplayFile{
		ID:            entity.ID,
		GameID:        gameID,
		NetworkID:     networkID,
		Size:          size,
		InternalURI:   uri,
		Status:        ReplayFileStatusPending,
		Error:         "",
		Header:        nil,
		ResourceOwner: resourceOwner,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}

type ReplayFile struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
	GameID        common.GameIDKey     `json:"game_id" bson:"game_id"`
	NetworkID     common.NetworkIDKey  `json:"network_id" bson:"network_id"`
	Size          int                  `json:"size" bson:"size"`
	InternalURI   string               `json:"uri" bson:"uri"`
	Status        ReplayFileStatus     `json:"status" bson:"status"`
	Error         string               `json:"error" bson:"error"`
	Header        interface{}          `json:"header" bson:"header"`
}

func (r ReplayFile) GetID() uuid.UUID {
	return r.ID
}
