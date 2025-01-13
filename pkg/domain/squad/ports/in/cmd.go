package squad_in

import (
	"context"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

// type CreateSquadCommandHandler interface {
// 	Handle(command CreateSquadCommand) error
// }

type CreateSquadCommand struct {
	Name        string                                                   `json:"name"`
	Symbol      string                                                   `json:"symbol"`
	Description string                                                   `json:"description"`
	GameID      common.GameIDKey                                         `json:"game_id"`
	AvatarURI   string                                                   `json:"avatar_uri"`
	Members     map[iam_entities.UserIDKey]squad_entities.MembershipType `json:"members"`
}

type CreateSquadCommandHandler interface {
	Exec(c context.Context, cmd CreateSquadCommand) (*squad_entities.Squad, error)
}
