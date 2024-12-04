package factories

import (
	"log/slog"
	"time"

	"github.com/psavelis/team-pro/replay-api/pkg/app/cs/state"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

func NewGameEvent[T any](eventType common.EventIDKey, matchContext *state.CS2MatchContext, roundIndex int, tickID common.TickIDType, gameTime time.Duration, payload T) (*replay_entity.GameEvent, error) {
	entities := make(map[common.ResourceType][]interface{})
	stats := make(map[common.StatType][]interface{})

	roundContext := matchContext.RoundContexts[roundIndex]

	if roundContext == nil {
		slog.Info("test roundContext == nil", "len(matchContext.RoundContexts)", len(matchContext.RoundContexts))
		panic("nilroundcontext")
	}

	entities[common.ResourceTypePlayer] = roundContext.GetUntypedPlayingEntities()

	battleContext := roundContext.BattleContext
	battleStats, err := battleContext.GetStatistics()

	if err != nil {
		slog.Error("unable to read stats from battle context")
		return nil, err
	}

	stats[common.BattleStatTypeKey] = append(stats[common.BattleStatTypeKey], battleStats)

	// stats[common.BattleStatTypeKey] =  // todo: definir + criar ticket + doc + (trade logic tá aqui) (* principais stats) -> stats em serie atendem: graficos principalmente (progressão de cada stat, replay/highlights*)
	// stats[common.UtilityStatTypeKey] =

	// switch eventType {
	// // case common.Event_MatchStartID:
	// // case common.Event_Economy:
	// // 	// return NewEconomyGameEvent
	// case common.Event_RoundEndID:
	// case common.Event_FragOrScoreID:
	// 	stats[common.BattleStatTypeKey] = battleContext.

	// // 	// default:
	// // 	// 	entities
	// // 	// TODO: gerar Stats default (Players, Round etc para todos os eventos com entidade)
	// // }

	return replay_entity.NewGameEvent(matchContext.MatchID, tickID, gameTime, eventType, payload, entities, stats, matchContext.ResourceOwner), nil
}

// func NewEconomyGameEvent(matchID uuid.UUID, ) {
// 	ID:            uuid.New(),
// 	MatchID:       matchContext.MatchID,
// 	Type:          common.Event_RoundEndID,
// 	Time:          p.CurrentTime(),
// 	Body:     event,
// 	Entities:
// 	ResourceOwner: matchContext.ResourceOwner,
// }
