package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type GameEvent struct {
	// header/meta
	ID       uuid.UUID         `json:"id" bson:"_id"`
	Type     common.EventIDKey `json:"type" bson:"type"`
	GameID   common.GameIDKey  `json:"game_id" bson:"game_id"`
	MatchID  uuid.UUID         `json:"match_id" bson:"match_id"`
	TickID   common.TickIDType `json:"tick_id" bson:"tick_id"`
	GameTime time.Duration     `json:"event_time" bson:"event_time"` // // CurrentTime

	// data
	Payload  interface{}                           `json:"-" bson:"payload"`
	Entities map[common.ResourceType][]interface{} `json:"-" bson:"-"`
	Stats    map[common.StatType][]interface{}     `json:"stats" bson:"stats"`

	// default/trail
	ResourceOwner common.ResourceOwner `json:"-" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"-" bson:"created_at"`
}

func NewGameEvent[T any](matchID uuid.UUID, tickID common.TickIDType, gameTime time.Duration, eventType common.EventIDKey, payload T, entities map[common.ResourceType][]interface{}, stats map[common.StatType][]interface{}, res common.ResourceOwner) *GameEvent {
	return &GameEvent{
		GameID:        common.CS2_GAME_ID, // TODO: refact => quando aplicavel para go/vlr
		MatchID:       matchID,
		TickID:        tickID,
		Type:          eventType,
		GameTime:      gameTime,
		Payload:       payload,
		Entities:      entities,
		Stats:         stats,
		ResourceOwner: res,
		CreatedAt:     time.Now(),
	}
}

func (ge *GameEvent) GetPlayerIDs() ([]common.PlayerIDType, error) {
	players, ok := ge.Entities[common.ResourceTypePlayer]

	if !ok {
		return nil, fmt.Errorf("PlayerID not present in GameEvent %v", ge)
	}

	playerIDs := make([]common.PlayerIDType, len(players))

	for _, p := range players {
		playerIDs = append(playerIDs, common.PlayerIDType(p.(Player).GetID()))
	}

	return playerIDs, nil
}

func (e GameEvent) GetID() uuid.UUID {
	return e.ID
}
