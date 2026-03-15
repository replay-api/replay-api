package kafka

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
)

// TournamentEventPublisherAdapter adapts the Kafka EventPublisher to the tournament domain port
type TournamentEventPublisherAdapter struct {
	publisher *EventPublisher
}

// Compile-time interface satisfaction check
var _ tournament_out.TournamentEventPublisher = (*TournamentEventPublisherAdapter)(nil)

// NewTournamentEventPublisherAdapter creates a new adapter
func NewTournamentEventPublisherAdapter(publisher *EventPublisher) *TournamentEventPublisherAdapter {
	return &TournamentEventPublisherAdapter{publisher: publisher}
}

// TournamentEvent represents a Kafka tournament event payload
type TournamentEvent struct {
	TournamentID   uuid.UUID `json:"tournament_id"`
	Name           string    `json:"name"`
	GameID         string    `json:"game_id"`
	Format         string    `json:"format"`
	Status         string    `json:"status"`
	OrganizerID    uuid.UUID `json:"organizer_id"`
	Participants   int       `json:"participants"`
	EventType      string    `json:"event_type"`
	PlayerID       string    `json:"player_id,omitempty"`
	MatchID        string    `json:"match_id,omitempty"`
	WinnerID       string    `json:"winner_id,omitempty"`
	Round          int       `json:"round,omitempty"`
	NewMatchCount  int       `json:"new_match_count,omitempty"`
}

func toTournamentEvent(t *tournament_entities.Tournament, eventType string) *TournamentEvent {
	return &TournamentEvent{
		TournamentID: t.ID,
		Name:         t.Name,
		GameID:       string(t.GameID),
		Format:       string(t.Format),
		Status:       string(t.Status),
		OrganizerID:  t.OrganizerID,
		Participants: len(t.Participants),
		EventType:    eventType,
	}
}

func (a *TournamentEventPublisherAdapter) publishEvent(ctx context.Context, event *TournamentEvent) error {
	if a.publisher == nil {
		slog.WarnContext(ctx, "tournament event publisher not configured, skipping event", "event_type", event.EventType)
		return nil
	}
	return a.publisher.PublishTournamentEvent(ctx, event.TournamentID.String(), event)
}

func (a *TournamentEventPublisherAdapter) PublishTournamentCreated(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return a.publishEvent(ctx, toTournamentEvent(tournament, "tournament.created"))
}

func (a *TournamentEventPublisherAdapter) PublishRegistrationOpened(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return a.publishEvent(ctx, toTournamentEvent(tournament, "tournament.registration.opened"))
}

func (a *TournamentEventPublisherAdapter) PublishRegistrationClosed(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return a.publishEvent(ctx, toTournamentEvent(tournament, "tournament.registration.closed"))
}

func (a *TournamentEventPublisherAdapter) PublishTournamentStarted(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return a.publishEvent(ctx, toTournamentEvent(tournament, "tournament.started"))
}

func (a *TournamentEventPublisherAdapter) PublishTournamentCompleted(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return a.publishEvent(ctx, toTournamentEvent(tournament, "tournament.completed"))
}

func (a *TournamentEventPublisherAdapter) PublishTournamentCancelled(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return a.publishEvent(ctx, toTournamentEvent(tournament, "tournament.cancelled"))
}

func (a *TournamentEventPublisherAdapter) PublishPlayerRegistered(ctx context.Context, tournament *tournament_entities.Tournament, playerID uuid.UUID) error {
	event := toTournamentEvent(tournament, "tournament.player.registered")
	event.PlayerID = playerID.String()
	return a.publishEvent(ctx, event)
}

func (a *TournamentEventPublisherAdapter) PublishPlayerUnregistered(ctx context.Context, tournament *tournament_entities.Tournament, playerID uuid.UUID) error {
	event := toTournamentEvent(tournament, "tournament.player.unregistered")
	event.PlayerID = playerID.String()
	return a.publishEvent(ctx, event)
}

func (a *TournamentEventPublisherAdapter) PublishMatchResultRecorded(ctx context.Context, tournament *tournament_entities.Tournament, matchID uuid.UUID, winnerID uuid.UUID) error {
	event := toTournamentEvent(tournament, "tournament.match.result_recorded")
	event.MatchID = matchID.String()
	event.WinnerID = winnerID.String()
	return a.publishEvent(ctx, event)
}

func (a *TournamentEventPublisherAdapter) PublishBracketAdvanced(ctx context.Context, tournament *tournament_entities.Tournament, newMatches []tournament_entities.TournamentMatch) error {
	event := toTournamentEvent(tournament, "tournament.bracket.advanced")
	if len(newMatches) > 0 {
		event.Round = newMatches[0].Round
	}
	event.NewMatchCount = len(newMatches)
	return a.publishEvent(ctx, event)
}
