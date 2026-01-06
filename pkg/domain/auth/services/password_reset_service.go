// Package auth_services implements authentication business logic
package auth_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
)

// Password reset configuration constants
const (
	PasswordResetTokenTTLMinutes = 60 // 1 hour
	MaxPasswordResetsPerHour     = 3
)

// PasswordResetService implements password reset business logic
type PasswordResetService struct {
	resetRepo      auth_out.PasswordResetRepository
	emailUserRepo  email_out.EmailUserReader
	emailUserWriter email_out.EmailUserWriter
	passwordHasher email_out.PasswordHasher
	emailSender    auth_out.EmailSender
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(
	resetRepo auth_out.PasswordResetRepository,
	emailUserRepo email_out.EmailUserReader,
	emailUserWriter email_out.EmailUserWriter,
	passwordHasher email_out.PasswordHasher,
	emailSender auth_out.EmailSender,
) auth_in.PasswordResetCommand {
	return &PasswordResetService{
		resetRepo:       resetRepo,
		emailUserRepo:   emailUserRepo,
		emailUserWriter: emailUserWriter,
		passwordHasher:  passwordHasher,
		emailSender:     emailSender,
	}
}

// RequestPasswordReset initiates a password reset and sends email
func (s *PasswordResetService) RequestPasswordReset(ctx context.Context, cmd auth_in.RequestPasswordResetCommand) (*auth_entities.PasswordReset, error) {
	// Rate limit check
	recentAttempts, err := s.resetRepo.CountRecentAttempts(ctx, cmd.Email, 60)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count recent attempts", "error", err)
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if recentAttempts >= MaxPasswordResetsPerHour {
		return nil, fmt.Errorf("too many password reset attempts. Please try again later")
	}

	// Find the email user
	search := s.newSearchByEmail(ctx, cmd.Email)
	emailUsers, err := s.emailUserRepo.Search(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search for email user", "error", err)
		// Don't reveal whether email exists
		return nil, nil
	}

	// If no user found, silently return (don't reveal email doesn't exist)
	if len(emailUsers) == 0 {
		slog.InfoContext(ctx, "password reset requested for non-existent email", "email", cmd.Email)
		// Return nil without error to not reveal email doesn't exist
		return nil, nil
	}

	emailUser := &emailUsers[0]

	// Invalidate any previous pending reset requests
	if err := s.resetRepo.InvalidatePreviousResets(ctx, emailUser.ID, cmd.Email); err != nil {
		slog.WarnContext(ctx, "failed to invalidate previous resets", "error", err)
	}

	// Create new password reset
	reset, err := auth_entities.NewPasswordReset(emailUser.ID, cmd.Email, PasswordResetTokenTTLMinutes)
	if err != nil {
		return nil, fmt.Errorf("failed to create password reset: %w", err)
	}

	// Set request info for audit
	reset.SetRequestInfo(cmd.IPAddress, cmd.UserAgent)

	// Save reset request
	if err := s.resetRepo.Save(ctx, reset); err != nil {
		return nil, fmt.Errorf("failed to save password reset: %w", err)
	}

	// Send email
	expiresAtFormatted := reset.ExpiresAt.Format("3:04 PM MST")
	if err := s.emailSender.SendPasswordResetEmail(ctx, cmd.Email, reset.Token, expiresAtFormatted); err != nil {
		slog.ErrorContext(ctx, "failed to send password reset email", "error", err)
		return nil, fmt.Errorf("failed to send password reset email: %w", err)
	}

	slog.InfoContext(ctx, "password reset email sent",
		"user_id", emailUser.ID,
		"email", cmd.Email,
		"reset_id", reset.ID)

	return reset, nil
}

// ConfirmPasswordReset completes the password reset with new password
func (s *PasswordResetService) ConfirmPasswordReset(ctx context.Context, cmd auth_in.ConfirmPasswordResetCommand) (*auth_in.PasswordResetResult, error) {
	// Validate password strength
	if len(cmd.NewPassword) < 8 {
		return &auth_in.PasswordResetResult{
			Success: false,
			Message: "Password must be at least 8 characters long",
		}, nil
	}

	// Find reset request by token
	reset, err := s.resetRepo.FindByToken(ctx, cmd.Token)
	if err != nil {
		return &auth_in.PasswordResetResult{
			Success: false,
			Message: "Invalid or expired reset token",
		}, nil
	}

	// Check if already used
	if reset.IsUsed() {
		return &auth_in.PasswordResetResult{
			Success: false,
			Message: "This reset link has already been used",
		}, nil
	}

	// Check if expired
	if reset.IsExpired() {
		reset.MarkAsExpired()
		_ = s.resetRepo.Update(ctx, reset)

		return &auth_in.PasswordResetResult{
			Success: false,
			Message: "Password reset link has expired",
		}, nil
	}

	// Hash the new password
	hashedPassword, err := s.passwordHasher.HashPassword(ctx, cmd.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new password", "error", err)
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// Update the user's password
	// Find the email user
	search := s.newSearchByEmail(ctx, reset.Email)
	emailUsers, err := s.emailUserRepo.Search(ctx, search)
	if err != nil || len(emailUsers) == 0 {
		slog.ErrorContext(ctx, "failed to find user for password reset", "error", err)
		return nil, fmt.Errorf("failed to update password: user not found")
	}

	emailUser := emailUsers[0]
	emailUser.PasswordHash = hashedPassword

	// Update password via writer
	_, err = s.emailUserWriter.Create(ctx, &emailUser) // This is actually an upsert
	if err != nil {
		slog.ErrorContext(ctx, "failed to update user password", "error", err)
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// Mark reset as used
	reset.MarkAsUsed()
	if err := s.resetRepo.Update(ctx, reset); err != nil {
		slog.WarnContext(ctx, "failed to update reset status", "error", err)
	}

	slog.InfoContext(ctx, "password reset completed",
		"user_id", reset.UserID,
		"email", reset.Email,
		"reset_id", reset.ID)

	return &auth_in.PasswordResetResult{
		Success: true,
		Message: "Password has been reset successfully",
	}, nil
}

// ValidateResetToken validates a reset token without using it
func (s *PasswordResetService) ValidateResetToken(ctx context.Context, token string) (*auth_entities.PasswordReset, error) {
	reset, err := s.resetRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if reset.IsUsed() {
		return nil, fmt.Errorf("reset token has already been used")
	}

	if reset.IsExpired() {
		return nil, fmt.Errorf("reset token has expired")
	}

	return reset, nil
}

// GetResetStatus gets the status of a password reset by user ID
func (s *PasswordResetService) GetResetStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.PasswordReset, error) {
	reset, err := s.resetRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("reset not found: %w", err)
	}

	// Update expired status
	if reset.IsExpired() && reset.Status == auth_entities.PasswordResetStatusPending {
		reset.MarkAsExpired()
		_ = s.resetRepo.Update(ctx, reset)
	}

	return reset, nil
}

// CancelPasswordReset cancels a pending password reset
func (s *PasswordResetService) CancelPasswordReset(ctx context.Context, resetID uuid.UUID) error {
	reset, err := s.resetRepo.FindByID(ctx, resetID)
	if err != nil {
		return fmt.Errorf("reset not found: %w", err)
	}

	if !reset.IsPending() {
		return fmt.Errorf("reset is not in pending status")
	}

	reset.Cancel()

	if err := s.resetRepo.Update(ctx, reset); err != nil {
		return fmt.Errorf("failed to cancel reset: %w", err)
	}

	slog.InfoContext(ctx, "password reset cancelled", "reset_id", resetID)

	return nil
}

// newSearchByEmail creates a search query for email users
func (s *PasswordResetService) newSearchByEmail(ctx context.Context, emailString string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field: "Email",
							Values: []interface{}{
								emailString,
							},
						},
					},
				},
			},
		},
	}

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

// Ensure PasswordResetService implements PasswordResetCommand
var _ auth_in.PasswordResetCommand = (*PasswordResetService)(nil)

