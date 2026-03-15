package oracle_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_metadata "github.com/replay-api/replay-api/pkg/domain/replay/services/metadata"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type gameImportCommandHandler struct {
	oracleCommandHandler    oracle_in.OracleCommandHandler
	oracleResultRepo        oracle_out.OracleResultRepository
	streamConfigRepo        oracle_out.OCRStreamConfigRepository
	reconciliationService   *replay_metadata.MatchReconciliationService
	matchResultRepo         scores_out.MatchResultRepository
	eventPublisher          oracle_out.OracleEventPublisher
}

// NewGameImportCommandHandler creates a new game import command handler.
func NewGameImportCommandHandler(
	oracleCommandHandler oracle_in.OracleCommandHandler,
	oracleResultRepo oracle_out.OracleResultRepository,
	streamConfigRepo oracle_out.OCRStreamConfigRepository,
	reconciliationService *replay_metadata.MatchReconciliationService,
	matchResultRepo scores_out.MatchResultRepository,
	eventPublisher oracle_out.OracleEventPublisher,
) oracle_in.GameImportCommandHandler {
	return &gameImportCommandHandler{
		oracleCommandHandler:  oracleCommandHandler,
		oracleResultRepo:      oracleResultRepo,
		streamConfigRepo:      streamConfigRepo,
		reconciliationService: reconciliationService,
		matchResultRepo:       matchResultRepo,
		eventPublisher:        eventPublisher,
	}
}

