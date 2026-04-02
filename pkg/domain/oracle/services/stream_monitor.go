package oracle_services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// StreamMonitorConfig configures a single stream monitoring session
type StreamMonitorConfig struct {
	StreamURL              string                 `json:"stream_url"`
	GameID                 replay_common.GameIDKey `json:"game_id"`
	ExternalMatchID        string                 `json:"external_match_id"`
	CaptureIntervalSeconds int                    `json:"capture_interval_seconds"`
	ScoreboardRegion       *oracle_out.Region     `json:"scoreboard_region,omitempty"`
	TeamAHint              string                 `json:"team_a_hint,omitempty"` // Optional - helps OCR team matching
	TeamBHint              string                 `json:"team_b_hint,omitempty"`
	IsVOD                  bool                   `json:"is_vod,omitempty"` // Skip liveness check for VODs
}

// StreamMonitor orchestrates frame capture, OCR, parsing, and ingestion
type StreamMonitor struct {
	streamCapture  oracle_out.StreamCapturePort
	ocrEngine      oracle_out.OCREnginePort
	teamResolver   oracle_out.TeamResolverPort
	scoreParser    *OCRScoreParser
	commandHandler oracle_in.OracleCommandHandler

	mu        sync.Mutex
	lastScore *ParsedScore // Deduplication: skip if score unchanged

	// Metrics
	FramesProcessed int64
	ScoresDetected  int64
	Errors          int64
}

// NewStreamMonitor creates a new stream monitor with all dependencies
func NewStreamMonitor(
	streamCapture oracle_out.StreamCapturePort,
	ocrEngine oracle_out.OCREnginePort,
	teamResolver oracle_out.TeamResolverPort,
	scoreParser *OCRScoreParser,
	commandHandler oracle_in.OracleCommandHandler,
) *StreamMonitor {
	return &StreamMonitor{
		streamCapture:  streamCapture,
		ocrEngine:      ocrEngine,
		teamResolver:   teamResolver,
		scoreParser:    scoreParser,
		commandHandler: commandHandler,
	}
}

// MonitorStream continuously monitors a stream, extracting scores at the configured interval.
// Blocks until the context is cancelled or the stream ends.
// For VODs (config.IsVOD=true), captures a single frame and exits after successful ingestion.
func (m *StreamMonitor) MonitorStream(ctx context.Context, config StreamMonitorConfig) error {
	interval := time.Duration(config.CaptureIntervalSeconds) * time.Second
	if interval < 3*time.Second {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.InfoContext(ctx, "stream monitoring started",
		slog.String("stream_url", config.StreamURL),
		slog.String("game_id", string(config.GameID)),
		slog.String("external_match_id", config.ExternalMatchID),
		slog.Int("interval_sec", int(interval.Seconds())),
		slog.Bool("is_vod", config.IsVOD),
	)

	// For VODs, limit retries — the video content is static
	const maxVODAttempts = 5
	vodAttempts := 0

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "stream monitoring stopped (context cancelled)",
				slog.String("stream_url", config.StreamURL),
				slog.Int64("frames_processed", m.FramesProcessed),
				slog.Int64("scores_detected", m.ScoresDetected),
			)
			return ctx.Err()

		case <-ticker.C:
			if err := m.processFrame(ctx, config); err != nil {
				m.Errors++
				slog.WarnContext(ctx, "frame processing failed",
					slog.String("stream_url", config.StreamURL),
					slog.String("error", err.Error()),
					slog.Int64("total_errors", m.Errors),
				)

				if config.IsVOD {
					vodAttempts++
					if vodAttempts >= maxVODAttempts {
						slog.InfoContext(ctx, "VOD processing exhausted attempts",
							slog.String("stream_url", config.StreamURL),
							slog.Int("attempts", vodAttempts),
						)
						return fmt.Errorf("VOD processing failed after %d attempts", vodAttempts)
					}
				}
				continue
			}

			// For VODs, exit after first successful score detection
			if config.IsVOD && m.ScoresDetected > 0 {
				slog.InfoContext(ctx, "VOD processing complete — score detected",
					slog.String("stream_url", config.StreamURL),
					slog.Int64("scores_detected", m.ScoresDetected),
				)
				return nil
			}
		}
	}
}

