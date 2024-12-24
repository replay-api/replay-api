package squad_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type MembershipType string

const (
	MembershipTypeOwner  MembershipType = "Owner"
	MembershipTypeAdmin  MembershipType = "Admin"
	MembershipTypeMember MembershipType = "Member"
)

type Membership struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	SquadID       uuid.UUID            `json:"squad_id" bson:"squad_id"`
	UserID        uuid.UUID            `json:"user_id" bson:"user_id"`
	Type          MembershipType       `json:"type" bson:"type"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}
