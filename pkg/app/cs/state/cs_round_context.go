package state

import (
	"fmt"

	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

type CSRoundType = string

const (
	CSRoundTypePistol CSRoundType = "pistol"
	CSRoundTypeLast   CSRoundType = "last"
)

type CS2RoundContext struct {
	RoundNumber         int
	RoundType           CSRoundType
	WinnerNetworkTeamID string // TODO: ver qual campo/type identifica o id  do time na steam pela demo (p/ Name => ie.: p.GameState().TeamCounterTerrorists().ClanName())
	Clutch              *CS2ClutchContext
	PlayerEntities      []*replay_entity.PlayerMetadata
	TeamT               cs_entity.TeamHashIDType
	TeamCT              cs_entity.TeamHashIDType
	TeamContext         map[cs_entity.TeamHashIDType]*CSTeamContext
	BattleContext       *CS2BattleContext
}

// TODO: adicionar parametro para tipo de rede, habilitar demais provides (fcit etc)
func (r *CS2RoundContext) SetPlayingEntities(playing []*infocs.Player, res common.ResourceOwner) []*replay_entity.PlayerMetadata {
	r.PlayerEntities = make([]*replay_entity.PlayerMetadata, len(playing))

	for index, p := range playing {
		r.PlayerEntities[index] = replay_entity.NewPlayerMetadata(
			p.Name,
			fmt.Sprintf("%d", p.SteamID64),
			common.SteamNetworkIDKey,
			p.ClanTag(),
			res,
		)
	}

	return r.PlayerEntities
}

func (r *CS2RoundContext) GetUntypedPlayingEntities() []interface{} {
	playersAsInterface := make([]interface{}, len(r.PlayerEntities))

	for index, p := range r.PlayerEntities {
		playersAsInterface[index] = p
	}

	return playersAsInterface
}

func (r *CS2RoundContext) SetRoundType(roundType CSRoundType) {
	r.RoundType = roundType
}

func (r *CS2RoundContext) GetRoundNumber() int {
	return r.RoundNumber
}

func (r *CS2RoundContext) GetWinnerNetworkTeamID() string {
	return r.WinnerNetworkTeamID
}

func (r *CS2RoundContext) GetClutch() *CS2ClutchContext {
	return r.Clutch
}
