package metadata

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ReconciliationInput contains all data needed to reconcile or create a match.
// This is the canonical input — all callers (OCR import, demo pipeline, external API)
// build this struct and delegate to MatchReconciliationService.
type ReconciliationInput struct {
	GameID          replay_common.GameIDKey
	Source          replay_entity.MatchSource
	ExternalMatchID string
	Provider        string // e.g., "pandascore", "ocr_stream", "demo_parser"
	TeamAName       string
	TeamBName       string
	TeamAScore      int
	TeamBScore      int
	MapName         string
	PlayedAt        time.Time
	GameNumber      int    // For BO3/BO5: which game in the series (1, 2, 3...)
	SeriesType      string // "bo1", "bo3", "bo5"
	Confidence      float64
	ResourceOwner   shared.ResourceOwner
}

// ReconciliationResult contains the outcome of a reconciliation attempt.
type ReconciliationResult struct {
	Match            *replay_entity.Match
	Outcome          replay_entity.ReconciliationOutcome
	ConflictDetected bool
	SlugUsed         string
}

// MatchReconciliationService is the SINGLE SOURCE OF TRUTH for match deduplication.
//
// All match creation flows MUST go through this service to ensure:
//   - Deterministic slug generation with canonicalized team/map names
//   - Atomic upsert (no TOCTOU race conditions)
//   - Multi-source provenance tracking via SourceConfirmations
//   - Conflict detection when scores disagree across sources
//   - ±1 day date variant matching for midnight boundary tolerance
//   - External match ID fallback reconciliation
//
// RECONCILIATION FLOW:
//  1. Validate input (reject empty GameID or both teams + external ID empty)
//  2. Canonicalize team + map names
//  3. Generate primary slug + date variants
//  4. Atomic FindOneAndUpsertBySlug (primary slug)
//  5. If existing → add SourceConfirmation + detect conflicts
//  6. If created → done (first source confirmation already set by factory)
//  7. If primary slug miss from date variant → try ±1 day slugs
//  8. Fallback: FindByExternalMatchID
type MatchReconciliationService struct {
	matchReader replay_out.MatchMetadataReader
	matchWriter replay_out.MatchMetadataWriter
}

// NewMatchReconciliationService creates a new reconciliation service.
func NewMatchReconciliationService(
	matchReader replay_out.MatchMetadataReader,
	matchWriter replay_out.MatchMetadataWriter,
) *MatchReconciliationService {
	return &MatchReconciliationService{
		matchReader: matchReader,
		matchWriter: matchWriter,
	}
}

// ReconcileMatch is the main entry point for all match reconciliation.
// It either finds an existing match (reconciles + adds source confirmation)
// or creates a new match atomically.
func (s *MatchReconciliationService) ReconcileMatch(
	ctx context.Context,
	input ReconciliationInput,
) (*ReconciliationResult, error) {
	// --- 1. Validate ---
	if input.GameID == "" {
		return nil, fmt.Errorf("reconciliation requires non-empty GameID")
	}
	if input.TeamAName == "" && input.TeamBName == "" && input.ExternalMatchID == "" {
		return nil, fmt.Errorf("reconciliation requires at least team names or an external match ID")
	}
	if input.Confidence == 0 {
		input.Confidence = 1.0
	}

	// --- 2. Generate slug variants ---
	slugVariants := replay_entity.GenerateSlugVariants(
		input.GameID,
		input.TeamAName,
		input.TeamBName,
		input.PlayedAt,
		input.MapName,
		input.GameNumber,
		input.SeriesType,
	)
	primarySlug := slugVariants[0]

	slog.InfoContext(ctx, "reconciling match",
		slog.String("slug", primarySlug),
		slog.String("external_match_id", input.ExternalMatchID),
		slog.String("source", string(input.Source)),
		slog.String("provider", input.Provider),
		slog.Int("slug_variants", len(slugVariants)),
	)

	// Build the candidate match for atomic upsert
	candidateMatch := replay_entity.NewMatchFromOCRImport(
		input.ResourceOwner,
		input.GameID,
		input.Source,
		input.ExternalMatchID,
		primarySlug,
		input.TeamAName,
		input.TeamBName,
		input.TeamAScore,
		input.TeamBScore,
		input.MapName,
		input.PlayedAt,
	)

	// --- 3. Atomic upsert by primary slug ---
	existing, created, err := s.matchWriter.FindOneAndUpsertBySlug(ctx, primarySlug, *candidateMatch)
	if err != nil {
		return nil, fmt.Errorf("atomic upsert by slug failed: %w", err)
	}

	if created {
		slog.InfoContext(ctx, "created new match via atomic upsert",
			slog.String("slug", primarySlug),
			slog.String("match_id", existing.GetID().String()),
		)
		return &ReconciliationResult{
			Match:    existing,
			Outcome:  replay_entity.ReconciliationCreated,
			SlugUsed: primarySlug,
		}, nil
	}

	// --- 4. Match already existed — add source confirmation ---
	return s.addConfirmationToExisting(ctx, existing, input, replay_entity.ReconciliationReconciled, primarySlug)
}

