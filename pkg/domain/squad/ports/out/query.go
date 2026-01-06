package squad_out

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

type SquadReader interface {
	shared.Searchable[squad_entities.Squad]
}

type PlayerProfileReader interface {
	shared.Searchable[squad_entities.PlayerProfile]
}
