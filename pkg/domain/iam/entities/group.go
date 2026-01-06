package iam_entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
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
	ResourceOwner shared.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func NewGroup(groupID uuid.UUID, name string, groupType GroupType, resourceOwner shared.ResourceOwner) *Group {
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

func NewAccountGroup(groupID uuid.UUID, resourceOwner shared.ResourceOwner) *Group {
	return NewGroup(groupID, DefaultUserGroupName, GroupTypeAccount, resourceOwner)
}

func (e Group) GetID() uuid.UUID {
	return e.ID
}

func NewGroupAccountSearchByUser(ctx context.Context) shared.Search {
	return shared.Search{
		SearchParams: []shared.SearchAggregation{
			{
				Params: []shared.SearchParameter{
					{
						ValueParams: []shared.SearchableValue{
							{
								Field:    "Type",
								Operator: shared.EqualsOperator,
								Values:   []interface{}{GroupTypeAccount},
							},
						},
					},
				},
			},
		},
		ResultOptions: shared.SearchResultOptions{
			Limit: 1,
		},
		VisibilityOptions: shared.SearchVisibilityOptions{
			RequestSource:    shared.GetResourceOwner(ctx),
			IntendedAudience: shared.UserAudienceIDKey,
		},
	}
}
