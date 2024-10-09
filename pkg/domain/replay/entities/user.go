package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

// This Relates User to Player in a many-to-many relationship
type ExternalProfile struct {
	NetworkID     string `json:"network_id" bson:"network_id"`
	NetworkUserID string `json:"network_player_id" bson:"network_player_id"`
}

type User struct {
	ID          uuid.UUID `json:"id" bson:"_id"`
	Username    string    `json:"username" bson:"username"`
	DisplayName string    `json:"display_name" bson:"display_name"`

	ExternalProfiles []ExternalProfile

	// EmailAddress string `json:"-" bson:"email_address"` // TODO: for email signin?

	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}