// processFrame captures a single frame, runs OCR, and ingests if the score changed
func (m *StreamMonitor) processFrame(ctx context.Context, config StreamMonitorConfig) error {
	// Step 1: Check if stream is still live (skip for VODs — they are never "live")
	if !config.IsVOD {
		live, err := m.streamCapture.IsStreamLive(ctx, config.StreamURL)
		if err != nil {
			return fmt.Errorf("stream liveness check failed: %w", err)
		}
		if !live {
			return fmt.Errorf("stream is not live")
		}
	}

	// Step 2: Capture frame
	frameData, err := m.streamCapture.CaptureFrame(ctx, config.StreamURL)
	if err != nil {
		return fmt.Errorf("frame capture failed: %w", err)
	}
	m.FramesProcessed++

	// Step 3: Run OCR
	textBlocks, err := m.ocrEngine.ExtractText(ctx, frameData, config.ScoreboardRegion)
	if err != nil {
		return fmt.Errorf("OCR extraction failed: %w", err)
	}

	if len(textBlocks) == 0 {
		return nil // No text found — not an error, just nothing to parse
	}

	// Step 4: Parse scoreboard
	parsed, err := m.scoreParser.ParseScoreboard(textBlocks, config.GameID)
	if err != nil {
		// Not finding a scoreboard is normal — the frame might not show it
		return nil
	}

	// Step 5: Deduplicate — skip if score hasn't changed
	if m.isDuplicate(parsed) {
		return nil
	}

	slog.InfoContext(ctx, "new score detected via OCR",
		slog.String("team_a", parsed.TeamAName),
		slog.Int("team_a_score", parsed.TeamAScore),
		slog.String("team_b", parsed.TeamBName),
		slog.Int("team_b_score", parsed.TeamBScore),
		slog.String("map", parsed.MapName),
		slog.Int("rounds", parsed.RoundsPlayed),
	)

	// Step 6: Resolve team names to IDs
	teamARef, teamBRef, err := m.resolveTeams(ctx, parsed, config)
	if err != nil {
		return fmt.Errorf("team resolution failed: %w", err)
	}

	// Step 7: Build and ingest the score submission
	if err := m.ingestScore(ctx, parsed, teamARef, teamBRef, config); err != nil {
		return fmt.Errorf("score ingestion failed: %w", err)
	}

	m.ScoresDetected++
	m.updateLastScore(parsed)
	return nil
}

// isDuplicate checks if the parsed score is identical to the last processed score
func (m *StreamMonitor) isDuplicate(parsed *ParsedScore) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastScore == nil {
		return false
	}

	return m.lastScore.TeamAScore == parsed.TeamAScore &&
		m.lastScore.TeamBScore == parsed.TeamBScore &&
		m.lastScore.TeamAName == parsed.TeamAName &&
		m.lastScore.TeamBName == parsed.TeamBName
}

func (m *StreamMonitor) updateLastScore(parsed *ParsedScore) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastScore = parsed
}

