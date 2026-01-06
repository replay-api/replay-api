package squad_entities

import (
	"context"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
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
	shared.BaseEntity
	GameID      replay_common.GameIDKey `json:"game_id" bson:"game_id"`
	Nickname    string                  `json:"nickname" bson:"nickname"`
	SlugURI     string                  `json:"slug_uri" bson:"slug_uri"`
	Avatar      string                  `json:"avatar" bson:"avatar"`
	Roles       []string                `json:"roles" bson:"roles"`
	Description string                  `json:"description" bson:"description"`
	NetworkIDs  map[string]string       `json:"-" bson:"network_ids"`
	// TODO: regions?
	// TODO: country!
	// TODO: languagues
	// TODO: gender (optional)
	// TODO: Status Or flag indicating to request participation in Squad (and/or status: looking for squad, looking for friends etcs)
}

func (e PlayerProfile) GetID() uuid.UUID {
	return e.BaseEntity.ID
}

func NewPlayerProfile(gameID replay_common.GameIDKey, nickname, avatar, slugURI, description string, roles []string, visbility shared.VisibilityTypeKey, rxn shared.ResourceOwner) *PlayerProfile {
	var baseEntity shared.BaseEntity

	switch visbility {
	case shared.PublicVisibilityTypeKey:
		baseEntity = shared.NewUnrestrictedEntity(rxn)
	case shared.RestrictedVisibilityTypeKey:
		baseEntity = shared.NewRestrictedEntity(rxn)
	case shared.PrivateVisibilityTypeKey:
		baseEntity = shared.NewPrivateEntity(rxn)
	case shared.CustomVisibilityTypeKey:
		baseEntity = shared.NewEntity(rxn)
	default:
		baseEntity = shared.NewEntity(rxn)
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

func NewSearchByNickname(ctx context.Context, nickname string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
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

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

func NewSearchByID(ctx context.Context, id uuid.UUID) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field: "ID",
							Values: []interface{}{
								id,
							},
						},
					},
				},
			},
		},
	}

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

func NewNicknameAndSlugExistenceCheck(ctx context.Context, id uuid.UUID, nickname, sluguri string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					AggregationClause: shared.OrAggregationClause,
					ValueParams: []shared.SearchableValue{
						{
							Field: "Nickname",
							Values: []interface{}{
								nickname,
							},
						},
						{
							Field: "SlugURI",
							Values: []interface{}{
								sluguri,
							},
						},
					},
				},
			},
		},
	}

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 2,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

func NewSearchBySlugURI(ctx context.Context, slugURI string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
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

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}
