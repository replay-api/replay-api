package cs2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
)

// ProcessingError represents categorized errors during replay processing
type ProcessingError struct {
	Code      ErrorCode
	Message   string
	Cause     error
	Tick      int
	Handler   string
	Recoverable bool
	Timestamp time.Time
}

func (e *ProcessingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s at tick %d (%s): %v", e.Code, e.Message, e.Tick, e.Handler, e.Cause)
	}
	return fmt.Sprintf("[%s] %s at tick %d (%s)", e.Code, e.Message, e.Tick, e.Handler)
}

func (e *ProcessingError) Unwrap() error {
	return e.Cause
}

// ErrorCode defines categories of processing errors
type ErrorCode string

const (
	// Fatal errors - processing cannot continue
	ErrCodeFileCorrupted    ErrorCode = "FILE_CORRUPTED"
	ErrCodeUnsupportedFormat ErrorCode = "UNSUPPORTED_FORMAT"
	ErrCodeIOFailure         ErrorCode = "IO_FAILURE"
	ErrCodeContextCanceled   ErrorCode = "CONTEXT_CANCELED"
	
	// Recoverable errors - processing can continue with degraded data
	ErrCodeMalformedEvent    ErrorCode = "MALFORMED_EVENT"
	ErrCodeMissingPlayer     ErrorCode = "MISSING_PLAYER"
	ErrCodeInvalidState      ErrorCode = "INVALID_STATE"
	ErrCodeHandlerPanic      ErrorCode = "HANDLER_PANIC"
	ErrCodeDataInconsistency ErrorCode = "DATA_INCONSISTENCY"
)

// ErrorStats tracks error statistics during processing
type ErrorStats struct {
	mu sync.RWMutex
	
	TotalErrors       int
	RecoverableErrors int
	FatalErrors       int
	ErrorsByCode      map[ErrorCode]int
	ErrorsByHandler   map[string]int
	FirstError        *ProcessingError
	LastError         *ProcessingError
	StartTime         time.Time
	EndTime           time.Time
}

// NewErrorStats creates a new error statistics tracker
func NewErrorStats() *ErrorStats {
	return &ErrorStats{
		ErrorsByCode:    make(map[ErrorCode]int),
		ErrorsByHandler: make(map[string]int),
		StartTime:       time.Now(),
	}
}

// Record adds an error to the statistics
func (s *ErrorStats) Record(err *ProcessingError) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.TotalErrors++
	if err.Recoverable {
		s.RecoverableErrors++
	} else {
		s.FatalErrors++
	}
	
	s.ErrorsByCode[err.Code]++
	if err.Handler != "" {
		s.ErrorsByHandler[err.Handler]++
	}
	
	if s.FirstError == nil {
		s.FirstError = err
	}
	s.LastError = err
}

// Summary returns a summary of the error statistics
func (s *ErrorStats) Summary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return fmt.Sprintf(
		"Total: %d (Recoverable: %d, Fatal: %d), Duration: %s",
		s.TotalErrors, s.RecoverableErrors, s.FatalErrors,
		time.Since(s.StartTime).Round(time.Millisecond),
	)
}

// HasFatalError returns true if any fatal error occurred
func (s *ErrorStats) HasFatalError() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FatalErrors > 0
}

// GetErrorRate returns the error rate (errors per second)
func (s *ErrorStats) GetErrorRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	duration := time.Since(s.StartTime).Seconds()
	if duration == 0 {
		return 0
	}
	return float64(s.TotalErrors) / duration
}

// ProcessingConfig defines configuration for replay processing
type ProcessingConfig struct {
	// MaxRecoverableErrors is the maximum number of recoverable errors before aborting
	MaxRecoverableErrors int
	
	// MaxErrorRate is the maximum errors per second before aborting
	MaxErrorRate float64
	
	// EnablePanicRecovery enables recovery from handler panics
	EnablePanicRecovery bool
	
	// SkipMalformedEvents skips events that cannot be parsed
	SkipMalformedEvents bool
	
	// LogLevel sets the logging verbosity for errors
	LogLevel slog.Level
}

