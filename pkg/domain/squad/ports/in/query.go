package squad_in

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type SquadSearchableReader interface {
	common.Searchable[squad_entities.Squad]
}
