package oracle_usecases

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_services "github.com/replay-api/replay-api/pkg/domain/oracle/services"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// DisputeWindowDuration is the time after publication during which
// a score can be disputed before finalization
const DisputeWindowDuration = 72 * time.Hour

// oracleCommandHandler implements oracle_in.OracleCommandHandler
type oracleCommandHandler struct {
	repository      oracle_out.OracleResultRepository
	eventPublisher  oracle_out.OracleEventPublisher
	providers       []oracle_out.ExternalScorePort
	consensusEngine *oracle_services.ConsensusEngine
	chainGateway    oracle_out.ChainScoreGateway
	policy          oracle_vo.ConsensusPolicy
}

// NewOracleCommandHandler creates the command handler with all dependencies
func NewOracleCommandHandler(
	repository oracle_out.OracleResultRepository,
	eventPublisher oracle_out.OracleEventPublisher,
	providers []oracle_out.ExternalScorePort,
	consensusEngine *oracle_services.ConsensusEngine,
	chainGateway oracle_out.ChainScoreGateway,
	policy oracle_vo.ConsensusPolicy,
) oracle_in.OracleCommandHandler {
	return &oracleCommandHandler{
		repository:      repository,
		eventPublisher:  eventPublisher,
		providers:       providers,
		consensusEngine: consensusEngine,
		chainGateway:    chainGateway,
		policy:          policy,
	}
}

// IngestExternalScore ingests a score submission from an external provider
func (h *oracleCommandHandler) IngestExternalScore(ctx context.Context, cmd oracle_in.IngestExternalScoreCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Find or create oracle result
	result, err := h.findOrCreateOracleResult(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to find/create oracle result: %w", err)
	}

	// Build submission
	rawData, _ := json.Marshal(cmd)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256(rawData))

	submission := oracle_entities.ScoreSubmission{
		SourceType:      cmd.SourceType,
		ProviderMatchID: cmd.ProviderMatchID,
		WinnerTeamID:    cmd.WinnerTeamID,
		IsDraw:          cmd.IsDraw,
		TeamAID:         cmd.TeamAID,
		TeamBID:         cmd.TeamBID,
		TeamAScore:      cmd.TeamAScore,
		TeamBScore:      cmd.TeamBScore,
		RoundsPlayed:    cmd.RoundsPlayed,
		MVPPlayerID:     cmd.MVPPlayerID,
		GameDetails:     cmd.GameDetails,
		PlayerScores:    cmd.PlayerScores,
		RawResponse:     rawData,
		SourceHash:      sourceHash,
	}

	// Add submission to oracle result
	if err := result.AddSubmission(submission); err != nil {
		return fmt.Errorf("failed to add submission: %w", err)
	}

	slog.InfoContext(ctx, "external score ingested",
		slog.String("oracle_result_id", result.ID.String()),
		slog.String("source_type", string(cmd.SourceType)),
		slog.String("provider_match_id", cmd.ProviderMatchID),
		slog.Int("submission_count", result.GetSubmissionCount()),
	)

	// Publish ingestion event
	if h.eventPublisher != nil {
		latestSub := result.Submissions[len(result.Submissions)-1]
		if err := h.eventPublisher.PublishExternalIngested(ctx, result, latestSub); err != nil {
			slog.ErrorContext(ctx, "failed to publish external ingested event",
				slog.String("error", err.Error()),
			)
		}
	}

	// Check if ready for consensus
	if result.IsReadyForConsensus(h.policy.MinSources) && result.Status == oracle_vo.OracleStatusPending {
		if err := h.evaluateAndSetConsensus(ctx, result); err != nil {
			slog.WarnContext(ctx, "consensus evaluation failed, will retry",
				slog.String("oracle_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
			// Don't fail the ingestion — consensus can be retried
		}
	}

	// Persist
	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update oracle result: %w", err)
	}

	return nil
}

// CreateExternalMatchOracle creates an oracle result for an external-only match
func (h *oracleCommandHandler) CreateExternalMatchOracle(ctx context.Context, cmd oracle_in.CreateExternalMatchOracleCommand) (*oracle_entities.OracleResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Check for existing oracle result
	existing, _ := h.repository.FindByExternalMatchID(ctx, cmd.ExternalMatchID)
	if existing != nil {
		return existing, nil // Idempotent
	}

	resourceOwner := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewExternalOracleResult(resourceOwner, cmd.ExternalMatchID, cmd.GameID)

	if err := h.repository.Save(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save oracle result: %w", err)
	}

	slog.InfoContext(ctx, "external oracle result created",
		slog.String("oracle_result_id", result.ID.String()),
		slog.String("external_match_id", cmd.ExternalMatchID),
		slog.String("game_id", string(cmd.GameID)),
	)

	return result, nil
}

