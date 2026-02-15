package kafka

import (
	"context"

	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
)

// ScoreEventPublisherAdapter adapts the Kafka EventPublisher to the scores domain port
type ScoreEventPublisherAdapter struct {
	publisher *EventPublisher
}

// Compile-time interface satisfaction check
var _ scores_out.ScoreEventPublisher = (*ScoreEventPublisherAdapter)(nil)

// NewScoreEventPublisherAdapter creates a new adapter
func NewScoreEventPublisherAdapter(publisher *EventPublisher) *ScoreEventPublisherAdapter {
	return &ScoreEventPublisherAdapter{publisher: publisher}
}

func (a *ScoreEventPublisherAdapter) toScoreEvent(result *scores_entities.MatchResult) *ScoreEvent {
	teamScores := make([]TeamScoreInfo, len(result.TeamResults))
	for i, tr := range result.TeamResults {
		teamScores[i] = TeamScoreInfo{
			TeamID:   tr.TeamID,
			TeamName: tr.TeamName,
			Score:    tr.Score,
			Position: tr.Position,
		}
	}

	return &ScoreEvent{
		MatchResultID:        result.ID,
		MatchID:              result.MatchID,
		TournamentID:         result.TournamentID,
		MatchmakingSessionID: result.MatchmakingSessionID,
		GameID:               string(result.GameID),
		Source:               string(result.Source),
		Status:               string(result.Status),
		WinnerTeamID:         result.WinnerTeamID,
		IsDraw:               result.IsDraw,
		TeamScores:           teamScores,
		DisputeReason:        result.DisputeReason,
		PrizeDistributionID:  result.PrizeDistributionID,
	}
}

func (a *ScoreEventPublisherAdapter) PublishMatchResultSubmitted(ctx context.Context, result *scores_entities.MatchResult) error {
	return a.publisher.PublishScoreSubmitted(ctx, a.toScoreEvent(result))
}

func (a *ScoreEventPublisherAdapter) PublishMatchResultVerified(ctx context.Context, result *scores_entities.MatchResult) error {
	return a.publisher.PublishScoreVerified(ctx, a.toScoreEvent(result))
}

func (a *ScoreEventPublisherAdapter) PublishMatchResultDisputed(ctx context.Context, result *scores_entities.MatchResult) error {
	event := a.toScoreEvent(result)
	event.DisputeReason = result.DisputeReason
	return a.publisher.PublishScoreDisputed(ctx, event)
}

func (a *ScoreEventPublisherAdapter) PublishMatchResultConciliated(ctx context.Context, result *scores_entities.MatchResult) error {
	return a.publisher.PublishScoreConciliated(ctx, a.toScoreEvent(result))
}

func (a *ScoreEventPublisherAdapter) PublishMatchResultFinalized(ctx context.Context, result *scores_entities.MatchResult) error {
	return a.publisher.PublishScoreFinalized(ctx, a.toScoreEvent(result))
}

func (a *ScoreEventPublisherAdapter) PublishMatchResultCancelled(ctx context.Context, result *scores_entities.MatchResult) error {
	return a.publisher.PublishScoreCancelled(ctx, a.toScoreEvent(result))
}