// DefaultProcessingConfig returns sensible defaults for processing
func DefaultProcessingConfig() ProcessingConfig {
	return ProcessingConfig{
		MaxRecoverableErrors: 100,      // Allow up to 100 recoverable errors
		MaxErrorRate:         10.0,     // Max 10 errors per second
		EnablePanicRecovery:  true,     // Recover from panics
		SkipMalformedEvents:  true,     // Skip malformed events
		LogLevel:             slog.LevelWarn,
	}
}

// StrictProcessingConfig returns config for strict processing (fail on any error)
func StrictProcessingConfig() ProcessingConfig {
	return ProcessingConfig{
		MaxRecoverableErrors: 0,
		MaxErrorRate:         0,
		EnablePanicRecovery:  true,
		SkipMalformedEvents:  false,
		LogLevel:             slog.LevelDebug,
	}
}

// ProcessingResult contains the result of replay processing
type ProcessingResult struct {
	MatchID     uuid.UUID
	Success     bool
	Events      int
	Duration    time.Duration
	ErrorStats  *ErrorStats
	Warnings    []string
}

// RobustParser wraps the demo parser with error handling
type RobustParser struct {
	config     ProcessingConfig
	errorStats *ErrorStats
	logger     *slog.Logger
}

// NewRobustParser creates a new robust parser with the given config
func NewRobustParser(config ProcessingConfig, logger *slog.Logger) *RobustParser {
	if logger == nil {
		logger = slog.Default()
	}
	return &RobustParser{
		config:     config,
		errorStats: NewErrorStats(),
		logger:     logger,
	}
}

// CheckErrorThreshold checks if error thresholds have been exceeded
func (rp *RobustParser) CheckErrorThreshold() error {
	if rp.config.MaxRecoverableErrors > 0 && 
		rp.errorStats.RecoverableErrors >= rp.config.MaxRecoverableErrors {
		return &ProcessingError{
			Code:        ErrCodeDataInconsistency,
			Message:     fmt.Sprintf("exceeded maximum recoverable errors (%d)", rp.config.MaxRecoverableErrors),
			Recoverable: false,
			Timestamp:   time.Now(),
		}
	}
	
	if rp.config.MaxErrorRate > 0 && rp.errorStats.GetErrorRate() > rp.config.MaxErrorRate {
		return &ProcessingError{
			Code:        ErrCodeDataInconsistency,
			Message:     fmt.Sprintf("error rate exceeded (%.2f/s > %.2f/s)", rp.errorStats.GetErrorRate(), rp.config.MaxErrorRate),
			Recoverable: false,
			Timestamp:   time.Now(),
		}
	}
	
	return nil
}

// RecordError records an error and checks thresholds
func (rp *RobustParser) RecordError(err *ProcessingError) error {
	rp.errorStats.Record(err)
	
	if err.Recoverable {
		rp.logger.Log(context.Background(), rp.config.LogLevel, "Recoverable error during parsing",
			"code", err.Code,
			"message", err.Message,
			"tick", err.Tick,
			"handler", err.Handler,
		)
	} else {
		rp.logger.Error("Fatal error during parsing",
			"code", err.Code,
			"message", err.Message,
			"tick", err.Tick,
			"handler", err.Handler,
			"error", err.Cause,
		)
	}
	
	return rp.CheckErrorThreshold()
}

// GetStats returns the current error statistics
func (rp *RobustParser) GetStats() *ErrorStats {
	return rp.errorStats
}

// WrapHandler wraps an event handler with panic recovery
func WrapHandler[T any](rp *RobustParser, handlerName string, handler func(T)) func(T) {
	if !rp.config.EnablePanicRecovery {
		return handler
	}
	
	return func(event T) {
		defer func() {
			if r := recover(); r != nil {
				err := &ProcessingError{
					Code:        ErrCodeHandlerPanic,
					Message:     fmt.Sprintf("panic in handler: %v", r),
					Handler:     handlerName,
					Recoverable: true,
					Timestamp:   time.Now(),
				}
				rp.RecordError(err)
			}
		}()
		handler(event)
	}
}

