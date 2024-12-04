package state

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
)

type CS2MatchContext struct {
	MatchID       uuid.UUID `json:"match_id"`
	Header        cs_entity.CSReplayFileHeader
	RoundContexts map[int]*CS2RoundContext `json:"round_contexts"`
	ResourceOwner common.ResourceOwner     `json:"resource_owner"`
}

func NewCS2MatchContext(userContext context.Context, matchID uuid.UUID) *CS2MatchContext {
	return &CS2MatchContext{
		MatchID:       matchID,
		RoundContexts: make(map[int]*CS2RoundContext),
		ResourceOwner: common.GetResourceOwner(userContext),
	}
}

func (m *CS2MatchContext) AddRoundContext(roundID int, roundContext *CS2RoundContext) {
	m.RoundContexts[roundID] = roundContext
}

func GetTeamIds(playing []*infocs.Player) []string {
	teamIds := make([]cs_entity.TeamHashIDType, 0)

	var tempTeamId cs_entity.TeamHashIDType
	for _, player := range playing {
		tempTeamId = fmt.Sprintf("%s-%d", player.TeamState.ClanName(), player.TeamState.ID())

		teamIds = append(teamIds, tempTeamId)
	}

	return teamIds
}

func (m *CS2MatchContext) WithRound(roundIndex int, gs dem.GameState) *CS2MatchContext {
	playing := gs.Participants().Playing()

	teamCT := gs.TeamCounterTerrorists()
	teamT := gs.TeamTerrorists()

	ctID := fmt.Sprintf("%s-%d", teamCT.ClanName(), teamCT.ID())
	tID := fmt.Sprintf("%s-%d", teamT.ClanName(), teamT.ID())

	_, ok := m.RoundContexts[roundIndex]
	if ok {
		return m
	}

	roundNumber := roundIndex + 1

	roundContext := CS2RoundContext{
		RoundNumber:         roundNumber,
		WinnerNetworkTeamID: "",
		Clutch: &CS2ClutchContext{
			RoundNumber: roundNumber,
			Status:      cs_entity.NotInClutchSituation,
		},
		TeamContext: make(map[cs_entity.TeamHashIDType]*CSTeamContext),
		BattleContext: &CS2BattleContext{
			Hits: make(map[common.TickIDType]cs_entity.CSHitStats),
		},
		TeamT:  tID,
		TeamCT: ctID,
	}

	roundContext.SetPlayingEntities(playing, m.ResourceOwner)

	// mover p'/ SetPlayingEntities
	teamsIds := make([]cs_entity.TeamHashIDType, 0)
	playersByTeam := make(map[cs_entity.TeamHashIDType][]*infocs.Player)

	var tempTeamId cs_entity.TeamHashIDType
	for _, player := range playing {
		tempTeamId = fmt.Sprintf("%s-%d", player.TeamState.ClanName(), player.TeamState.ID())

		teamsIds = append(teamsIds, tempTeamId)
		playersByTeam[tempTeamId] = append(playersByTeam[tempTeamId], player)
	}

	for _, tempTeamId := range teamsIds {
		roundContext.TeamContext[tempTeamId] = NewCSTeamContext(playersByTeam[tempTeamId])

		if tempTeamId == tID {
			roundContext.TeamT = tID
		} else if tempTeamId == ctID {
			roundContext.TeamCT = ctID
		}
	}

	if roundNumber == 1 || roundNumber == 16 {
		roundContext.SetRoundType(CSRoundTypePistol)
	}
	// mover

	m.AddRoundContext(roundIndex, &roundContext)

	return m
}

func (m *CS2MatchContext) SetHeader(h cs_entity.CSReplayFileHeader) {
	m.Header = h
}

func (m *CS2MatchContext) SetClutchRoundContext(roundIndex int, clutchStatus cs_entity.ClutchSituationStatusKey) {
	roundContext, ok := m.RoundContexts[roundIndex]

	if !ok {
		return
	}

	roundContext.Clutch.Status = clutchStatus
}

func (m *CS2MatchContext) WithClutch(roundIndex int, playerInClutch *infocs.Player, opponents []infocs.Player) *CS2MatchContext {
	_, ok := m.RoundContexts[roundIndex]
	roundNumber := roundIndex + 1

	if !ok {
		m.AddRoundContext(roundIndex, &CS2RoundContext{
			Clutch:      NewCS2ClutchContext(roundNumber, playerInClutch, opponents),
			RoundNumber: roundNumber,
		})

		return m
	}

	roundContext := m.RoundContexts[roundIndex]

	if roundContext.Clutch == nil || roundContext.Clutch.Status == cs_entity.NotInClutchSituation {
		roundContext.Clutch = NewCS2ClutchContext(roundNumber, playerInClutch, opponents)
	}

	m.RoundContexts[roundIndex] = roundContext

	return m
}

func (m *CS2MatchContext) GetClutchPlayer(roundIndex int) *infocs.Player {
	return m.RoundContexts[roundIndex].Clutch.Player
}

func (m *CS2MatchContext) GetClutchRoundContext(roundIndex int) *CS2RoundContext {
	return m.RoundContexts[roundIndex]
}

func (m *CS2MatchContext) UpdateClutchState(roundIndex int, clutchStatus cs_entity.ClutchSituationStatusKey, opponents []infocs.Player) *CS2MatchContext {
	roundContext, ok := m.RoundContexts[roundIndex]

	if !ok {
		msg := "Round context not found to UpdateClutch"
		panic(msg)
		// return m
	}

	roundContext.Clutch.Status = clutchStatus

	return m
}

func (m *CS2MatchContext) InClutch(roundIndex int) bool {
	roundContext, ok := m.RoundContexts[roundIndex]

	if !ok {
		return false
	}

	return roundContext.Clutch.Status != cs_entity.NotInClutchSituation
}
