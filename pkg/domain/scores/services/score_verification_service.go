package scores_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// AutoFinalizationWindow is the time after verification during which results can be disputed
// before being automatically finalized. This mirrors the 72-hour escrow period in matchmaking.
const AutoFinalizationWindow = 72 * time.Hour

// ScoreVerificationService handles the automated verification and finalization pipeline
type ScoreVerificationService struct {
	repository              scores_out.MatchResultRepository
	eventPublisher          scores_out.ScoreEventPublisher
	prizeDistributionGateway scores_out.PrizeDistributionGateway
	commandHandler          scores_in.MatchResultCommandHandler
}

// NewScoreVerificationService creates a new score verification service
func NewScoreVerificationService(
	repository scores_out.MatchResultRepository,
	eventPublisher scores_out.ScoreEventPublisher,
	prizeDistributionGateway scores_out.PrizeDistributionGateway,
	commandHandler scores_in.MatchResultCommandHandler,
) *ScoreVerificationService {
	return &ScoreVerificationService{
		repository:              repository,
		eventPublisher:          eventPublisher,
		prizeDistributionGateway: prizeDistributionGateway,
		commandHandler:          commandHandler,
	}
}

// ProcessReplayCompleted handles the replays.completed Kafka event
// and creates a verified match result from the parsed replay data
func (s *ScoreVerificationService) ProcessReplayCompleted(
	ctx context.Context,
	matchID uuid.UUID,
	replayID uuid.UUID,
	gameID replay_common.GameIDKey,
	mapName string,
	mode string,
	teamResults []scores_entities.TeamResult,
	playerResults []scores_entities.PlayerResult,
	playedAt time.Time,
	duration time.Duration,
	roundsPlayed int,
	matchmakingSessionID *uuid.UUID,
	tournamentID *uuid.UUID,
) error {
	slog.InfoContext(ctx, "processing completed replay for score submission",
		slog.String("match_id", matchID.String()),
		slog.String("replay_id", replayID.String()),
	)

	cmd := scores_in.SubmitReplayResultCommand{
		MatchID:              matchID,
		ReplayID:             replayID,
		GameID:               gameID,
		MapName:              mapName,
		Mode:                 mode,
		TeamResults:          teamResults,
		PlayerResults:        playerResults,
		PlayedAt:             playedAt,
		Duration:             duration,
		RoundsPlayed:         roundsPlayed,
		MatchmakingSessionID: matchmakingSessionID,
		TournamentID:         tournamentID,
	}

	result, err := s.commandHandler.SubmitMatchResultFromReplay(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to submit replay result: %w", err)
	}

	slog.InfoContext(ctx, "replay result submitted and auto-verified",
		slog.String("match_result_id", result.ID.String()),
		slog.String("status", string(result.Status)),
	)

	return nil
}

// ProcessAutoFinalization checks for verified/conciliated results that have passed
// the dispute window and automatically finalizes them. This should be called
// periodically (e.g., every 15 minutes) by a scheduled worker.
func (s *ScoreVerificationService) ProcessAutoFinalization(ctx context.Context) (int, error) {
	// Find verified results eligible for auto-finalization
	verifiedResults, _, err := s.repository.FindByStatus(ctx, scores_vo.ResultStatusVerified, 100, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to find verified results: %w", err)
	}

	// Also check conciliated results
	conciliatedResults, _, err := s.repository.FindByStatus(ctx, scores_vo.ResultStatusConciliated, 100, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to find conciliated results: %w", err)
	}

	candidates := append(verifiedResults, conciliatedResults...)
	finalizedCount := 0

	for _, result := range candidates {
		// Check if the dispute window has passed
		var referenceTime time.Time
		if result.ConciliatedAt != nil {
			referenceTime = *result.ConciliatedAt
		} else if result.VerifiedAt != nil {
			referenceTime = *result.VerifiedAt
		} else {
			continue
		}

		if time.Since(referenceTime) < AutoFinalizationWindow {
			continue
		}

		// Auto-finalize
		slog.InfoContext(ctx, "auto-finalizing match result",
			slog.String("match_result_id", result.ID.String()),
			slog.String("verified_at", referenceTime.String()),
		)

		cmd := scores_in.FinalizeMatchResultCommand{
			MatchResultID: result.ID,
		}
		if err := s.commandHandler.FinalizeMatchResult(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to auto-finalize match result",
				slog.String("match_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		finalizedCount++
	}

	if finalizedCount > 0 {
		slog.InfoContext(ctx, "auto-finalization batch completed",
			slog.Int("finalized_count", finalizedCount),
			slog.Int("candidates_checked", len(candidates)),
		)
	}

	return finalizedCount, nil
}

// ValidateScoreConsistency performs cross-validation checks on submitted scores
func (s *ScoreVerificationService) ValidateScoreConsistency(result *scores_entities.MatchResult) []string {
	var warnings []string

	// Check if total player kills match reasonable bounds
	totalKills := 0
	totalDeaths := 0
	for _, pr := range result.PlayerResults {
		totalKills += pr.Kills
		totalDeaths += pr.Deaths
	}

	// In CS2, kills should roughly equal deaths (except for suicide/fall damage)
	if totalKills > 0 && totalDeaths > 0 {
		ratio := float64(totalKills) / float64(totalDeaths)
		if ratio < 0.8 || ratio > 1.2 {
			warnings = append(warnings, fmt.Sprintf("kill/death ratio is unusual: %.2f (expected ~1.0)", ratio))
		}
	}

	// Check if team scores are reasonable for the game mode
	for _, tr := range result.TeamResults {
		if result.Mode == "competitive" && tr.Score > 32 {
			warnings = append(warnings, fmt.Sprintf("team %s score %d seems too high for competitive mode", tr.TeamName, tr.Score))
		}
	}

	// Check for duplicate player IDs across teams
	playerTeams := make(map[uuid.UUID]uuid.UUID)
	for _, pr := range result.PlayerResults {
		if existingTeam, exists := playerTeams[pr.PlayerID]; exists {
			warnings = append(warnings, fmt.Sprintf("player %s appears in multiple teams: %s and %s", pr.PlayerID, existingTeam, pr.TeamID))
		}
		playerTeams[pr.PlayerID] = pr.TeamID
	}

	return warnings
}