// addConfirmationToExisting appends a source confirmation to an existing match
// and persists it atomically. Detects score conflicts.
func (s *MatchReconciliationService) addConfirmationToExisting(
	ctx context.Context,
	existing *replay_entity.Match,
	input ReconciliationInput,
	outcome replay_entity.ReconciliationOutcome,
	slugUsed string,
) (*ReconciliationResult, error) {
	confirmation := replay_entity.SourceConfirmation{
		Source:          input.Source,
		ExternalMatchID: input.ExternalMatchID,
		Provider:        input.Provider,
		TeamAName:       input.TeamAName,
		TeamBName:       input.TeamBName,
		TeamAScore:      input.TeamAScore,
		TeamBScore:      input.TeamBScore,
		MapName:         replay_entity.CanonicalizeMapName(input.MapName),
		ConfirmedAt:     time.Now().UTC(),
		Confidence:      input.Confidence,
	}

	conflictDetected := existing.AddSourceConfirmation(confirmation)

	if conflictDetected {
		outcome = replay_entity.ReconciliationReconciledConflict
		slog.WarnContext(ctx, "SCORE CONFLICT detected during reconciliation",
			slog.String("match_id", existing.GetID().String()),
			slog.String("slug", slugUsed),
			slog.String("conflict", existing.ConflictDetails),
			slog.String("new_source", string(input.Source)),
		)
	}

	// Persist the confirmation atomically via $push (not full doc $set)
	if err := s.matchWriter.AppendSourceConfirmation(
		ctx,
		existing.GetID(),
		confirmation,
		existing.NeedsReview,
		existing.ConflictDetails,
	); err != nil {
		slog.ErrorContext(ctx, "failed to persist source confirmation",
			slog.String("match_id", existing.GetID().String()),
			slog.String("error", err.Error()),
		)
		// Non-fatal: the match was already found, return it anyway
	}

	slog.InfoContext(ctx, "reconciled with existing match",
		slog.String("outcome", string(outcome)),
		slog.String("match_id", existing.GetID().String()),
		slog.String("slug", slugUsed),
		slog.Int("total_confirmations", len(existing.SourceConfirmations)),
		slog.Bool("conflict", conflictDetected),
	)

	return &ReconciliationResult{
		Match:            existing,
		Outcome:          outcome,
		ConflictDetected: conflictDetected,
		SlugUsed:         slugUsed,
	}, nil
}

