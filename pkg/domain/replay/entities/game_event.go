package entities

import (
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	replay_entities "github.com/replay-api/replay-common/pkg/replay/entities"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// Re-export GameEvent from replay-common for backward compatibility
type GameEvent = replay_entities.GameEvent

// Re-export game types
type GameIDKey = replay_common.GameIDKey
type TickIDType = replay_common.TickIDType

// Re-export game constants
const (
	CS2_GAME_ID   = replay_common.CS2_GAME_ID
	CSGO_GAME_ID  = replay_common.CSGO_GAME_ID
	VLRNT_GAME_ID = replay_common.VLRNT_GAME_ID
)

// Re-export network types
type NetworkIDKey = replay_common.NetworkIDKey

// Re-export network constants
const (
	SteamNetworkIDKey     = replay_common.SteamNetworkIDKey
	FaceItNetworkIDKey    = replay_common.FaceItNetworkIDKey
	BattleNetNetworkIDKey = replay_common.BattleNetNetworkIDKey
)

// Helper function to create GameEvent using replay-common
func NewGameEvent[T any](matchID uuid.UUID, tickID replay_common.TickIDType, gameTime time.Duration, eventType fps_events.EventIDKey, payload T, entities map[shared.ResourceType][]interface{}, stats map[replay_common.StatType][]interface{}, res shared.ResourceOwner) *replay_entities.GameEvent {
	return replay_entities.NewGameEvent(matchID, tickID, gameTime, eventType, payload, entities, stats, res)
}
