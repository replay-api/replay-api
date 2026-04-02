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
	"github.com/replay-api/replay-api/pkg/infra/kafka"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// setMatchmakingContext creates a context with system-level admin privileges for matchmaking ingestion.
// This sets the AudienceKey to TenantAudienceIDKey so that IsAdmin() returns true,
// allowing the SubmitMatchResult RBAC check to pass for ScoreSourceMatchmaking.
func setMatchmakingContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.Nil)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	ctx = context.WithValue(ctx, shared.AudienceKey, shared.TenantAudienceIDKey)
	return ctx
}

// MatchmakingIngestionService ingests match results from the matchmaking Kafka topic
// and creates verified MatchResult records in the scores domain.
type MatchmakingIngestionService struct {
	repository     scores_out.MatchResultRepository
	commandHandler scores_in.MatchResultCommandHandler
}

// NewMatchmakingIngestionService creates a new matchmaking ingestion service
func NewMatchmakingIngestionService(
	repository scores_out.MatchResultRepository,
	commandHandler scores_in.MatchResultCommandHandler,
) *MatchmakingIngestionService {
	return &MatchmakingIngestionService{
		repository:     repository,
		commandHandler: commandHandler,
	}
}

// ProcessMatchEvent handles a matchmaking MatchEvent from Kafka.
// It converts the event into a SubmitMatchResultCommand and submits it through the scores domain.
// Matchmaking results are auto-verified as they come from a trusted system source.
func (s *MatchmakingIngestionService) ProcessMatchEvent(ctx context.Context, event *kafka.MatchEvent) error {
	if event == nil {
		return fmt.Errorf("nil match event")
	}

	if event.MatchID == uuid.Nil {
		return fmt.Errorf("match event missing match_id")
	}

	if event.Result == nil {
		slog.InfoContext(ctx, "ignoring match event without result (not completed)",
			slog.String("match_id", event.MatchID.String()),
			slog.String("event_type", event.EventType),
		)
		return nil // Not an error — match events without results (e.g., MATCH_CREATED) are skipped
	}

	slog.InfoContext(ctx, "processing matchmaking result for score ingestion",
		slog.String("match_id", event.MatchID.String()),
		slog.String("event_type", event.EventType),
		slog.String("game_type", event.GameType),
	)

	// Idempotency: check if a match result already exists for this match
	existing, _ := s.repository.FindByMatchID(ctx, event.MatchID)
	if existing != nil && !existing.Status.IsTerminal() {
		slog.InfoContext(ctx, "match result already exists, skipping",
			slog.String("match_id", event.MatchID.String()),
			slog.String("existing_status", string(existing.Status)),
		)
		return nil // Idempotent — already ingested
	}

	// Map game type
	gameID := replay_common.GameIDKey(event.GameType)
	if gameID == "" {
		gameID = replay_common.CS2.ID
	}

	// Build team results from event
	teamResults := s.buildTeamResults(event)
	if len(teamResults) < 2 {
		return fmt.Errorf("matchmaking event has fewer than 2 teams for match %s", event.MatchID)
	}

	// Build player results from event
	playerResults := s.buildPlayerResults(event)

	// Determine played time
	playedAt := time.Now().UTC()
	if event.Result.CompletedAt > 0 {
		playedAt = time.Unix(0, event.Result.CompletedAt*int64(time.Millisecond))
	} else if event.Timestamp > 0 {
		playedAt = time.Unix(0, event.Timestamp*int64(time.Millisecond))
	}

	// Duration
	duration := time.Duration(event.Result.Duration) * time.Second

	// Compute map name from metadata
	mapName := ""
	if event.Metadata != nil {
		mapName = event.Metadata["map_name"]
	}

	// Build a system context for the matchmaking service (system-level submission)
	systemCtx := setMatchmakingContext(ctx)

	// Build the lobby/session UUID for linking
	var sessionID *uuid.UUID
	if event.LobbyID != uuid.Nil {
		sessionID = &event.LobbyID
	}

	// Create command
	cmd := scores_in.SubmitMatchResultCommand{
		MatchID:              event.MatchID,
		MatchmakingSessionID: sessionID,
		GameID:               gameID,
		MapName:              mapName,
		Mode:                 "matchmaking",
		Source:               scores_vo.ScoreSourceMatchmaking,
		TeamResults:          teamResults,
		PlayerResults:        playerResults,
		PlayedAt:             playedAt,
		Duration:             duration,
	}

	result, err := s.commandHandler.SubmitMatchResult(systemCtx, cmd)
	if err != nil {
		return fmt.Errorf("failed to submit matchmaking result for match %s: %w", event.MatchID, err)
	}

	slog.InfoContext(ctx, "matchmaking result ingested successfully",
		slog.String("match_result_id", result.ID.String()),
		slog.String("match_id", event.MatchID.String()),
		slog.String("status", string(result.Status)),
		slog.Int("team_count", len(teamResults)),
		slog.Int("player_count", len(playerResults)),
	)

	return nil
}

// buildTeamResults converts Kafka TeamInfo + MatchResult.Scores into domain TeamResults
func (s *MatchmakingIngestionService) buildTeamResults(event *kafka.MatchEvent) []scores_entities.TeamResult {
	teamResults := make([]scores_entities.TeamResult, 0, len(event.Teams))

	for i, team := range event.Teams {
		score := 0
		if event.Result != nil && event.Result.Scores != nil {
			score = event.Result.Scores[team.TeamID.String()]
		}

		position := i + 1
		if event.Result != nil && event.Result.WinnerTeamID != nil && *event.Result.WinnerTeamID == team.TeamID {
			position = 1
		} else if position == 1 && event.Result != nil && event.Result.WinnerTeamID != nil {
			position = 2 // Shift non-winner if winner isn't first
		}

		teamResults = append(teamResults, scores_entities.TeamResult{
			TeamID:   team.TeamID,
			TeamName: team.Name,
			Score:    score,
			Position: position,
			Players:  team.PlayerIDs,
		})
	}

	return teamResults
}

// buildPlayerResults converts Kafka PlayerMatchStat into domain PlayerResults
func (s *MatchmakingIngestionService) buildPlayerResults(event *kafka.MatchEvent) []scores_entities.PlayerResult {
	if event.Result == nil || len(event.Result.PlayerStats) == 0 {
		return nil
	}

	// Build a player→team lookup
	playerTeamMap := make(map[uuid.UUID]uuid.UUID)
	for _, team := range event.Teams {
		for _, playerID := range team.PlayerIDs {
			playerTeamMap[playerID] = team.TeamID
		}
	}

	playerResults := make([]scores_entities.PlayerResult, 0, len(event.Result.PlayerStats))
	for _, ps := range event.Result.PlayerStats {
		teamID := playerTeamMap[ps.PlayerID]

		playerResults = append(playerResults, scores_entities.PlayerResult{
			PlayerID: ps.PlayerID,
			TeamID:   teamID,
			Score:    ps.Score,
			Kills:    ps.Kills,
			Deaths:   ps.Deaths,
			Assists:  ps.Assists,
			Stats: map[string]interface{}{
				"mmr_change": ps.MMRChange,
				"source":     "matchmaking",
			},
		})
	}

	return playerResults
}
