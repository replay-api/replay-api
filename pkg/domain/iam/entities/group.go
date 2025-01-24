package iam_entities

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

const (
	DefaultUserGroupName = "private:default"
)

type Group struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	Name          string               `json:"name" bson:"name"`
	Type          GroupType            `json:"type" bson:"type"`
	ParentGroupID *uuid.UUID           `json:"parent_group_id" bson:"parent_group_id"` // (REVIEW: already has a parent in the resource_owner)
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func NewGroup(groupID uuid.UUID, name string, groupType GroupType, resourceOwner common.ResourceOwner) *Group {
	resourceOwner.GroupID = groupID

	return &Group{
		ID:            groupID,
		Name:          name,
		Type:          groupType,
		ResourceOwner: resourceOwner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// func (e *Group) GetID() uuid.UUID {
// 	return e.ID
// }

func (e Group) GetID() uuid.UUID {
	return e.ID
}