// PublishToChain publishes a consensus-reached result to configured blockchains
func (h *oracleCommandHandler) PublishToChain(ctx context.Context, cmd oracle_in.PublishToChainCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.OracleResultID)
	if err != nil {
		return fmt.Errorf("oracle result not found: %w", err)
	}

	if !result.Status.IsPublishable() {
		return fmt.Errorf("oracle result %s is not publishable (status: %s)", result.ID, result.Status)
	}

	// Transition to publishing
	if err := result.MarkPublishing(); err != nil {
		return fmt.Errorf("failed to mark publishing: %w", err)
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update oracle result: %w", err)
	}

	// Determine which chains to publish to
	chainIDs := cmd.ChainIDs
	if len(chainIDs) == 0 && h.chainGateway != nil {
		chainIDs = h.chainGateway.SupportedChains()
	}

	// Publish to each chain
	for _, chainID := range chainIDs {
		if h.chainGateway == nil {
			slog.WarnContext(ctx, "no chain gateway configured, skipping chain publish",
				slog.Int64("chain_id", int64(chainID)),
			)
			continue
		}

		pub, err := h.chainGateway.PublishScore(ctx, chainID, result)
		if err != nil {
			slog.ErrorContext(ctx, "failed to publish to chain",
				slog.Int64("chain_id", int64(chainID)),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := result.AddPublication(*pub); err != nil {
			slog.ErrorContext(ctx, "failed to add publication",
				slog.String("error", err.Error()),
			)
			continue
		}

		slog.InfoContext(ctx, "score published to chain",
			slog.String("oracle_result_id", result.ID.String()),
			slog.Int64("chain_id", int64(chainID)),
			slog.String("tx_hash", pub.TxHash),
		)
	}

	// Update result
	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update oracle result: %w", err)
	}

	// Publish event
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishScorePublished(ctx, result); err != nil {
			slog.ErrorContext(ctx, "failed to publish score published event",
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// HandleDisputeEscalation handles a dispute against a published score
func (h *oracleCommandHandler) HandleDisputeEscalation(ctx context.Context, cmd oracle_in.HandleDisputeCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	result, err := h.repository.FindByID(ctx, cmd.OracleResultID)
	if err != nil {
		return fmt.Errorf("oracle result not found: %w", err)
	}

	if err := result.Dispute(cmd.Reason, cmd.DisputedBy); err != nil {
		return fmt.Errorf("failed to dispute: %w", err)
	}

	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update oracle result: %w", err)
	}

	slog.InfoContext(ctx, "oracle result disputed",
		slog.String("oracle_result_id", result.ID.String()),
		slog.String("reason", cmd.Reason),
		slog.String("disputed_by", cmd.DisputedBy.String()),
	)

	// Publish event
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishScoreDisputed(ctx, result); err != nil {
			slog.ErrorContext(ctx, "failed to publish score disputed event",
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// TriggerIngestionFromAllProviders triggers score ingestion from all available providers
func (h *oracleCommandHandler) TriggerIngestionFromAllProviders(ctx context.Context, cmd oracle_in.TriggerIngestionCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Find or create oracle result
	var result *oracle_entities.OracleResult
	var err error

	if cmd.MatchID != nil {
		result, err = h.repository.FindByMatchID(ctx, *cmd.MatchID)
		if err != nil {
			// Create new oracle result
			resourceOwner := shared.GetResourceOwner(ctx)
			result = oracle_entities.NewOracleResult(resourceOwner, *cmd.MatchID, cmd.GameID)
			if err := h.repository.Save(ctx, result); err != nil {
				return fmt.Errorf("failed to save oracle result: %w", err)
			}
		}
	} else if cmd.ExternalMatchID != nil {
		result, err = h.repository.FindByExternalMatchID(ctx, *cmd.ExternalMatchID)
		if err != nil {
			resourceOwner := shared.GetResourceOwner(ctx)
			result = oracle_entities.NewExternalOracleResult(resourceOwner, *cmd.ExternalMatchID, cmd.GameID)
			if err := h.repository.Save(ctx, result); err != nil {
				return fmt.Errorf("failed to save oracle result: %w", err)
			}
		}
	}

	// Iterate all providers that support this game
	var externalMatchID string
	if cmd.ExternalMatchID != nil {
		externalMatchID = *cmd.ExternalMatchID
	} else if result.ExternalMatchID != nil {
		externalMatchID = *result.ExternalMatchID
	}

	if externalMatchID == "" {
		slog.WarnContext(ctx, "no external match ID available for provider ingestion",
			slog.String("oracle_result_id", result.ID.String()),
		)
		return nil
	}

	ingested := 0
	for _, provider := range h.providers {
		if !provider.SupportsGame(cmd.GameID) {
			continue
		}

		if result.HasSubmissionFromSource(provider.ProviderID()) {
			continue // Already ingested from this source
		}

		submission, err := provider.FetchMatchScore(ctx, externalMatchID, cmd.GameID)
		if err != nil {
			slog.WarnContext(ctx, "failed to fetch score from provider",
				slog.String("provider", string(provider.ProviderID())),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := result.AddSubmission(*submission); err != nil {
			slog.WarnContext(ctx, "failed to add submission from provider",
				slog.String("provider", string(provider.ProviderID())),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Publish ingestion event
		if h.eventPublisher != nil {
			if err := h.eventPublisher.PublishExternalIngested(ctx, result, *submission); err != nil {
				slog.ErrorContext(ctx, "failed to publish external ingested event",
					slog.String("error", err.Error()),
				)
			}
		}

		ingested++
	}

	slog.InfoContext(ctx, "provider ingestion complete",
		slog.String("oracle_result_id", result.ID.String()),
		slog.Int("providers_ingested", ingested),
		slog.Int("total_submissions", result.GetSubmissionCount()),
	)

	// Evaluate consensus if ready
	if result.IsReadyForConsensus(h.policy.MinSources) && result.Status == oracle_vo.OracleStatusPending {
		if err := h.evaluateAndSetConsensus(ctx, result); err != nil {
			slog.WarnContext(ctx, "consensus evaluation failed",
				slog.String("oracle_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	// Persist
	if err := h.repository.Update(ctx, result); err != nil {
		return fmt.Errorf("failed to update oracle result: %w", err)
	}

	return nil
}

// --- Internal Helpers ---

func (h *oracleCommandHandler) findOrCreateOracleResult(ctx context.Context, cmd oracle_in.IngestExternalScoreCommand) (*oracle_entities.OracleResult, error) {
	// Try by oracle result ID first
	if cmd.OracleResultID != nil {
		return h.repository.FindByID(ctx, *cmd.OracleResultID)
	}

	// Try by match ID
	if cmd.MatchID != nil {
		result, err := h.repository.FindByMatchID(ctx, *cmd.MatchID)
		if err == nil {
			return result, nil
		}
		// Create new
		resourceOwner := shared.GetResourceOwner(ctx)
		result = oracle_entities.NewOracleResult(resourceOwner, *cmd.MatchID, cmd.GameID)
		if err := h.repository.Save(ctx, result); err != nil {
			return nil, err
		}
		return result, nil
	}

	// Try by external match ID
	if cmd.ExternalMatchID != nil {
		result, err := h.repository.FindByExternalMatchID(ctx, *cmd.ExternalMatchID)
		if err == nil {
			return result, nil
		}
		resourceOwner := shared.GetResourceOwner(ctx)
		result = oracle_entities.NewExternalOracleResult(resourceOwner, *cmd.ExternalMatchID, cmd.GameID)
		if err := h.repository.Save(ctx, result); err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, fmt.Errorf("no identifier provided to find oracle result")
}

func (h *oracleCommandHandler) evaluateAndSetConsensus(ctx context.Context, result *oracle_entities.OracleResult) error {
	outcome, err := h.consensusEngine.EvaluateConsensus(result.Submissions, h.policy)
	if err != nil {
		return fmt.Errorf("consensus evaluation failed: %w", err)
	}

	if err := result.SetConsensusResult(*outcome); err != nil {
		return fmt.Errorf("failed to set consensus result: %w", err)
	}

	slog.InfoContext(ctx, "consensus reached",
		slog.String("oracle_result_id", result.ID.String()),
		slog.Float64("agreement_ratio", outcome.AgreementRatio),
		slog.Int("confidence_level", outcome.ConfidenceLevel),
		slog.Int("source_count", outcome.SourceCount),
	)

	// Publish consensus reached event
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishConsensusReached(ctx, result); err != nil {
			slog.ErrorContext(ctx, "failed to publish consensus reached event",
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}
