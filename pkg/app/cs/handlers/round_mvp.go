package handlers

import (
	"fmt"
	"log/slog"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/replay-api/replay-api/pkg/app/cs/builders"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	common "github.com/replay-api/replay-api/pkg/domain"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type RoundMVPPayload struct {
	NetworkPlayerID string
	Name            string
	Reason          string
	ClanName        string
	PlayerStats     *cs_entity.CSPlayerStats
}

func RoundMVP(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.RoundMVPAnnouncement) {
	return func(event evt.RoundMVPAnnouncement) {
		gs := p.GameState()

		// slog.Info("RoundMVP event: %v", "event", event)

		// stats := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts).Build()

		// ID            common.PlayerIDType `json:"id" bson:"_id"`
		// GameID        common.GameIDKey    `json:"game_id" bson:"game_id"`
		// UserID        *uuid.UUID          `json:"-" bson:"user_id"`
		// NetworkUserID string              `json:"-" bson:"network_user_id"`
		// NetworkID     common.NetworkIDKey `json:"network_id" bson:"network_id"`
		// Name          string              `json:"name" bson:"name"`
		// NameHistory   []string            `json:"-" bson:"name_history"`
		// ClanName      string              `json:"clan_name" bson:"clan_name"`
		// AvatarURI     string              `json:"avatar_uri" bson:"avatar_uri"`

		// NetworkClanID string     `json:"network_clan_id" bson:"network_clan_id"`
		// VerifiedAt    *time.Time `json:"verified_at" bson:"verified_at"`

		// ResourceOwner common.ResourceOwner `json:"-" bson:"resource_owner"`
		// ShareTokens   []ShareToken         `json:"-" bson:"share_tokens"`
		// CreatedAt     time.Time            `json:"-" bson:"created_at"`
		// UpdatedAt     *time.Time           `json:"-" bson:"updated_at"`

		mvp := &RoundMVPPayload{
			NetworkPlayerID: fmt.Sprintf("%d", event.Player.SteamID64),
			Name:            event.Player.Name,
			ClanName:        event.Player.ClanTag(),
		}
		roundIndex := gs.TotalRoundsPlayed()
		stats := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts).StatsFromPlayerWithRound(roundIndex+1, event.Player)

		switch event.Reason {
		case evt.MVPReasonMostEliminations:
			mvp.Reason = "Most Eliminations"
		case evt.MVPReasonBombDefused:
			mvp.Reason = "Defused the bomb"
		case evt.MVPReasonBombPlanted:
			mvp.Reason = "Planted the bomb"
		}

		mvp.PlayerStats = stats

		currentTick := common.TickIDType(gs.IngameTick())

		gameEvent, err := event_factory.NewGameEvent(
			common.Event_RoundMVPAnnouncementID,
			matchContext,
			0,
			currentTick,
			p.CurrentTime(),
			mvp,
		)

		if err != nil {
			slog.Error("unable to create new round mvp announcement event")
			return
		}

		out <- gameEvent
	}
}

// ID        uuid.UUID     `json:"id"`
// 	MatchID   string        `json:"match_id"`
// 	Type      string        `json:"type"`
// 	Time      time.Duration `json:"event_time"`
// 	EventData interface{}   `json:"event_data"`
