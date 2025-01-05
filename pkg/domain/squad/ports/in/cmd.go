package squad_in

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

// type CreateSquadCommandHandler interface {
// 	Handle(command CreateSquadCommand) error
// }

type CreateSquadCommand struct {
	FullName    string                                      `json:"full_name"`
	ShortName   string                                      `json:"short_name"`
	Symbol      string                                      `json:"symbol"`
	Description string                                      `json:"description"`
	GameID      uuid.UUID                                   `json:"game_id"`
	AvatarURI   string                                      `json:"avatar_uri"`
	Members     map[uuid.UUID]squad_entities.MembershipType `json:"members"`
}

type CreateSquadCommandHandler interface {
	Exec(c context.Context, cmd CreateSquadCommand) (*squad_entities.Squad, error)
}
