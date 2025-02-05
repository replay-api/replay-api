package squad_in

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type SquadReader interface {
	common.Searchable[squad_entities.Squad]
}

type PlayerReader interface {
	common.Searchable[squad_entities.PlayerProfile]
}
