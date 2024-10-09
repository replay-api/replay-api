package handlers

import (
	"log/slog"

	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	state "github.com/psavelis/team-pro/replay-api/pkg/app/cs/state"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	"github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

func RoundMVP(p dem.Parser, matchContext *state.CS2MatchContext, out chan entities.GameEvent) func(e evt.RoundMVPAnnouncement) {
	return func(event evt.RoundMVPAnnouncement) {
		slog.Info("RoundMVP event: %v", "event", event)

		out <- entities.GameEvent{
			ID:            uuid.New(),
			MatchID:       matchContext.MatchID,
			Type:          common.Event_RoundMVPAnnouncementID,
			GameTime:      p.CurrentTime(),
			Payload:       event,
			ResourceOwner: matchContext.ResourceOwner,
		}
	}
}

// ID        uuid.UUID     `json:"id"`
// 	MatchID   string        `json:"match_id"`
// 	Type      string        `json:"type"`
// 	Time      time.Duration `json:"event_time"`
// 	EventData interface{}   `json:"event_data"`
