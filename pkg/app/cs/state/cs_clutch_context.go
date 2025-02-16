package state

import (
	"fmt"
	"log/slog"

	cs2 "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
)

type ClutchSituationTypeKey string

const (
	ClutchType1v2 ClutchSituationTypeKey = "1v2"
	ClutchType1v3 ClutchSituationTypeKey = "1v3"
	ClutchType1v4 ClutchSituationTypeKey = "1v4"
	ClutchType1v5 ClutchSituationTypeKey = "1v5"
)

var (
	ClutchTypeMap map[int]ClutchSituationTypeKey = map[int]ClutchSituationTypeKey{
		2: ClutchType1v2,
		3: ClutchType1v3,
		4: ClutchType1v4,
		5: ClutchType1v5,
	}
)

type CS2ClutchContext struct {
	RoundNumber int
	Player      *cs2.Player
	Opponents   []cs2.Player
	Status      cs_entity.ClutchSituationStatusKey
	Type        ClutchSituationTypeKey
}

func NewCS2ClutchContext(roundNumber int, playerInClutch *cs2.Player, opponents []cs2.Player) *CS2ClutchContext {
	hasPlayerInClutchSituation := playerInClutch != nil

	clutchSituationState := cs_entity.ClutchInitiatedKey

	clutchSituationType, ok := ClutchTypeMap[len(opponents)]

	if !ok || !hasPlayerInClutchSituation {
		clutchSituationState = cs_entity.NotInClutchSituation
		msg := "(NewCSClutchContext) Error initializing ClutchContext entity: Unexpected CS2 ClutchSituationType: Mapping not found a/or not supported for [hasPlayerInClutchSituation=%v] with #%d opponents at round #%d. (Supported: %v))"
		slog.Info(fmt.Sprintf(msg, hasPlayerInClutchSituation, len(opponents), roundNumber, ClutchTypeMap))
	}

	return &CS2ClutchContext{
		RoundNumber: roundNumber,
		Status:      clutchSituationState,
		Player:      playerInClutch,
		Opponents:   opponents,
		Type:        clutchSituationType,
	}
}

func (c *CS2ClutchContext) GetRoundNumber() int {
	return c.RoundNumber
}

func (c *CS2ClutchContext) GetPlayer() *cs2.Player {
	return c.Player
}

func (c *CS2ClutchContext) GetNetworkPlayerID() uint64 {
	return (*c.Player).SteamID64
}

func (c *CS2ClutchContext) GetOpponents() []cs2.Player {
	return c.Opponents
}

func (c *CS2ClutchContext) SetPlayer(player *cs2.Player) {
	c.Player = player
}

func (c *CS2ClutchContext) SetOpponents(opponents []cs2.Player) {
	c.Opponents = opponents
}

func (c *CS2ClutchContext) SetStatus(status cs_entity.ClutchSituationStatusKey) {
	c.Status = status
}
