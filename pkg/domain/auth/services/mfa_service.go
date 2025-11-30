// Package auth_services implements authentication business logic
package auth_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	_ "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
)

// MFAService implements MFA business logic
type MFAService struct {
	verificationService auth_in.EmailVerificationCommand
	mfaSettingsRepo     MFASettingsRepository
	mfaSessionRepo      MFASessionRepository
}

// MFASettingsRepository defines persistence for MFA settings
type MFASettingsRepository interface {
	Save(ctx context.Context, settings *auth_entities.MFASettings) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.MFASettings, error)
	Update(ctx context.Context, settings *auth_entities.MFASettings) error
}

// MFASessionRepository defines persistence for MFA sessions
type MFASessionRepository interface {
	Save(ctx context.Context, session *auth_entities.MFASession) error
	FindByID(ctx context.Context, id uuid.UUID) (*auth_entities.MFASession, error)
	FindPendingByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.MFASession, error)
	Update(ctx context.Context, session *auth_entities.MFASession) error
	InvalidatePreviousSessions(ctx context.Context, userID uuid.UUID) error
}

// MFACommand defines the MFA command interface
type MFACommand interface {
	// EnableMFA enables MFA for a user
	EnableMFA(ctx context.Context, userID uuid.UUID, method auth_entities.MFAMethod) (*auth_entities.MFASettings, error)

	// DisableMFA disables MFA for a user (requires verification)
	DisableMFA(ctx context.Context, userID uuid.UUID) error

	// GetMFASettings gets MFA settings for a user
	GetMFASettings(ctx context.Context, userID uuid.UUID) (*auth_entities.MFASettings, error)

	// InitiateMFA starts an MFA challenge
	InitiateMFA(ctx context.Context, userID uuid.UUID, action string, actionData map[string]any, ipAddress, userAgent string) (*auth_entities.MFASession, error)

	// VerifyMFA verifies an MFA code
	VerifyMFA(ctx context.Context, sessionID uuid.UUID, code string, ipAddress, userAgent string) (*MFAVerificationResult, error)

	// GetMFASession gets an MFA session
	GetMFASession(ctx context.Context, sessionID uuid.UUID) (*auth_entities.MFASession, error)

	// GenerateRecoveryCodes generates new recovery codes
	GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error)

	// UseRecoveryCode uses a recovery code to bypass MFA
	UseRecoveryCode(ctx context.Context, sessionID uuid.UUID, code string) (*MFAVerificationResult, error)
}

// MFAVerificationResult represents the result of MFA verification
type MFAVerificationResult struct {
	Success           bool                       `json:"success"`
	Message           string                     `json:"message"`
	Session           *auth_entities.MFASession  `json:"session,omitempty"`
	RemainingAttempts int                        `json:"remaining_attempts,omitempty"`
}

// NewMFAService creates a new MFA service
func NewMFAService(
	verificationService auth_in.EmailVerificationCommand,
	mfaSettingsRepo MFASettingsRepository,
	mfaSessionRepo MFASessionRepository,
) MFACommand {
	return &MFAService{
		verificationService: verificationService,
		mfaSettingsRepo:     mfaSettingsRepo,
		mfaSessionRepo:      mfaSessionRepo,
	}
}

// EnableMFA enables MFA for a user
func (s *MFAService) EnableMFA(ctx context.Context, userID uuid.UUID, method auth_entities.MFAMethod) (*auth_entities.MFASettings, error) {
	// Get or create MFA settings
	settings, err := s.mfaSettingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		// Create new settings
		settings = auth_entities.NewMFASettings(userID)
	}

	// Enable the specified method
	settings.EnableMFA(method)

	// Save settings
	if settings.ID == uuid.Nil {
		if err := s.mfaSettingsRepo.Save(ctx, settings); err != nil {
			return nil, fmt.Errorf("failed to save MFA settings: %w", err)
		}
	} else {
		if err := s.mfaSettingsRepo.Update(ctx, settings); err != nil {
			return nil, fmt.Errorf("failed to update MFA settings: %w", err)
		}
	}

	slog.InfoContext(ctx, "MFA enabled",
		"user_id", userID,
		"method", method)

	return settings, nil
}

// DisableMFA disables MFA for a user
func (s *MFAService) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	settings, err := s.mfaSettingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("MFA not configured for user")
	}

	settings.DisableMFA()

	if err := s.mfaSettingsRepo.Update(ctx, settings); err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	slog.InfoContext(ctx, "MFA disabled", "user_id", userID)

	return nil
}

// GetMFASettings gets MFA settings for a user
func (s *MFAService) GetMFASettings(ctx context.Context, userID uuid.UUID) (*auth_entities.MFASettings, error) {
	settings, err := s.mfaSettingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		// Return default disabled settings
		return auth_entities.NewMFASettings(userID), nil
	}
	return settings, nil
}

