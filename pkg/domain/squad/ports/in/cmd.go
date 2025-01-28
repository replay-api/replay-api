package squad_in

import (
	"context"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type CreateSquadCommand struct {
	Name        string                                                             `json:"name"`
	Symbol      string                                                             `json:"symbol"`
	Description string                                                             `json:"description"`
	GameID      common.GameIDKey                                                   `json:"game_id"`
	AvatarURI   string                                                             `json:"avatar_uri"`
	Members     map[squad_entities.PlayerIDKey]squad_value_objects.SquadMembership `json:"members"`
}

type CreateSquadCommandHandler interface {
	Exec(c context.Context, cmd CreateSquadCommand) (*squad_entities.Squad, error)
}

type CreatePlayerCommand struct {
	GameID          common.GameIDKey `json:"game_id"`
	Nickname        string           `json:"nickname"`
	AvatarURI       string           `json:"avatar_uri"`
	VisibilityType  common.VisibilityTypeKey
	VisibilityLevel common.IntendedAudienceKey
}

type CreatePlayerCommandHandler interface {
	Exec(c context.Context, cmd CreatePlayerCommand) (*squad_entities.Player, error)
}