// ReconcileByExternalMatchID performs reconciliation using only an external match ID.
// This is a fallback when slug-based reconciliation isn't possible.
func (s *MatchReconciliationService) ReconcileByExternalMatchID(
	ctx context.Context,
	input ReconciliationInput,
) (*ReconciliationResult, error) {
	if input.ExternalMatchID == "" {
		return nil, fmt.Errorf("ReconcileByExternalMatchID requires non-empty ExternalMatchID")
	}

	existing, err := s.matchReader.FindByExternalMatchID(ctx, input.ExternalMatchID)
	if err != nil {
		return nil, fmt.Errorf("failed to find match by external_match_id: %w", err)
	}

	if existing != nil {
		// Back-fill slug if missing
		if existing.Slug == "" && input.TeamAName != "" {
			slug := replay_entity.GenerateMatchSlug(
				input.GameID, input.TeamAName, input.TeamBName,
				input.PlayedAt, input.MapName, input.GameNumber, input.SeriesType,
			)
			existing.Slug = slug
			if updateErr := s.matchWriter.Update(ctx, *existing); updateErr != nil {
				slog.WarnContext(ctx, "failed to back-fill slug on existing match",
					slog.String("match_id", existing.GetID().String()),
					slog.String("error", updateErr.Error()),
				)
			}
		}

		return s.addConfirmationToExisting(ctx, existing, input, replay_entity.ReconciliationReconciledExtID, existing.Slug)
	}

	return nil, nil // Not found
}

// ReconcileMatchFull performs full reconciliation with all fallback strategies:
// 1. Primary slug (atomic upsert)
// 2. Date-shifted slug variants (±1 day)
// 3. External match ID fallback
// This is the recommended entry point for maximum reconciliation coverage.
func (s *MatchReconciliationService) ReconcileMatchFull(
	ctx context.Context,
	input ReconciliationInput,
) (*ReconciliationResult, error) {
	// Try primary reconciliation (includes atomic upsert)
	result, err := s.ReconcileMatch(ctx, input)
	if err != nil {
		return nil, err
	}

	// If created (no existing match found via primary slug), check date variants
	// to catch matches recorded on adjacent UTC dates
	if result.Outcome == replay_entity.ReconciliationCreated && !input.PlayedAt.IsZero() {
		slugVariants := replay_entity.GenerateSlugVariants(
			input.GameID, input.TeamAName, input.TeamBName,
			input.PlayedAt, input.MapName, input.GameNumber, input.SeriesType,
		)

		// Check date-shifted variants (skip primary, already tried via atomic upsert)
		for _, variant := range slugVariants[1:] {
			existing, findErr := s.matchReader.FindBySlug(ctx, variant)
			if findErr != nil {
				slog.WarnContext(ctx, "failed to check date-variant slug",
					slog.String("variant", variant),
					slog.String("error", findErr.Error()),
				)
				continue
			}

			if existing != nil {
				slog.InfoContext(ctx, "found match via date-shifted slug variant",
					slog.String("primary_slug", result.SlugUsed),
					slog.String("matched_variant", variant),
					slog.String("existing_match_id", existing.GetID().String()),
				)

				// Link the newly created match to the date-variant match
				existing.LinkMatch(result.Match.GetID())
				result.Match.LinkMatch(existing.GetID())

				// Add confirmation to the older match
				confirmResult, confErr := s.addConfirmationToExisting(
					ctx, existing, input,
					replay_entity.ReconciliationReconciledDateShift, variant,
				)
				if confErr != nil {
					slog.WarnContext(ctx, "failed to add confirmation to date-variant match",
						slog.String("error", confErr.Error()),
					)
				}

				// Persist the links
				_ = s.matchWriter.Update(ctx, *result.Match)

				if confirmResult != nil {
					return confirmResult, nil
				}
				break
			}
		}

		// Fallback: try external match ID
		if input.ExternalMatchID != "" {
			extResult, extErr := s.ReconcileByExternalMatchID(ctx, input)
			if extErr != nil {
				slog.WarnContext(ctx, "external match ID fallback failed",
					slog.String("error", extErr.Error()),
				)
			}
			if extResult != nil {
				// Link both ways
				extResult.Match.LinkMatch(result.Match.GetID())
				result.Match.LinkMatch(extResult.Match.GetID())
				_ = s.matchWriter.Update(ctx, *result.Match)
				_ = s.matchWriter.Update(ctx, *extResult.Match)
				return extResult, nil
			}
		}
	}

	return result, nil
}
