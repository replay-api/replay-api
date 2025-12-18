package integrity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// IntegrityStatus represents the verification status
type IntegrityStatus string

const (
	IntegrityStatusPending   IntegrityStatus = "PENDING"
	IntegrityStatusValid     IntegrityStatus = "VALID"
	IntegrityStatusInvalid   IntegrityStatus = "INVALID"
	IntegrityStatusSuspicious IntegrityStatus = "SUSPICIOUS"
	IntegrityStatusFlagged   IntegrityStatus = "FLAGGED"
)

// ViolationType categorizes detected violations
type ViolationType string

const (
	// File integrity violations
	ViolationHashMismatch      ViolationType = "HASH_MISMATCH"
	ViolationFileTampered      ViolationType = "FILE_TAMPERED"
	ViolationInvalidFormat     ViolationType = "INVALID_FORMAT"
	ViolationCorrupted         ViolationType = "CORRUPTED"
	
	// Gameplay violations
	ViolationAimbotDetected    ViolationType = "AIMBOT_DETECTED"
	ViolationWallhackIndicator ViolationType = "WALLHACK_INDICATOR"
	ViolationSpeedAnomaly      ViolationType = "SPEED_ANOMALY"
	ViolationTeleportDetected  ViolationType = "TELEPORT_DETECTED"
	ViolationSpinbotPattern    ViolationType = "SPINBOT_PATTERN"
	ViolationBunnyHopAnomaly   ViolationType = "BHOP_ANOMALY"
	
	// Statistical violations
	ViolationImpossibleHeadshots ViolationType = "IMPOSSIBLE_HEADSHOTS"
	ViolationAbnormalReaction    ViolationType = "ABNORMAL_REACTION_TIME"
	ViolationPerfectRecoil       ViolationType = "PERFECT_RECOIL_CONTROL"
	ViolationPrefirePattern      ViolationType = "SUSPICIOUS_PREFIRE"
	
	// Data manipulation
	ViolationTimestampAnomaly  ViolationType = "TIMESTAMP_ANOMALY"
	ViolationTickManipulation  ViolationType = "TICK_MANIPULATION"
	ViolationEventInjection    ViolationType = "EVENT_INJECTION"
)

// IntegrityViolation represents a detected violation
type IntegrityViolation struct {
	ID            uuid.UUID      `json:"id" bson:"_id"`
	Type          ViolationType  `json:"type" bson:"type"`
	Severity      ViolationSeverity `json:"severity" bson:"severity"`
	Description   string         `json:"description" bson:"description"`
	PlayerID      *uuid.UUID     `json:"player_id,omitempty" bson:"player_id,omitempty"`
	Timestamp     time.Time      `json:"timestamp" bson:"timestamp"`
	TickNumber    *int64         `json:"tick_number,omitempty" bson:"tick_number,omitempty"`
	Evidence      map[string]interface{} `json:"evidence" bson:"evidence"`
	Confidence    float64        `json:"confidence" bson:"confidence"` // 0.0 - 1.0
}

// ViolationSeverity indicates violation importance
type ViolationSeverity string

const (
	SeverityLow      ViolationSeverity = "LOW"
	SeverityMedium   ViolationSeverity = "MEDIUM"
	SeverityHigh     ViolationSeverity = "HIGH"
	SeverityCritical ViolationSeverity = "CRITICAL"
)

