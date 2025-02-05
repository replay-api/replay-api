package squad_out

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type SquadReader interface {
	common.Searchable[squad_entities.Squad]
}

type PlayerProfileReader interface {
	common.Searchable[squad_entities.PlayerProfile]
}
