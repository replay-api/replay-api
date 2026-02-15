package cs2

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	handlers "github.com/replay-api/replay-api/pkg/app/cs/handlers"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	e "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type CS2ReplayAdapter struct {
	config ProcessingConfig
	logger *slog.Logger
}

// CS2ReplayAdapterOption is a functional option for configuring the adapter
type CS2ReplayAdapterOption func(*CS2ReplayAdapter)

// WithProcessingConfig sets the processing configuration
func WithProcessingConfig(config ProcessingConfig) CS2ReplayAdapterOption {
	return func(a *CS2ReplayAdapter) {
		a.config = config
	}
}

// WithLogger sets the logger
func WithLogger(logger *slog.Logger) CS2ReplayAdapterOption {
	return func(a *CS2ReplayAdapter) {
		a.logger = logger
	}
}

func NewCS2ReplayAdapter(opts ...CS2ReplayAdapterOption) *CS2ReplayAdapter {
	adapter := &CS2ReplayAdapter{
		config: DefaultProcessingConfig(),
		logger: slog.Default(),
	}
	
	for _, opt := range opts {
		opt(adapter)
	}
	
	return adapter
}

func registerParsers(p dem.Parser, matchContext *state.CS2MatchContext, eventsChan chan *e.GameEvent, robustParser *RobustParser) {
	// Register all event handlers - they include their own panic handling via defer/recover
	// The robust parser tracks errors but handlers manage their own recovery
	p.RegisterEventHandler(handlers.BeginNewMatch(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.WeaponFire(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.HitEvent(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.RoundMVP(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.Kill(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ClutchStart(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ClutchProgress(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ClutchEnd(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.PlayerFlashed(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.BombPlanted(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.BombDefused(p, matchContext, eventsChan))
	// Grenade events for heatmaps
	p.RegisterEventHandler(handlers.HeGrenadeExplode(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.FlashExplode(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.SmokeStart(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.InfernoStart(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.DecoyStart(p, matchContext, eventsChan))
	// CS2/Source2 reliable grenade handler - uses GrenadeProjectileDestroy which works for all grenades
	p.RegisterEventHandler(handlers.GrenadeProjectileDestroy(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.GrenadeProjectileThrow(p, matchContext, eventsChan)) // Debug
	p.RegisterEventHandler(handlers.GenericGameEvent(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.RoundEnd(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.AnnouncementWinPanelMatch(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.GameHalfEnded(p, matchContext, eventsChan))
	p.RegisterEventHandler(handlers.ScoreUpdated(p, matchContext, eventsChan))
}

// Parse parses a demo file and sends events to the channel
func (c *CS2ReplayAdapter) Parse(ctx context.Context, matchID uuid.UUID, content io.Reader, eventsChan chan *e.GameEvent) error {
	startTime := time.Now()
	robustParser := NewRobustParser(c.config, c.logger)
	
	c.logger.Info("Starting demo file parsing",
		"matchID", matchID,
		"config", fmt.Sprintf("maxErrors=%d, maxErrorRate=%.2f", c.config.MaxRecoverableErrors, c.config.MaxErrorRate),
	)
	
	matchContext := state.NewCS2MatchContext(ctx, matchID)
	parser := dem.NewParser(content)
	defer parser.Close()
	defer close(eventsChan)

	registerParsers(parser, matchContext, eventsChan, robustParser)

	// Parse with retry logic for transient failures
	err := ParseWithRetry(ctx, parser, 2) // Up to 2 retries

	// Calculate processing duration
	duration := time.Since(startTime)
	stats := robustParser.GetStats()
	stats.EndTime = time.Now()

	if err != nil {
		// Check if this is a context cancellation
		if ctx.Err() != nil {
			c.logger.WarnContext(ctx, "Parsing canceled by context",
				"matchID", matchID,
				"duration", duration,
				"errorStats", stats.Summary(),
			)
			return &ProcessingError{
				Code:        ErrCodeContextCanceled,
				Message:     "context canceled during parsing",
				Cause:       ctx.Err(),
				Recoverable: false,
				Timestamp:   time.Now(),
			}
		}
		
		// Wrap the error with our processing error type
		procErr := &ProcessingError{
			Code:        categorizeParsingError(err),
			Message:     "demo parsing failed",
			Cause:       err,
			Recoverable: false,
			Timestamp:   time.Now(),
		}
		
		c.logger.ErrorContext(ctx, "Failed to parse demo",
			"matchID", matchID,
			"duration", duration,
			"errorStats", stats.Summary(),
			"error", err,
		)
		
		return procErr
	}

	// Log success with statistics
	if stats.TotalErrors > 0 {
		c.logger.WarnContext(ctx, "Demo parsed with recoverable errors",
			"matchID", matchID,
			"duration", duration,
			"errorStats", stats.Summary(),
		)
	} else {
		c.logger.InfoContext(ctx, "Demo parsed successfully",
			"matchID", matchID,
			"duration", duration,
		)
	}

	return nil
}

// ParseWithResult parses and returns detailed result
func (c *CS2ReplayAdapter) ParseWithResult(ctx context.Context, matchID uuid.UUID, content io.Reader, eventsChan chan *e.GameEvent) (*ProcessingResult, error) {
	startTime := time.Now()
	robustParser := NewRobustParser(c.config, c.logger)
	
	matchContext := state.NewCS2MatchContext(ctx, matchID)
	parser := dem.NewParser(content)
	defer parser.Close()
	defer close(eventsChan)

	registerParsers(parser, matchContext, eventsChan, robustParser)

	err := ParseWithRetry(ctx, parser, 2)
	
	duration := time.Since(startTime)
	stats := robustParser.GetStats()
	stats.EndTime = time.Now()

	result := &ProcessingResult{
		MatchID:    matchID,
		Success:    err == nil,
		Duration:   duration,
		ErrorStats: stats,
	}

	if stats.HasFatalError() || err != nil {
		result.Success = false
	}

	if stats.TotalErrors > 0 {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Processed with %d recoverable errors", stats.RecoverableErrors))
	}

	return result, err
}

// categorizeParsingError categorizes the parsing error type
func categorizeParsingError(err error) ErrorCode {
	if err == nil {
		return ""
	}
	
	errStr := err.Error()
	
	// Check for common error patterns
	switch {
	case contains(errStr, "unexpected EOF"):
		return ErrCodeFileCorrupted
	case contains(errStr, "invalid"):
		return ErrCodeFileCorrupted
	case contains(errStr, "unsupported"):
		return ErrCodeUnsupportedFormat
	case contains(errStr, "io"):
		return ErrCodeIOFailure
	default:
		return ErrCodeFileCorrupted
	}
}