// ImportDiscoveredMatch handles the full import of a discovered external match:
// 1. Creates a Match entity (if not existing)
// 2. Creates an OracleResult for the external match
// 3. Triggers score ingestion from the provider
// 4. Optionally creates VOD OCR configs
func (h *gameImportCommandHandler) ImportDiscoveredMatch(ctx context.Context, cmd oracle_in.ImportDiscoveredMatchCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	ext := cmd.ExternalMatch

	slog.InfoContext(ctx, "importing discovered match",
		slog.String("external_match_id", ext.ExternalMatchID),
		slog.String("game_id", string(ext.GameID)),
		slog.String("provider", string(ext.Provider)),
		slog.String("teams", fmt.Sprintf("%s vs %s", ext.TeamAName, ext.TeamBName)),
	)

	// Check if oracle result already exists (dedup)
	existing, err := h.oracleResultRepo.FindByExternalMatchID(ctx, ext.ExternalMatchID)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to check existing oracle result: %w", err)
	}
	if existing != nil {
		slog.InfoContext(ctx, "oracle result already exists, skipping import",
			slog.String("external_match_id", ext.ExternalMatchID),
			slog.String("oracle_result_id", existing.GetID().String()),
		)
		return nil
	}

	// 1. Create the Match entity on the platform (enriched with OCR data)
	// Use the resource owner from context (set by middleware or worker)
	systemOwner := shared.GetResourceOwner(ctx)

	// Map oracle source type → match source
	matchSource := replay_entity.MatchSourceFromOracleSource(string(ext.Provider))

	// Determine if this is a YouTube VOD import
	if len(ext.VODURLs) > 0 || ext.StreamURL != "" {
		if ext.Provider.IsOCR() {
			matchSource = replay_entity.MatchSourceOCRStream
		}
	}

	// Delegate ALL reconciliation to the dedicated service (single source of truth)
	reconInput := replay_metadata.ReconciliationInput{
		GameID:          ext.GameID,
		Source:          matchSource,
		ExternalMatchID: ext.ExternalMatchID,
		Provider:        string(ext.Provider),
		TeamAName:       ext.TeamAName,
		TeamBName:       ext.TeamBName,
		TeamAScore:      ext.TeamAScore,
		TeamBScore:      ext.TeamBScore,
		MapName:         ext.MapName,
		PlayedAt:        ext.PlayedAt,
		GameNumber:      0,  // Discovery doesn't know game number yet
		SeriesType:      "", // Discovery doesn't know series type yet
		Confidence:      1.0,
		ResourceOwner:   systemOwner,
	}

	reconResult, err := h.reconciliationService.ReconcileMatchFull(ctx, reconInput)
	if err != nil {
		slog.ErrorContext(ctx, "match reconciliation failed",
			slog.String("external_match_id", ext.ExternalMatchID),
			slog.String("error", err.Error()),
		)
		// Continue to create oracle result even if match reconciliation fails
	}

	var match *replay_entity.Match
	if reconResult != nil {
		match = reconResult.Match
		slog.InfoContext(ctx, "match reconciliation completed",
			slog.String("outcome", string(reconResult.Outcome)),
			slog.String("slug", reconResult.SlugUsed),
			slog.Bool("conflict", reconResult.ConflictDetected),
			slog.String("match_id", match.GetID().String()),
		)
	}

	// 2. Create oracle result and trigger ingestion
	oracleCmd := oracle_in.CreateExternalMatchOracleCommand{
		ExternalMatchID: ext.ExternalMatchID,
		GameID:          ext.GameID,
	}

	oracleResult, err := h.oracleCommandHandler.CreateExternalMatchOracle(ctx, oracleCmd)
	if err != nil {
		return fmt.Errorf("failed to create oracle result: %w", err)
	}

	// 3. Ingest the score from the discovery data (direct provider submission)
	if cmd.TriggerAPIIngest {
		ingestCmd := oracle_in.IngestExternalScoreCommand{
			OracleResultID:  gameImportPtrUUID(oracleResult.GetID()),
			ExternalMatchID: &ext.ExternalMatchID,
			GameID:          ext.GameID,
			SourceType:      ext.Provider,
			ProviderMatchID: ext.ExternalMatchID,
			WinnerTeamID:    ext.WinnerTeamID,
			IsDraw:          ext.IsDraw,
			TeamAID:         ext.TeamAID,
			TeamBID:         ext.TeamBID,
			TeamAScore:      ext.TeamAScore,
			TeamBScore:      ext.TeamBScore,
			RoundsPlayed:    0,
		}

		if err := h.oracleCommandHandler.IngestExternalScore(ctx, ingestCmd); err != nil {
			slog.WarnContext(ctx, "failed to ingest initial score from discovery",
				slog.String("external_match_id", ext.ExternalMatchID),
				slog.String("error", err.Error()),
			)
		}
	}

	// 4. Also trigger ingestion from ALL registered providers
	triggerCmd := oracle_in.TriggerIngestionCommand{
		ExternalMatchID: &ext.ExternalMatchID,
		GameID:          ext.GameID,
	}
	if err := h.oracleCommandHandler.TriggerIngestionFromAllProviders(ctx, triggerCmd); err != nil {
		slog.WarnContext(ctx, "failed to trigger multi-provider ingestion",
			slog.String("external_match_id", ext.ExternalMatchID),
			slog.String("error", err.Error()),
		)
	}

	// 5. Create OCR stream config if VODs are available
	if cmd.TriggerOCR && len(ext.VODURLs) > 0 {
		for _, vodURL := range ext.VODURLs {
			streamConfig := oracle_entities.NewOCRStreamConfig(vodURL, ext.GameID, ext.ExternalMatchID)
			streamConfig.TeamAHint = ext.TeamAName
			streamConfig.TeamBHint = ext.TeamBName
			streamConfig.OracleResultID = gameImportPtrUUID(oracleResult.GetID())

			// CS2 HUD auto-crop at 720p
			if ext.GameID == "cs2" || ext.GameID == "csgo" {
				streamConfig.ScoreboardRegion = &oracle_entities.ScoreboardRegion{
					X: 350, Y: 0, Width: 530, Height: 80,
				}
			}

			if err := h.streamConfigRepo.Save(ctx, streamConfig); err != nil {
				slog.WarnContext(ctx, "failed to save OCR stream config",
					slog.String("vod_url", vodURL),
					slog.String("error", err.Error()),
				)
			}
		}
	} else if cmd.TriggerOCR && ext.StreamURL != "" {
		streamConfig := oracle_entities.NewOCRStreamConfig(ext.StreamURL, ext.GameID, ext.ExternalMatchID)
		streamConfig.TeamAHint = ext.TeamAName
		streamConfig.TeamBHint = ext.TeamBName
		streamConfig.OracleResultID = gameImportPtrUUID(oracleResult.GetID())

		if ext.GameID == "cs2" || ext.GameID == "csgo" {
			streamConfig.ScoreboardRegion = &oracle_entities.ScoreboardRegion{
				X: 350, Y: 0, Width: 530, Height: 80,
			}
		}

		if err := h.streamConfigRepo.Save(ctx, streamConfig); err != nil {
			slog.WarnContext(ctx, "failed to save OCR stream config",
				slog.String("stream_url", ext.StreamURL),
				slog.String("error", err.Error()),
			)
		}
	}

	slog.InfoContext(ctx, "match import completed",
		slog.String("external_match_id", ext.ExternalMatchID),
		slog.String("oracle_result_id", oracleResult.GetID().String()),
		slog.String("match_id", match.GetID().String()),
	)

	return nil
}

