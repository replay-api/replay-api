package iam_entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type GroupType string

const (
	GroupTypeAccount GroupType = "Account"
	GroupTypeUser    GroupType = "User"   // Team* (!= Squad, WorkTeam, ETC), Profile*, Channel*, Page*, Friends*, Me* (Self), Custom-Private (Custom User-defined)
	GroupTypeSystem  GroupType = "System" // Public, Public(Anyone with the link, link/:slug-id route), Private, Namespace (directory/path trees), TagXyz, Friends, BugReport#1, Users(Region,Match, etc... ==> tag!! user-defined tag ())
)

const (
	DefaultUserGroupName = "private:default"
)

type Group struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	Name          string               `json:"name" bson:"name"`
	Type          GroupType            `json:"type" bson:"type"`
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

func NewAccountGroup(groupID uuid.UUID, resourceOwner common.ResourceOwner) *Group {
	return NewGroup(groupID, DefaultUserGroupName, GroupTypeAccount, resourceOwner)
}

func (e Group) GetID() uuid.UUID {
	return e.ID
}

func NewGroupAccountSearchByUser(ctx context.Context) common.Search {
	return common.Search{
		SearchParams: []common.SearchAggregation{
			{
				Params: []common.SearchParameter{
					{
						ValueParams: []common.SearchableValue{
							{
								Field:    "Type",
								Operator: common.EqualsOperator,
								Values:   []interface{}{GroupTypeAccount},
							},
						},
					},
				},
			},
		},
		ResultOptions: common.SearchResultOptions{
			Limit: 1,
		},
		VisibilityOptions: common.SearchVisibilityOptions{
			RequestSource:    common.GetResourceOwner(ctx),
			IntendedAudience: common.UserAudienceIDKey,
		},
	}
}
