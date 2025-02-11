package squad_entities

import (
	"context"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type PlayerHistoryAction string

const (
	PlayerHistoryActionCreate                 PlayerHistoryAction = "Create"
	PlayerHistoryActionUpdate                 PlayerHistoryAction = "Update"
	PlayerHistoryActionDelete                 PlayerHistoryAction = "Delete"
	PlayerHistoryActionVisibilityRestricted   PlayerHistoryAction = "VisibilityRestricted"
	PlayerHistoryActionVisibilityUnrestricted PlayerHistoryAction = "VisibilityUnrestricted"
)

type PlayerProfile struct {
	common.BaseEntity
	GameID      common.GameIDKey `json:"game_id" bson:"game_id"`
	Nickname    string           `json:"nickname" bson:"nickname"`
	SlugURI     string           `json:"slug_uri" bson:"slug_uri"`
	Avatar      string           `json:"avatar" bson:"avatar"`
	Roles       []string         `json:"roles" bson:"roles"`
	Description string           `json:"description" bson:"description"`
}

func (e PlayerProfile) GetID() uuid.UUID {
	return e.BaseEntity.ID
}

func NewPlayerProfile(gameID common.GameIDKey, nickname, avatar, slugURI, description string, roles []string, visbility common.VisibilityTypeKey, rxn common.ResourceOwner) *PlayerProfile {
	var baseEntity common.BaseEntity

	switch visbility {
	case common.PublicVisibilityTypeKey:
		baseEntity = common.NewUnrestrictedEntity(rxn)
	case common.RestrictedVisibilityTypeKey:
		baseEntity = common.NewRestrictedEntity(rxn)
	case common.PrivateVisibilityTypeKey:
		baseEntity = common.NewPrivateEntity(rxn)
	case common.CustomVisibilityTypeKey:
		baseEntity = common.NewEntity(rxn)
	default:
		baseEntity = common.NewEntity(rxn)
	}

	return &PlayerProfile{
		BaseEntity:  baseEntity,
		GameID:      gameID,
		Nickname:    nickname,
		SlugURI:     slugURI,
		Avatar:      avatar,
		Description: description,
		Roles:       roles,
	}
}

func NewSearchByNickname(ctx context.Context, nickname string) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "Nickname",
							Values: []interface{}{
								nickname,
							},
						},
					},
				},
			},
		},
	}

	visibility := common.SearchVisibilityOptions{
		RequestSource:    common.GetResourceOwner(ctx),
		IntendedAudience: common.ClientApplicationAudienceIDKey,
	}

	result := common.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return common.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

func NewSearchBySlugURI(ctx context.Context, slugURI string) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "SlugURI",
							Values: []interface{}{
								slugURI,
							},
						},
					},
				},
			},
		},
	}

	visibility := common.SearchVisibilityOptions{
		RequestSource:    common.GetResourceOwner(ctx),
		IntendedAudience: common.ClientApplicationAudienceIDKey,
	}

	result := common.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return common.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}