// ImportFromYouTubeVOD creates an OCR stream config for a YouTube VOD and queues it for processing.
func (h *gameImportCommandHandler) ImportFromYouTubeVOD(ctx context.Context, cmd oracle_in.ImportFromYouTubeVODCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Check if stream config already exists for this external match ID
	existing, err := h.streamConfigRepo.FindByExternalMatchID(ctx, cmd.ExternalMatchID)
	if err != nil {
		return fmt.Errorf("failed to check existing stream config: %w", err)
	}
	if existing != nil {
		slog.InfoContext(ctx, "OCR stream config already exists",
			slog.String("external_match_id", cmd.ExternalMatchID),
		)
		return nil
	}

	streamConfig := oracle_entities.NewOCRStreamConfig(cmd.VideoURL, cmd.GameID, cmd.ExternalMatchID)
	streamConfig.TeamAHint = cmd.TeamAHint
	streamConfig.TeamBHint = cmd.TeamBHint

	if cmd.ScoreboardRegion != nil {
		streamConfig.ScoreboardRegion = cmd.ScoreboardRegion
	} else if cmd.GameID == "cs2" || cmd.GameID == "csgo" {
		// Default CS2 HUD scoreboard region at 720p
		streamConfig.ScoreboardRegion = &oracle_entities.ScoreboardRegion{
			X: 350, Y: 0, Width: 530, Height: 80,
		}
	}

	if err := h.streamConfigRepo.Save(ctx, streamConfig); err != nil {
		return fmt.Errorf("failed to save OCR stream config: %w", err)
	}

	slog.InfoContext(ctx, "YouTube VOD import queued",
		slog.String("video_url", cmd.VideoURL),
		slog.String("external_match_id", cmd.ExternalMatchID),
		slog.String("game_id", string(cmd.GameID)),
	)

	return nil
}

// BridgeOracleToMatchResult converts a finalized OracleResult consensus into a MatchResult.
// This is the final step: oracle consensus → platform score record.
func (h *gameImportCommandHandler) BridgeOracleToMatchResult(ctx context.Context, cmd oracle_in.BridgeOracleToMatchResultCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	oracleResult, err := h.oracleResultRepo.FindByID(ctx, cmd.OracleResultID)
	if err != nil {
		return fmt.Errorf("failed to find oracle result: %w", err)
	}
	if oracleResult == nil {
		return fmt.Errorf("oracle result not found: %s", cmd.OracleResultID)
	}

	// Oracle must have reached consensus
	if oracleResult.Status != oracle_vo.OracleStatusConsensusReached &&
		oracleResult.Status != oracle_vo.OracleStatusPublished &&
		oracleResult.Status != oracle_vo.OracleStatusFinalized {
		return fmt.Errorf("oracle result not in consensus/published/finalized state: %s", oracleResult.Status)
	}

	if oracleResult.ConsensusResult == nil {
		return fmt.Errorf("oracle result has no consensus outcome")
	}

	consensus := oracleResult.ConsensusResult

	// Build TeamResults from consensus
	teamResults := make([]scores_entities.TeamResult, 0, len(consensus.TeamScores))
	for i, ts := range consensus.TeamScores {
		tr := scores_entities.TeamResult{
			TeamID:   ts.TeamID,
			Score:    ts.Score,
			Position: i + 1,
		}
		teamResults = append(teamResults, tr)
	}

	// Determine match ID
	var matchID uuid.UUID
	if cmd.MatchID != nil {
		matchID = *cmd.MatchID
	} else if oracleResult.MatchID != nil {
		matchID = *oracleResult.MatchID
	} else {
		matchID = uuid.New()
	}

	// Determine map name from game outcomes
	mapName := ""
	if len(consensus.GameOutcomes) > 0 {
		mapName = consensus.GameOutcomes[0].MapName
	}

	// Use the resource owner from context
	systemOwner := shared.GetResourceOwner(ctx)

	matchResult, err := scores_entities.NewMatchResultFromOracle(
		systemOwner,
		matchID,
		oracleResult.GameID,
		mapName,
		consensus.SeriesFormat,
		teamResults,
		nil, // No player-level results from consensus (yet)
		oracleResult.CreatedAt,
		0, // Duration not tracked by oracle
	)
	if err != nil {
		return fmt.Errorf("failed to create match result from oracle: %w", err)
	}

	// Auto-verify oracle-sourced results with high confidence
	if consensus.ConfidenceLevel >= 80 {
		if err := matchResult.AutoVerify(); err != nil {
			slog.WarnContext(ctx, "failed to auto-verify match result",
				slog.String("oracle_result_id", cmd.OracleResultID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	if err := h.matchResultRepo.Save(ctx, matchResult); err != nil {
		return fmt.Errorf("failed to save match result: %w", err)
	}

	slog.InfoContext(ctx, "oracle result bridged to match result",
		slog.String("oracle_result_id", cmd.OracleResultID.String()),
		slog.String("match_id", matchID.String()),
		slog.String("match_result_id", matchResult.GetID().String()),
		slog.Int("confidence_level", consensus.ConfidenceLevel),
		slog.Int("team_count", len(teamResults)),
	)

	return nil
}

func gameImportPtrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
