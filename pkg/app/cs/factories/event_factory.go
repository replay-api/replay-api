package factories

import (
	"log/slog"
	"time"

	"github.com/replay-api/replay-api/pkg/app/cs/state"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

func NewGameEvent[T any](eventType fps_events.EventIDKey, matchContext *state.CS2MatchContext, roundIndex int, tickID replay_common.TickIDType, gameTime time.Duration, payload T) (*replay_entity.GameEvent, error) {
	entities := make(map[shared.ResourceType][]interface{})
	stats := make(map[replay_common.StatType][]interface{})

	roundContext := matchContext.RoundContexts[roundIndex]

	if roundContext == nil {
		slog.Info("test roundContext == nil", "len(matchContext.RoundContexts)", len(matchContext.RoundContexts))
		panic("nilroundcontext")
	}

	entities[replay_common.ResourceTypePlayerMetadata] = roundContext.GetUntypedPlayingEntities()

	battleContext := roundContext.BattleContext
	battleStats, err := battleContext.GetStatistics()

	if err != nil {
		slog.Error("unable to read stats from battle context")
		return nil, err
	}

	stats[replay_common.BattleStatTypeKey] = append(stats[replay_common.BattleStatTypeKey], battleStats)

	// stats[shared.BattleStatTypeKey] =  // todo: definir + criar ticket + doc + (trade logic tá aqui) (* principais stats) -> stats em serie atendem: graficos principalmente (progressão de cada stat, replay/highlights*)
	// stats[shared.UtilityStatTypeKey] =

	// switch eventType {
	// // case fps_events.Event_MatchStartID:
	// // case shared.Event_Economy:
	// // 	// return NewEconomyGameEvent
	// case shared.Event_RoundEndID:
	// case shared.Event_FragOrScoreID:
	// 	stats[shared.BattleStatTypeKey] = battleContext.

	// // 	// default:
	// // 	// 	entities
	// // 	// TODO: gerar Stats default (Players, Round etc para todos os eventos com entidade)
	// // }

	return replay_entity.NewGameEvent(matchContext.MatchID, tickID, gameTime, eventType, payload, entities, stats, matchContext.ResourceOwner), nil
}

// func NewEconomyGameEvent(matchID uuid.UUID, ) {
// 	ID:            uuid.New(),
// 	MatchID:       matchContext.MatchID,
// 	Type:          shared.Event_RoundEndID,
// 	Time:          p.CurrentTime(),
// 	Body:     event,
// 	Entities:
// 	ResourceOwner: matchContext.ResourceOwner,
// }