// InitiateMFA starts an MFA challenge
func (s *MFAService) InitiateMFA(ctx context.Context, userID uuid.UUID, action string, actionData map[string]any, ipAddress, userAgent string) (*auth_entities.MFASession, error) {
	// Get MFA settings
	settings, err := s.mfaSettingsRepo.FindByUserID(ctx, userID)
	if err != nil || !settings.Enabled {
		return nil, fmt.Errorf("MFA not enabled for user")
	}

	// Invalidate previous pending sessions
	_ = s.mfaSessionRepo.InvalidatePreviousSessions(ctx, userID)

	// Send verification code via email
	verification, err := s.verificationService.SendVerificationEmail(ctx, auth_in.SendVerificationEmailCommand{
		UserID:    userID,
		Email:     settings.BackupEmail, // Use backup email or primary
		Type:      auth_entities.VerificationTypeMFA,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send MFA code: %w", err)
	}

	// Create MFA session
	session := auth_entities.NewMFASession(userID, verification.ID, settings.PrimaryMethod, MFACodeTTLMinutes)
	session.SetRequestInfo(ipAddress, userAgent)
	session.SetPendingAction(action, actionData)

	if err := s.mfaSessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save MFA session: %w", err)
	}

	slog.InfoContext(ctx, "MFA challenge initiated",
		"user_id", userID,
		"session_id", session.ID,
		"action", action)

	return session, nil
}

// VerifyMFA verifies an MFA code
func (s *MFAService) VerifyMFA(ctx context.Context, sessionID uuid.UUID, code string, ipAddress, userAgent string) (*MFAVerificationResult, error) {
	// Get session
	session, err := s.mfaSessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return &MFAVerificationResult{
			Success: false,
			Message: "Invalid MFA session",
		}, nil
	}

	// Check session status
	if !session.IsPending() {
		if session.IsExpired() {
			return &MFAVerificationResult{
				Success: false,
				Message: "MFA session has expired",
			}, nil
		}
		if session.IsVerified() {
			return &MFAVerificationResult{
				Success: true,
				Message: "MFA already verified",
				Session: session,
			}, nil
		}
		return &MFAVerificationResult{
			Success: false,
			Message: "MFA session is no longer valid",
		}, nil
	}

	// Verify the code through email verification service
	result, err := s.verificationService.VerifyEmail(ctx, auth_in.VerifyEmailCommand{
		Code:      code,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
	if err != nil {
		session.IncrementAttempts()
		_ = s.mfaSessionRepo.Update(ctx, session)

		return &MFAVerificationResult{
			Success:           false,
			Message:           "Verification failed",
			RemainingAttempts: session.RemainingAttempts(),
		}, nil
	}

	if result.Verified {
		session.MarkVerified()
		if err := s.mfaSessionRepo.Update(ctx, session); err != nil {
			slog.ErrorContext(ctx, "failed to update MFA session", "error", err)
		}

		slog.InfoContext(ctx, "MFA verified successfully",
			"user_id", session.UserID,
			"session_id", session.ID)

		return &MFAVerificationResult{
			Success: true,
			Message: "MFA verification successful",
			Session: session,
		}, nil
	}

	session.IncrementAttempts()
	_ = s.mfaSessionRepo.Update(ctx, session)

	return &MFAVerificationResult{
		Success:           false,
		Message:           result.Message,
		RemainingAttempts: session.RemainingAttempts(),
	}, nil
}

// GetMFASession gets an MFA session
func (s *MFAService) GetMFASession(ctx context.Context, sessionID uuid.UUID) (*auth_entities.MFASession, error) {
	return s.mfaSessionRepo.FindByID(ctx, sessionID)
}

// GenerateRecoveryCodes generates new recovery codes
func (s *MFAService) GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	settings, err := s.mfaSettingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("MFA not configured for user")
	}

	codes, err := settings.GenerateRecoveryCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	if err := s.mfaSettingsRepo.Update(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to save recovery codes: %w", err)
	}

	slog.InfoContext(ctx, "Recovery codes generated", "user_id", userID)

	return codes, nil
}

// UseRecoveryCode uses a recovery code to bypass MFA
func (s *MFAService) UseRecoveryCode(ctx context.Context, sessionID uuid.UUID, code string) (*MFAVerificationResult, error) {
	session, err := s.mfaSessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return &MFAVerificationResult{
			Success: false,
			Message: "Invalid MFA session",
		}, nil
	}

	if !session.IsPending() {
		return &MFAVerificationResult{
			Success: false,
			Message: "MFA session is no longer valid",
		}, nil
	}

	settings, err := s.mfaSettingsRepo.FindByUserID(ctx, session.UserID)
	if err != nil {
		return &MFAVerificationResult{
			Success: false,
			Message: "MFA settings not found",
		}, nil
	}

	if settings.UseRecoveryCode(code) {
		session.MarkVerified()
		_ = s.mfaSessionRepo.Update(ctx, session)
		_ = s.mfaSettingsRepo.Update(ctx, settings)

		slog.WarnContext(ctx, "Recovery code used",
			"user_id", session.UserID,
			"remaining_codes", len(settings.RecoveryCodes))

		return &MFAVerificationResult{
			Success: true,
			Message: "MFA bypassed with recovery code",
			Session: session,
		}, nil
	}

	return &MFAVerificationResult{
		Success: false,
		Message: "Invalid recovery code",
	}, nil
}

// Ensure MFAService implements MFACommand
var _ MFACommand = (*MFAService)(nil)
