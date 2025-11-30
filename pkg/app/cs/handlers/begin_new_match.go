package handlers

import (
	"log/slog"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/replay-api/replay-api/pkg/app/cs/builders"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	common "github.com/replay-api/replay-api/pkg/domain"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

func BeginNewMatch(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.MatchStart) {
	return func(event evt.MatchStart) {
		h := p.Header()

		gs := p.GameState()

		matchContext = matchContext.WithRound(0, gs)

		matchContext.SetHeader(cs_entity.CSReplayFileHeader{
			Filestamp:       h.Filestamp,
			Protocol:        h.Protocol,
			NetworkProtocol: h.NetworkProtocol,
			ServerName:      h.ServerName,
			ClientName:      h.ClientName,
			MapName:         h.MapName,
			Length:          h.PlaybackTime,
			Ticks:           h.PlaybackTicks,
			Frames:          h.PlaybackFrames,
		})

		b := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts)

		payload := b.Build()

		currentTick := common.TickIDType(gs.IngameTick())

		gameEvent, err := event_factory.NewGameEvent(
			common.Event_MatchStartID,
			matchContext,
			0,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		slog.Info("MatchStart:", "gameEvent", common.Event_MatchStartID)

		if err != nil {
			slog.Error("unable to create new match event")
			return
		}

		out <- gameEvent
	}
}