// ReplayIntegrityReport contains full integrity analysis
type ReplayIntegrityReport struct {
	ID              uuid.UUID          `json:"id" bson:"_id"`
	ReplayID        uuid.UUID          `json:"replay_id" bson:"replay_id"`
	GameID          string             `json:"game_id" bson:"game_id"`
	Status          IntegrityStatus    `json:"status" bson:"status"`
	
	// File integrity
	FileHash        string             `json:"file_hash" bson:"file_hash"`
	ExpectedHash    string             `json:"expected_hash,omitempty" bson:"expected_hash,omitempty"`
	FileSize        int64              `json:"file_size" bson:"file_size"`
	
	// Analysis results
	Violations      []IntegrityViolation `json:"violations" bson:"violations"`
	ViolationCount  int                `json:"violation_count" bson:"violation_count"`
	OverallScore    float64            `json:"overall_score" bson:"overall_score"` // 0-100, higher = more suspicious
	
	// Player analysis
	PlayerReports   []PlayerIntegrityReport `json:"player_reports" bson:"player_reports"`
	
	// Timing
	AnalyzedAt      time.Time          `json:"analyzed_at" bson:"analyzed_at"`
	AnalysisDuration time.Duration     `json:"analysis_duration" bson:"analysis_duration"`
	
	// Metadata
	AnalyzerVersion string             `json:"analyzer_version" bson:"analyzer_version"`
	GameVersion     string             `json:"game_version" bson:"game_version"`
	MapName         string             `json:"map_name" bson:"map_name"`
	MatchDuration   time.Duration      `json:"match_duration" bson:"match_duration"`
	
	// Review status
	ReviewRequired  bool               `json:"review_required" bson:"review_required"`
	ReviewedBy      *uuid.UUID         `json:"reviewed_by,omitempty" bson:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time         `json:"reviewed_at,omitempty" bson:"reviewed_at,omitempty"`
	ReviewNotes     string             `json:"review_notes,omitempty" bson:"review_notes,omitempty"`
	FinalVerdict    *IntegrityStatus   `json:"final_verdict,omitempty" bson:"final_verdict,omitempty"`
}

// PlayerIntegrityReport contains per-player analysis
type PlayerIntegrityReport struct {
	PlayerID        uuid.UUID          `json:"player_id" bson:"player_id"`
	SteamID         string             `json:"steam_id,omitempty" bson:"steam_id,omitempty"`
	Status          IntegrityStatus    `json:"status" bson:"status"`
	Score           float64            `json:"score" bson:"score"`
	Violations      []IntegrityViolation `json:"violations" bson:"violations"`
	
	// Statistical analysis
	Stats           PlayerMatchStats   `json:"stats" bson:"stats"`
	AnomalyFlags    []string           `json:"anomaly_flags" bson:"anomaly_flags"`
}

// PlayerMatchStats contains statistical data for analysis
type PlayerMatchStats struct {
	Kills           int     `json:"kills"`
	Deaths          int     `json:"deaths"`
	Assists         int     `json:"assists"`
	Headshots       int     `json:"headshots"`
	HeadshotPercent float64 `json:"headshot_percent"`
	
	// Aiming metrics
	ShotsTotal      int     `json:"shots_total"`
	ShotsHit        int     `json:"shots_hit"`
	Accuracy        float64 `json:"accuracy"`
	
	// Reaction times (ms)
	AvgReactionTime float64 `json:"avg_reaction_time_ms"`
	MinReactionTime float64 `json:"min_reaction_time_ms"`
	MaxReactionTime float64 `json:"max_reaction_time_ms"`
	
	// Movement
	AvgMovementSpeed float64 `json:"avg_movement_speed"`
	MaxMovementSpeed float64 `json:"max_movement_speed"`
	BunnyHopCount    int     `json:"bunny_hop_count"`
	PerfectBhops     int     `json:"perfect_bhops"`
	
	// View angles
	AvgAimSpeed      float64 `json:"avg_aim_speed_deg_per_sec"`
	MaxAimSpeed      float64 `json:"max_aim_speed_deg_per_sec"`
	SnapCount        int     `json:"snap_count"` // Suspicious instant aim adjustments
}

// ReplayIntegrityService provides replay integrity verification
type ReplayIntegrityService struct {
	mu               sync.RWMutex
	reportRepo       IntegrityReportRepository
	antiCheatHooks   []AntiCheatHook
	thresholds       IntegrityThresholds
	analyzerVersion  string
}

// IntegrityReportRepository defines persistence
type IntegrityReportRepository interface {
	Create(ctx context.Context, report *ReplayIntegrityReport) error
	GetByReplayID(ctx context.Context, replayID uuid.UUID) (*ReplayIntegrityReport, error)
	Update(ctx context.Context, report *ReplayIntegrityReport) error
	GetPendingReviews(ctx context.Context, limit int) ([]ReplayIntegrityReport, error)
}

// AntiCheatHook defines an interface for anti-cheat integrations
type AntiCheatHook interface {
	Name() string
	Analyze(ctx context.Context, data *ReplayAnalysisData) ([]IntegrityViolation, error)
}

// ReplayAnalysisData contains replay data for analysis
type ReplayAnalysisData struct {
	ReplayID    uuid.UUID
	GameID      string
	FileReader  io.Reader
	FileSize    int64
	FileHash    string
	Metadata    map[string]interface{}
	
	// Parsed data (set by analyzers)
	Events      []GameEvent
	Players     []PlayerData
	TickData    []TickData
}

// GameEvent represents a game event from the replay
type GameEvent struct {
	Tick        int64
	Type        string
	PlayerID    uuid.UUID
	TargetID    *uuid.UUID
	Position    Vector3
	ViewAngles  Vector2
	Data        map[string]interface{}
	Timestamp   time.Duration
}

// PlayerData contains player information from replay
type PlayerData struct {
	PlayerID    uuid.UUID
	SteamID     string
	Team        int
	Name        string
}

// TickData contains per-tick game state
type TickData struct {
	Tick        int64
	Players     []PlayerTickState
	Timestamp   time.Duration
}

// PlayerTickState contains player state at a specific tick
type PlayerTickState struct {
	PlayerID    uuid.UUID
	Position    Vector3
	ViewAngles  Vector2
	Velocity    Vector3
	Health      int
	IsAlive     bool
}

// Vector3 represents 3D coordinates
type Vector3 struct {
	X, Y, Z float64
}

// Vector2 represents 2D angles
type Vector2 struct {
	Pitch, Yaw float64
}

// IntegrityThresholds defines detection thresholds
type IntegrityThresholds struct {
	MaxHeadshotPercent    float64 // Above this triggers investigation
	MinReactionTimeMs     float64 // Below this is suspicious
	MaxAimSpeedDegPerSec  float64 // Above this triggers snap detection
	PerfectBhopThreshold  float64 // Ratio of perfect bhops
	HighSuspicionScore    float64 // Triggers manual review
}

// DefaultThresholds provides reasonable defaults
var DefaultThresholds = IntegrityThresholds{
	MaxHeadshotPercent:    75.0,
	MinReactionTimeMs:     80.0,
	MaxAimSpeedDegPerSec:  3000.0,
	PerfectBhopThreshold:  0.8,
	HighSuspicionScore:    50.0,
}

// NewReplayIntegrityService creates a new integrity service
func NewReplayIntegrityService(
	repo IntegrityReportRepository,
	thresholds IntegrityThresholds,
) *ReplayIntegrityService {
	return &ReplayIntegrityService{
		reportRepo:      repo,
		antiCheatHooks:  make([]AntiCheatHook, 0),
		thresholds:      thresholds,
		analyzerVersion: "2.0.0",
	}
}

// RegisterAntiCheatHook registers an anti-cheat integration
func (s *ReplayIntegrityService) RegisterAntiCheatHook(hook AntiCheatHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.antiCheatHooks = append(s.antiCheatHooks, hook)
	slog.Info("Anti-cheat hook registered", "name", hook.Name())
}

// AnalyzeReplay performs comprehensive integrity analysis
func (s *ReplayIntegrityService) AnalyzeReplay(ctx context.Context, data *ReplayAnalysisData) (*ReplayIntegrityReport, error) {
	startTime := time.Now()

	report := &ReplayIntegrityReport{
		ID:              uuid.New(),
		ReplayID:        data.ReplayID,
		GameID:          data.GameID,
		Status:          IntegrityStatusPending,
		FileHash:        data.FileHash,
		FileSize:        data.FileSize,
		Violations:      make([]IntegrityViolation, 0),
		PlayerReports:   make([]PlayerIntegrityReport, 0),
		AnalyzerVersion: s.analyzerVersion,
		AnalyzedAt:      time.Now().UTC(),
	}

	// Calculate file hash if not provided
	if data.FileHash == "" && data.FileReader != nil {
		hash, err := s.calculateFileHash(data.FileReader)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate file hash: %w", err)
		}
		data.FileHash = hash
		report.FileHash = hash
	}

	// Run built-in analyzers
	violations, err := s.runBuiltInAnalyzers(ctx, data)
	if err != nil {
		slog.ErrorContext(ctx, "Built-in analyzer error", "error", err)
	}
	report.Violations = append(report.Violations, violations...)

	// Run registered anti-cheat hooks
	s.mu.RLock()
	hooks := s.antiCheatHooks
	s.mu.RUnlock()

	for _, hook := range hooks {
		hookViolations, err := hook.Analyze(ctx, data)
		if err != nil {
			slog.WarnContext(ctx, "Anti-cheat hook error",
				"hook", hook.Name(),
				"error", err,
			)
			continue
		}
		report.Violations = append(report.Violations, hookViolations...)
	}

	// Analyze each player
	for _, player := range data.Players {
		playerReport := s.analyzePlayer(ctx, data, player)
		report.PlayerReports = append(report.PlayerReports, playerReport)
	}

	// Calculate overall score and determine status
	report.ViolationCount = len(report.Violations)
	report.OverallScore = s.calculateOverallScore(report)
	report.Status = s.determineStatus(report)
	report.ReviewRequired = report.Status == IntegrityStatusSuspicious || report.Status == IntegrityStatusFlagged

	report.AnalysisDuration = time.Since(startTime)

	// Save report
	if s.reportRepo != nil {
		if err := s.reportRepo.Create(ctx, report); err != nil {
			slog.ErrorContext(ctx, "Failed to save integrity report", "error", err)
		}
	}

	slog.InfoContext(ctx, "Replay integrity analysis complete",
		"replay_id", data.ReplayID,
		"status", report.Status,
		"violations", report.ViolationCount,
		"score", report.OverallScore,
		"duration", report.AnalysisDuration,
	)

	return report, nil
}

// VerifyFileHash verifies the replay file hash
func (s *ReplayIntegrityService) VerifyFileHash(ctx context.Context, replayID uuid.UUID, expectedHash string, actualReader io.Reader) (*IntegrityViolation, error) {
	actualHash, err := s.calculateFileHash(actualReader)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	if actualHash != expectedHash {
		return &IntegrityViolation{
			ID:          uuid.New(),
			Type:        ViolationHashMismatch,
			Severity:    SeverityCritical,
			Description: "Replay file hash does not match expected value",
			Timestamp:   time.Now().UTC(),
			Evidence: map[string]interface{}{
				"expected_hash": expectedHash,
				"actual_hash":   actualHash,
			},
			Confidence: 1.0,
		}, nil
	}

	return nil, nil
}

// ReviewReport allows manual review of flagged reports
func (s *ReplayIntegrityService) ReviewReport(ctx context.Context, reportID uuid.UUID, reviewerID uuid.UUID, verdict IntegrityStatus, notes string) error {
	report, err := s.reportRepo.GetByReplayID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("report not found: %w", err)
	}

	now := time.Now().UTC()
	report.ReviewedBy = &reviewerID
	report.ReviewedAt = &now
	report.ReviewNotes = notes
	report.FinalVerdict = &verdict

	if verdict == IntegrityStatusValid {
		report.ReviewRequired = false
	}

	return s.reportRepo.Update(ctx, report)
}

// GetPendingReviews returns reports that need manual review
func (s *ReplayIntegrityService) GetPendingReviews(ctx context.Context, limit int) ([]ReplayIntegrityReport, error) {
	return s.reportRepo.GetPendingReviews(ctx, limit)
}

// calculateFileHash computes SHA-256 hash of file
func (s *ReplayIntegrityService) calculateFileHash(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// runBuiltInAnalyzers runs the built-in analysis routines
func (s *ReplayIntegrityService) runBuiltInAnalyzers(ctx context.Context, data *ReplayAnalysisData) ([]IntegrityViolation, error) {
	violations := make([]IntegrityViolation, 0)

	// Analyze game events for anomalies
	if len(data.Events) > 0 {
		eventViolations := s.analyzeEvents(data.Events)
		violations = append(violations, eventViolations...)
	}

	// Analyze tick data for manipulation
	if len(data.TickData) > 0 {
		tickViolations := s.analyzeTickData(data.TickData)
		violations = append(violations, tickViolations...)
	}

	return violations, nil
}

// analyzeEvents checks for suspicious event patterns
func (s *ReplayIntegrityService) analyzeEvents(events []GameEvent) []IntegrityViolation {
	violations := make([]IntegrityViolation, 0)

	var lastTick int64
	for i, event := range events {
		// Check for timestamp anomalies
		if event.Tick < lastTick {
			violations = append(violations, IntegrityViolation{
				ID:          uuid.New(),
				Type:        ViolationTimestampAnomaly,
				Severity:    SeverityHigh,
				Description: "Event tick is earlier than previous tick",
				TickNumber:  &event.Tick,
				Timestamp:   time.Now().UTC(),
				Evidence: map[string]interface{}{
					"event_index":   i,
					"current_tick":  event.Tick,
					"previous_tick": lastTick,
				},
				Confidence: 0.9,
			})
		}
		lastTick = event.Tick

		// Check for impossible view angle changes (snap detection)
		if i > 0 && events[i-1].PlayerID == event.PlayerID {
			angleDiff := calculateAngleDifference(events[i-1].ViewAngles, event.ViewAngles)
			tickDiff := event.Tick - events[i-1].Tick
			if tickDiff > 0 {
				angleSpeed := angleDiff / float64(tickDiff) * 128 // Assuming 128 tick
				if angleSpeed > s.thresholds.MaxAimSpeedDegPerSec {
					violations = append(violations, IntegrityViolation{
						ID:          uuid.New(),
						Type:        ViolationSpinbotPattern,
						Severity:    SeverityCritical,
						Description: "Impossible aim speed detected",
						PlayerID:    &event.PlayerID,
						TickNumber:  &event.Tick,
						Timestamp:   time.Now().UTC(),
						Evidence: map[string]interface{}{
							"angle_speed":    angleSpeed,
							"threshold":      s.thresholds.MaxAimSpeedDegPerSec,
							"angle_diff":     angleDiff,
							"tick_diff":      tickDiff,
						},
						Confidence: 0.85,
					})
				}
			}
		}
	}

	return violations
}

// analyzeTickData checks for tick manipulation
func (s *ReplayIntegrityService) analyzeTickData(ticks []TickData) []IntegrityViolation {
	violations := make([]IntegrityViolation, 0)

	for i := 1; i < len(ticks); i++ {
		current := ticks[i]
		previous := ticks[i-1]

		// Check for tick gaps
		tickDiff := current.Tick - previous.Tick
		if tickDiff > 5 { // More than 5 ticks gap is suspicious
			violations = append(violations, IntegrityViolation{
				ID:          uuid.New(),
				Type:        ViolationTickManipulation,
				Severity:    SeverityMedium,
				Description: "Unusual tick gap detected",
				TickNumber:  &current.Tick,
				Timestamp:   time.Now().UTC(),
				Evidence: map[string]interface{}{
					"tick_gap":     tickDiff,
					"current_tick": current.Tick,
				},
				Confidence: 0.6,
			})
		}

		// Check for teleportation (position changes too large for time elapsed)
		for _, playerState := range current.Players {
			for _, prevState := range previous.Players {
				if playerState.PlayerID == prevState.PlayerID && playerState.IsAlive && prevState.IsAlive {
					distance := calculateDistance(playerState.Position, prevState.Position)
					maxPossibleDistance := float64(tickDiff) * 400 // ~400 units/tick max
					if distance > maxPossibleDistance*1.5 { // 50% tolerance
						violations = append(violations, IntegrityViolation{
							ID:          uuid.New(),
							Type:        ViolationTeleportDetected,
							Severity:    SeverityCritical,
							Description: "Player position change exceeds possible movement speed",
							PlayerID:    &playerState.PlayerID,
							TickNumber:  &current.Tick,
							Timestamp:   time.Now().UTC(),
							Evidence: map[string]interface{}{
								"distance":     distance,
								"max_possible": maxPossibleDistance,
								"from":         prevState.Position,
								"to":           playerState.Position,
							},
							Confidence: 0.95,
						})
					}
				}
			}
		}
	}

	return violations
}

// analyzePlayer performs per-player analysis
func (s *ReplayIntegrityService) analyzePlayer(ctx context.Context, data *ReplayAnalysisData, player PlayerData) PlayerIntegrityReport {
	report := PlayerIntegrityReport{
		PlayerID:     player.PlayerID,
		SteamID:      player.SteamID,
		Status:       IntegrityStatusValid,
		Violations:   make([]IntegrityViolation, 0),
		AnomalyFlags: make([]string, 0),
	}

	// Calculate player stats from events
	stats := s.calculatePlayerStats(data, player.PlayerID)
	report.Stats = stats

	// Check for statistical anomalies
	if stats.HeadshotPercent > s.thresholds.MaxHeadshotPercent {
		report.AnomalyFlags = append(report.AnomalyFlags, "HIGH_HEADSHOT_RATE")
		report.Violations = append(report.Violations, IntegrityViolation{
			ID:          uuid.New(),
			Type:        ViolationImpossibleHeadshots,
			Severity:    SeverityHigh,
			Description: fmt.Sprintf("Headshot rate %.1f%% exceeds threshold", stats.HeadshotPercent),
			PlayerID:    &player.PlayerID,
			Timestamp:   time.Now().UTC(),
			Evidence: map[string]interface{}{
				"headshot_percent": stats.HeadshotPercent,
				"threshold":        s.thresholds.MaxHeadshotPercent,
				"headshots":        stats.Headshots,
				"kills":            stats.Kills,
			},
			Confidence: 0.7,
		})
	}

	if stats.MinReactionTime > 0 && stats.MinReactionTime < s.thresholds.MinReactionTimeMs {
		report.AnomalyFlags = append(report.AnomalyFlags, "INHUMAN_REACTION")
		report.Violations = append(report.Violations, IntegrityViolation{
			ID:          uuid.New(),
			Type:        ViolationAbnormalReaction,
			Severity:    SeverityHigh,
			Description: fmt.Sprintf("Reaction time %.1fms below human threshold", stats.MinReactionTime),
			PlayerID:    &player.PlayerID,
			Timestamp:   time.Now().UTC(),
			Evidence: map[string]interface{}{
				"min_reaction_time": stats.MinReactionTime,
				"threshold":         s.thresholds.MinReactionTimeMs,
				"avg_reaction_time": stats.AvgReactionTime,
			},
			Confidence: 0.8,
		})
	}

	// Calculate player score
	report.Score = s.calculatePlayerScore(report)
	if report.Score >= s.thresholds.HighSuspicionScore {
		report.Status = IntegrityStatusSuspicious
	}
	if len(report.Violations) > 0 {
		maxSeverity := s.getMaxSeverity(report.Violations)
		if maxSeverity == SeverityCritical {
			report.Status = IntegrityStatusFlagged
		}
	}

	return report
}

// calculatePlayerStats computes player statistics from events
func (s *ReplayIntegrityService) calculatePlayerStats(data *ReplayAnalysisData, playerID uuid.UUID) PlayerMatchStats {
	stats := PlayerMatchStats{}

	for _, event := range data.Events {
		if event.PlayerID != playerID {
			continue
		}

		switch event.Type {
		case "kill":
			stats.Kills++
			if headshot, ok := event.Data["headshot"].(bool); ok && headshot {
				stats.Headshots++
			}
		case "death":
			stats.Deaths++
		case "assist":
			stats.Assists++
		case "shot":
			stats.ShotsTotal++
			if hit, ok := event.Data["hit"].(bool); ok && hit {
				stats.ShotsHit++
			}
		}
	}

	if stats.Kills > 0 {
		stats.HeadshotPercent = float64(stats.Headshots) / float64(stats.Kills) * 100
	}
	if stats.ShotsTotal > 0 {
		stats.Accuracy = float64(stats.ShotsHit) / float64(stats.ShotsTotal) * 100
	}

	return stats
}

// calculateOverallScore computes overall suspicion score
func (s *ReplayIntegrityService) calculateOverallScore(report *ReplayIntegrityReport) float64 {
	var score float64

	// Weight violations by severity
	for _, v := range report.Violations {
		switch v.Severity {
		case SeverityCritical:
			score += 25 * v.Confidence
		case SeverityHigh:
			score += 15 * v.Confidence
		case SeverityMedium:
			score += 8 * v.Confidence
		case SeverityLow:
			score += 3 * v.Confidence
		}
	}

	// Add player scores
	for _, pr := range report.PlayerReports {
		score += pr.Score * 0.5
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculatePlayerScore computes player suspicion score
func (s *ReplayIntegrityService) calculatePlayerScore(report PlayerIntegrityReport) float64 {
	var score float64

	for _, v := range report.Violations {
		switch v.Severity {
		case SeverityCritical:
			score += 30 * v.Confidence
		case SeverityHigh:
			score += 20 * v.Confidence
		case SeverityMedium:
			score += 10 * v.Confidence
		case SeverityLow:
			score += 5 * v.Confidence
		}
	}

	score += float64(len(report.AnomalyFlags)) * 5

	if score > 100 {
		score = 100
	}

	return score
}

// determineStatus determines overall status based on analysis
func (s *ReplayIntegrityService) determineStatus(report *ReplayIntegrityReport) IntegrityStatus {
	if report.OverallScore >= 80 {
		return IntegrityStatusFlagged
	}
	if report.OverallScore >= 50 {
		return IntegrityStatusSuspicious
	}
	if report.OverallScore >= 20 {
		return IntegrityStatusValid // Valid but worth watching
	}
	if len(report.Violations) == 0 {
		return IntegrityStatusValid
	}
	return IntegrityStatusValid
}

// getMaxSeverity returns the highest severity from violations
func (s *ReplayIntegrityService) getMaxSeverity(violations []IntegrityViolation) ViolationSeverity {
	maxSev := SeverityLow
	for _, v := range violations {
		if v.Severity == SeverityCritical {
			return SeverityCritical
		}
		if v.Severity == SeverityHigh && maxSev != SeverityCritical {
			maxSev = SeverityHigh
		}
		if v.Severity == SeverityMedium && maxSev == SeverityLow {
			maxSev = SeverityMedium
		}
	}
	return maxSev
}

// Helper functions
func calculateAngleDifference(a, b Vector2) float64 {
	pitchDiff := a.Pitch - b.Pitch
	yawDiff := a.Yaw - b.Yaw
	
	// Handle wrap-around for yaw
	if yawDiff > 180 {
		yawDiff -= 360
	} else if yawDiff < -180 {
		yawDiff += 360
	}
	
	return abs(pitchDiff) + abs(yawDiff)
}

func calculateDistance(a, b Vector3) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return sqrt(dx*dx + dy*dy + dz*dz)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

