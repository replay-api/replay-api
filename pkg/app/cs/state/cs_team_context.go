package state

import (
	"fmt"

	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
)

// type

type CSTeamContext struct {
	TeamID         cs_entity.TeamIDType
	TeamHashID     cs_entity.TeamHashIDType
	EconomyContext *CSEconomyContext
	Side           cs_entity.CSTeamSideIDType
}

func NewCSTeamContext(players []*infocs.Player) *CSTeamContext {
	var h cs_entity.TeamHashIDType
	ids := make([]string, len(players))

	team := CSTeamContext{
		EconomyContext: NewCSEconomyContext(players),
	}

	for i, p := range players {
		ids = append(ids, fmt.Sprintf("%d", p.SteamID64))
		for k := i; k > 0; k-- {
			if ids[k] > ids[k-1] {
				aux := ids[k-1]
				ids[k-1] = ids[k]
				ids[k] = aux
			}
		}

	}

	for i := 0; i < len(ids); i++ {
		h = fmt.Sprintf("%s_%s", h, ids[i])
	}

	team.TeamHashID = h

	return &team
}
