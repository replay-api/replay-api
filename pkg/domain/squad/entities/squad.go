package squad_entities

import (
	"context"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type Squad struct {
	common.BaseEntity
	GameID      common.GameIDKey                      `json:"game_id" bson:"game_id"`
	Name        string                                `json:"name" bson:"name"`
	Symbol      string                                `json:"symbol" bson:"symbol"`
	Description string                                `json:"description" bson:"description"`
	LogoURI     string                                `json:"logo_uri" bson:"logo_uri"`
	SlugURI     string                                `json:"slug_uri" bson:"slug_uri"`
	BannerURI   string                                `json:"banner_uri" bson:"banner_uri"` // TODO: create media collection, for multiple purposes
	Membership  []squad_value_objects.SquadMembership `json:"membership" bson:"membership"`
}

func NewSquad(groupID uuid.UUID, gameID common.GameIDKey, logorURI, name, symbol, description, slugURI string, membership []squad_value_objects.SquadMembership, rxn common.ResourceOwner) *Squad {
	squad := Squad{
		BaseEntity:  common.NewUnrestrictedEntity(rxn),
		GameID:      gameID,
		Name:        name,
		Symbol:      symbol,
		SlugURI:     slugURI,
		Description: description,
		Membership:  membership,
	}

	return &squad
}

func (e Squad) GetID() uuid.UUID {
	return e.ID
}

func NewSearchByName(ctx context.Context, name string) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "Name",
							Values: []interface{}{
								name,
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
