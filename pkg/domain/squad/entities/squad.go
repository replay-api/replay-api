package squad_entities

import (
	"context"

	"github.com/google/uuid"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type Squad struct {
	shared.BaseEntity
	GameID      replay_common.GameIDKey             `json:"game_id" bson:"game_id"`
	Name        string                              `json:"name" bson:"name"`
	Symbol      string                              `json:"symbol" bson:"symbol"`
	Description string                              `json:"description" bson:"description"`
	LogoURI     string                              `json:"logo_uri" bson:"logo_uri"`
	SlugURI     string                              `json:"slug_uri" bson:"slug_uri"`
	BannerURI   string                              `json:"banner_uri" bson:"banner_uri"` // TODO: create media collection, for multiple purposes
	Membership  []squad_value_objects.SquadMembership `json:"membership" bson:"membership"`
	// TODO: regions
	// TODO: countries
	// TODO: languagues
	// TODO: genders
	// TODO: leagues?
}

func NewSquad(groupID uuid.UUID, gameID replay_common.GameIDKey, logorURI, name, symbol, description, slugURI string, membership []squad_value_objects.SquadMembership, rxn shared.ResourceOwner) *Squad {
	squad := Squad{
		BaseEntity:  shared.NewUnrestrictedEntity(rxn),
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

func NewSearchByName(ctx context.Context, name string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
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