// resolveTeams resolves OCR team names to team IDs, falling back to hints or deterministic UUIDs
func (m *StreamMonitor) resolveTeams(ctx context.Context, parsed *ParsedScore, config StreamMonitorConfig) (*oracle_out.TeamRef, *oracle_out.TeamRef, error) {
	// Use team name hints if provided and OCR didn't detect names
	teamAName := parsed.TeamAName
	if teamAName == "" && config.TeamAHint != "" {
		teamAName = config.TeamAHint
	}
	teamBName := parsed.TeamBName
	if teamBName == "" && config.TeamBHint != "" {
		teamBName = config.TeamBHint
	}

	var teamARef, teamBRef *oracle_out.TeamRef

	if teamAName != "" && m.teamResolver != nil {
		ref, err := m.teamResolver.ResolveTeam(ctx, teamAName, config.GameID)
		if err != nil {
			slog.WarnContext(ctx, "could not resolve team A, using deterministic ID",
				slog.String("team_name", teamAName),
				slog.String("error", err.Error()),
			)
			teamARef = &oracle_out.TeamRef{
				TeamID:      deterministicTeamUUID(teamAName),
				MatchedName: teamAName,
				Confidence:  0.3,
			}
		} else {
			teamARef = ref
		}
	} else {
		teamARef = &oracle_out.TeamRef{
			TeamID:      deterministicTeamUUID(teamAName),
			MatchedName: teamAName,
			Confidence:  0.1,
		}
	}

	if teamBName != "" && m.teamResolver != nil {
		ref, err := m.teamResolver.ResolveTeam(ctx, teamBName, config.GameID)
		if err != nil {
			slog.WarnContext(ctx, "could not resolve team B, using deterministic ID",
				slog.String("team_name", teamBName),
				slog.String("error", err.Error()),
			)
			teamBRef = &oracle_out.TeamRef{
				TeamID:      deterministicTeamUUID(teamBName),
				MatchedName: teamBName,
				Confidence:  0.3,
			}
		} else {
			teamBRef = ref
		}
	} else {
		teamBRef = &oracle_out.TeamRef{
			TeamID:      deterministicTeamUUID(teamBName),
			MatchedName: teamBName,
			Confidence:  0.1,
		}
	}

	return teamARef, teamBRef, nil
}

// ingestScore builds an IngestExternalScoreCommand and submits it
func (m *StreamMonitor) ingestScore(ctx context.Context, parsed *ParsedScore, teamA, teamB *oracle_out.TeamRef, config StreamMonitorConfig) error {
	// Determine winner
	var winnerID *uuid.UUID
	isDraw := parsed.TeamAScore == parsed.TeamBScore
	if !isDraw {
		if parsed.TeamAScore > parsed.TeamBScore {
			w := teamA.TeamID
			winnerID = &w
		} else {
			w := teamB.TeamID
			winnerID = &w
		}
	}

	// Build game details if map is known
	var gameDetails []oracle_entities.SubmissionGameDetail
	if parsed.MapName != "" {
		gameDetails = []oracle_entities.SubmissionGameDetail{
			{
				Position:   1,
				MapName:    parsed.MapName,
				TeamAScore: parsed.TeamAScore,
				TeamBScore: parsed.TeamBScore,
				TeamAWon:   parsed.TeamAScore > parsed.TeamBScore,
			},
		}
	}

	// Build raw response for provenance
	rawData, _ := json.Marshal(parsed)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256(rawData))

	// Use a unique provider match ID combining external match ID + score state
	providerMatchID := fmt.Sprintf("ocr_%s_%d_%d_%s",
		config.ExternalMatchID,
		parsed.TeamAScore,
		parsed.TeamBScore,
		sourceHash[:8],
	)

	extMatchID := config.ExternalMatchID
	cmd := oracle_in.IngestExternalScoreCommand{
		ExternalMatchID: &extMatchID,
		GameID:          config.GameID,
		SourceType:      oracle_vo.OracleSourceOCRStream,
		ProviderMatchID: providerMatchID,
		WinnerTeamID:    winnerID,
		IsDraw:          isDraw,
		TeamAID:         teamA.TeamID,
		TeamBID:         teamB.TeamID,
		TeamAScore:      parsed.TeamAScore,
		TeamBScore:      parsed.TeamBScore,
		RoundsPlayed:    parsed.RoundsPlayed,
		GameDetails:     gameDetails,
	}

	return m.commandHandler.IngestExternalScore(ctx, cmd)
}

// deterministicTeamUUID generates a deterministic UUID from a team name string
func deterministicTeamUUID(name string) uuid.UUID {
	data := fmt.Sprintf("ocr_team:%s", name)
	hash := sha256.Sum256([]byte(data))
	id, _ := uuid.FromBytes(hash[:16])
	return id
}
