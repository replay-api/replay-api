package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type SquadRole string

const (
	RoleOwner SquadRole = "Owner"
	RoleAdmin SquadRole = "Admin"
	RoleUser  SquadRole = "Member"
)

type Squad struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	Name          string               `json:"name" bson:"name"`
	UserIDs       []uuid.UUID          `json:"users" bson:"users"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

type SquadMember struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	SquadID       uuid.UUID            `json:"squad_id" bson:"squad_id"`
	UserID        uuid.UUID            `json:"user_id" bson:"user_id"`
	Role          SquadRole            `json:"role" bson:"role"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}
