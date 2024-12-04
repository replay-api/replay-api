package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type GroupType string

const (
	GroupTypeUser   GroupType = "User"   // Team*, Profile*, Channel*, Page*, Friends*, Me* (Self), Custom-Private (Custom User-defined)
	GroupTypeSystem GroupType = "System" // Public, Public(Anyone with the link, link/:slug-id route), Private, Namespace (directory/path trees), TagXyz, Friends, BugReport#1, Users(Region,Match, etc... ==> tag!! user-defined tag ())
)

type Group struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	Name          string               `json:"name" bson:"name"`
	Type          GroupType            `json:"type" bson:"type"`
	Parent        common.ResourceOwner `json:"parent" bson:"parent"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (e Group) GetID() uuid.UUID {
	return e.ID
}
