package squad_in

import (
	"context"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type CreateSquadCommand struct {
	Name        string                                `json:"name"`
	Symbol      string                                `json:"symbol"`
	Description string                                `json:"description"`
	GameID      common.GameIDKey                      `json:"game_id"`
	AvatarURI   string                                `json:"avatar_uri"`
	SlugURI     string                                `json:"slug_uri"`
	Members     map[string]CreateSquadMembershipInput `json:"members"`
}

type CreateSquadMembershipInput struct {
	Type  squad_value_objects.SquadMembershipType `json:"type" bson:"type"`
	Roles []string                                `json:"role" bson:"role"`
}

type CreateSquadCommandHandler interface {
	Exec(c context.Context, cmd CreateSquadCommand) (*squad_entities.Squad, error)
}

type CreatePlayerProfileCommand struct {
	GameID         common.GameIDKey         `json:"game_id"`
	Nickname       string                   `json:"nickname"`
	AvatarURI      string                   `json:"avatar_uri"`
	SlugURI        string                   `json:"slug_uri"`
	Roles          []string                 `json:"roles"`
	Description    string                   `json:"description"`
	VisibilityType common.VisibilityTypeKey `json:"visibility_type"`
}

type CreatePlayerProfileCommandHandler interface {
	Exec(c context.Context, cmd CreatePlayerProfileCommand) (*squad_entities.PlayerProfile, error)
}