// ValidationError creates a data validation error
func ValidationError(handler string, tick int, message string) *ProcessingError {
	return &ProcessingError{
		Code:        ErrCodeMalformedEvent,
		Message:     message,
		Handler:     handler,
		Tick:        tick,
		Recoverable: true,
		Timestamp:   time.Now(),
	}
}

// MissingPlayerError creates a missing player error
func MissingPlayerError(handler string, tick int, playerID string) *ProcessingError {
	return &ProcessingError{
		Code:        ErrCodeMissingPlayer,
		Message:     fmt.Sprintf("player not found: %s", playerID),
		Handler:     handler,
		Tick:        tick,
		Recoverable: true,
		Timestamp:   time.Now(),
	}
}

// InvalidStateError creates an invalid state error
func InvalidStateError(handler string, tick int, message string) *ProcessingError {
	return &ProcessingError{
		Code:        ErrCodeInvalidState,
		Message:     message,
		Handler:     handler,
		Tick:        tick,
		Recoverable: true,
		Timestamp:   time.Now(),
	}
}

// IOError creates an I/O error (usually fatal)
func IOError(cause error) *ProcessingError {
	return &ProcessingError{
		Code:        ErrCodeIOFailure,
		Message:     "I/O operation failed",
		Cause:       cause,
		Recoverable: false,
		Timestamp:   time.Now(),
	}
}

// CorruptedFileError creates a corrupted file error (fatal)
func CorruptedFileError(cause error, tick int) *ProcessingError {
	return &ProcessingError{
		Code:        ErrCodeFileCorrupted,
		Message:     "demo file is corrupted",
		Cause:       cause,
		Tick:        tick,
		Recoverable: false,
		Timestamp:   time.Now(),
	}
}

// ParseWithRetry attempts to parse with retry logic for transient failures
func ParseWithRetry(ctx context.Context, parser dem.Parser, maxRetries int) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return &ProcessingError{
				Code:        ErrCodeContextCanceled,
				Message:     "context canceled",
				Cause:       ctx.Err(),
				Recoverable: false,
				Timestamp:   time.Now(),
			}
		}
		
		err := parser.ParseToEnd()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if isRetryableError(err) {
			slog.WarnContext(ctx, "Retryable parsing error, attempting retry",
				"attempt", attempt+1,
				"maxRetries", maxRetries,
				"error", err,
			)
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
			continue
		}
		
		// Non-retryable error
		break
	}
	
	return lastErr
}

// isRetryableError determines if an error is transient and worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common transient errors
	errStr := err.Error()
	
	// IO timeout errors
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	
	// Temporary I/O errors
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return false // File is likely corrupted, not transient
	}
	
	// Check error message patterns for retryable scenarios
	retryablePatterns := []string{
		"temporary",
		"timeout",
		"retry",
	}
	
	for _, pattern := range retryablePatterns {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}
	
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
			contains(toLower(s), toLower(substr)))
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// SafeGetPlayerName safely gets player name with fallback
func SafeGetPlayerName(player interface{ Name() string }, fallback string) string {
	defer func() {
		recover() // Ignore panics
	}()
	
	if player == nil {
		return fallback
	}
	
	name := player.Name()
	if name == "" {
		return fallback
	}
	return name
}

// SafeGetPlayerSteamID safely gets player Steam ID with fallback
func SafeGetPlayerSteamID(player interface{ SteamID64() uint64 }) uint64 {
	defer func() {
		recover() // Ignore panics
	}()
	
	if player == nil {
		return 0
	}
	
	return player.SteamID64()
}

// SafeGetPlayerTeam safely gets player team
func SafeGetPlayerTeam(player interface{ Team() int }) int {
	defer func() {
		recover() // Ignore panics
	}()
	
	if player == nil {
		return 0
	}
	
	return player.Team()
}
