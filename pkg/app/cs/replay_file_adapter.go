package cs2

import (
	"context"
	"io"
	"log/slog"

	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	handlers "github.com/replay-api/replay-api/pkg/app/cs/handlers"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	e "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type CS2ReplayAdapter struct {
}

func NewCS2ReplayAdapter() *CS2ReplayAdapter {
	return &CS2ReplayAdapter{}
}

func registerParsers(p dem.Parser, matchContext *state.CS2MatchContext, eventsChan chan *e.GameEvent) {
	p.RegisterEventHandler(handlers.BeginNewMatch(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.WeaponFire(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.HitEvent(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.RoundMVP(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ClutchStart(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ClutchProgress(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ClutchEnd(p, matchContext, eventsChan))
	// p.RegisterEventHandler(handlers.EconomyEvent(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.GenericGameEvent(p, matchContext, eventsChan))
}

func (c *CS2ReplayAdapter) Parse(ctx context.Context, matchID uuid.UUID, content io.Reader, eventsChan chan *e.GameEvent) error {
	matchContext := state.NewCS2MatchContext(ctx, matchID)
	parser := dem.NewParser(content)
	slog.Info("Parsing demo file at %s", "CS2ReplayAdapter.GetEvents", matchID)
	defer parser.Close()
	defer close(eventsChan)

	registerParsers(parser, matchContext, eventsChan)

	err := parser.ParseToEnd()

	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse demo: %v", "err", err)
		return err
	}

	return nil
}
