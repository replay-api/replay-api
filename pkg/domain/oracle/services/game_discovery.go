package oracle_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// GameDiscoveryService polls external providers for recently completed matches
// that are not yet tracked by the platform. It discovers new games and emits them
// for import processing.
type GameDiscoveryService struct {
	providers         []oracle_out.ExternalScorePort
	oracleResultRepo  oracle_out.OracleResultRepository
	streamConfigRepo  oracle_out.OCRStreamConfigRepository
	eventPublisher    oracle_out.OracleEventPublisher
	supportedGames    []replay_common.GameIDKey

	// Discovery configuration
	PollingInterval   time.Duration
	LookbackWindow    time.Duration
	MaxMatchesPerPoll int
}

// GameDiscoveryConfig holds configuration for the discovery service.
type GameDiscoveryConfig struct {
	PollingInterval   time.Duration
	LookbackWindow    time.Duration
	MaxMatchesPerPoll int
	SupportedGames    []replay_common.GameIDKey
}

// DefaultGameDiscoveryConfig returns sensible defaults for game discovery.
func DefaultGameDiscoveryConfig() GameDiscoveryConfig {
	return GameDiscoveryConfig{
		PollingInterval:   5 * time.Minute,
		LookbackWindow:    24 * time.Hour,
		MaxMatchesPerPoll: 50,
		SupportedGames: []replay_common.GameIDKey{
			replay_common.CS2_GAME_ID,
		},
	}
}

// NewGameDiscoveryService creates a new game discovery service.
func NewGameDiscoveryService(
	providers []oracle_out.ExternalScorePort,
	oracleResultRepo oracle_out.OracleResultRepository,
	streamConfigRepo oracle_out.OCRStreamConfigRepository,
	eventPublisher oracle_out.OracleEventPublisher,
	config GameDiscoveryConfig,
) *GameDiscoveryService {
	return &GameDiscoveryService{
		providers:         providers,
		oracleResultRepo:  oracleResultRepo,
		streamConfigRepo:  streamConfigRepo,
		eventPublisher:    eventPublisher,
		supportedGames:    config.SupportedGames,
		PollingInterval:   config.PollingInterval,
		LookbackWindow:    config.LookbackWindow,
		MaxMatchesPerPoll: config.MaxMatchesPerPoll,
	}
}

// DiscoveredMatch encapsulates a match discovered by the service with its metadata.
type DiscoveredMatch struct {
	ExternalMatch oracle_out.ExternalMatch
	IsNew         bool   // true if not yet tracked in the oracle system
	HasVOD        bool   // true if VOD URLs are available for OCR
	HasStream     bool   // true if a live stream is available
}

// DiscoverRecentMatches polls all providers for recently completed matches,
// deduplicates against existing oracle results, and returns new discoveries.
func (s *GameDiscoveryService) DiscoverRecentMatches(ctx context.Context) ([]DiscoveredMatch, error) {
	since := time.Now().UTC().Add(-s.LookbackWindow)
	var allDiscoveries []DiscoveredMatch

	for _, gameID := range s.supportedGames {
		for _, provider := range s.providers {
			if !provider.SupportsGame(gameID) {
				continue
			}

			matches, err := provider.ListRecentMatches(ctx, gameID, since, s.MaxMatchesPerPoll)
			if err != nil {
				slog.WarnContext(ctx, "failed to fetch recent matches from provider",
					slog.String("provider", string(provider.ProviderID())),
					slog.String("game_id", string(gameID)),
					slog.String("error", err.Error()),
				)
				continue
			}

			for _, m := range matches {
				// Skip unfinished matches
				if m.Status != "finished" {
					continue
				}

				// Dedup: check if oracle result already exists for this external match
				existing, err := s.oracleResultRepo.FindByExternalMatchID(ctx, m.ExternalMatchID)
				if err != nil {
					slog.WarnContext(ctx, "failed to check oracle result dedup",
						slog.String("external_match_id", m.ExternalMatchID),
						slog.String("error", err.Error()),
					)
					continue
				}

				isNew := existing == nil

				discovery := DiscoveredMatch{
					ExternalMatch: m,
					IsNew:         isNew,
					HasVOD:        len(m.VODURLs) > 0,
					HasStream:     m.StreamURL != "",
				}

				allDiscoveries = append(allDiscoveries, discovery)
			}
		}
	}

	newCount := 0
	for _, d := range allDiscoveries {
		if d.IsNew {
			newCount++
		}
	}

	slog.InfoContext(ctx, "game discovery completed",
		slog.Int("total_discovered", len(allDiscoveries)),
		slog.Int("new_matches", newCount),
		slog.Int("providers", len(s.providers)),
		slog.Int("games", len(s.supportedGames)),
	)

	return allDiscoveries, nil
}

// RunDiscoveryLoop runs the discovery polling loop until the context is cancelled.
// For each new match discovered, it calls the onDiscovered callback.
func (s *GameDiscoveryService) RunDiscoveryLoop(ctx context.Context, onDiscovered func(ctx context.Context, match DiscoveredMatch) error) error {
	ticker := time.NewTicker(s.PollingInterval)
	defer ticker.Stop()

	slog.InfoContext(ctx, "game discovery loop starting",
		slog.Duration("polling_interval", s.PollingInterval),
		slog.Duration("lookback_window", s.LookbackWindow),
		slog.Int("max_per_poll", s.MaxMatchesPerPoll),
	)

	// Run immediately on start
	if err := s.discoverAndProcess(ctx, onDiscovered); err != nil {
		slog.WarnContext(ctx, "initial discovery pass failed", slog.String("error", err.Error()))
	}

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "game discovery loop stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.discoverAndProcess(ctx, onDiscovered); err != nil {
				slog.WarnContext(ctx, "discovery pass failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (s *GameDiscoveryService) discoverAndProcess(ctx context.Context, onDiscovered func(ctx context.Context, match DiscoveredMatch) error) error {
	discoveries, err := s.DiscoverRecentMatches(ctx)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	for _, d := range discoveries {
		if !d.IsNew {
			continue
		}

		if err := onDiscovered(ctx, d); err != nil {
			slog.WarnContext(ctx, "failed to process discovered match",
				slog.String("external_match_id", d.ExternalMatch.ExternalMatchID),
				slog.String("provider", string(d.ExternalMatch.Provider)),
				slog.String("error", err.Error()),
			)
			continue
		}
	}

	return nil
}
