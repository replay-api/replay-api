// Package auth_services implements authentication business logic
package auth_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
)

// Configuration constants
const (
	VerificationTokenTTLMinutes = 30
	MFACodeTTLMinutes          = 10
	MaxVerificationsPerHour    = 5
	MaxResendAttempts          = 3
)

// EmailVerificationService implements email verification business logic
type EmailVerificationService struct {
	verificationRepo auth_out.EmailVerificationRepository
	emailSender      auth_out.EmailSender
}

// NewEmailVerificationService creates a new email verification service
func NewEmailVerificationService(
	verificationRepo auth_out.EmailVerificationRepository,
	emailSender auth_out.EmailSender,
) auth_in.EmailVerificationCommand {
	return &EmailVerificationService{
		verificationRepo: verificationRepo,
		emailSender:      emailSender,
	}
}

// SendVerificationEmail sends a verification email to the user
func (s *EmailVerificationService) SendVerificationEmail(ctx context.Context, cmd auth_in.SendVerificationEmailCommand) (*auth_entities.EmailVerification, error) {
	// Rate limit check
	recentAttempts, err := s.verificationRepo.CountRecentAttempts(ctx, cmd.Email, 60)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count recent attempts", "error", err)
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if recentAttempts >= MaxVerificationsPerHour {
		return nil, fmt.Errorf("too many verification attempts. Please try again later")
	}

	// Invalidate any previous pending verifications
	if err := s.verificationRepo.InvalidatePreviousVerifications(ctx, cmd.UserID, cmd.Email); err != nil {
		slog.WarnContext(ctx, "failed to invalidate previous verifications", "error", err)
	}

	// Determine TTL based on type
	ttlMinutes := VerificationTokenTTLMinutes
	if cmd.Type == auth_entities.VerificationTypeMFA {
		ttlMinutes = MFACodeTTLMinutes
	}

	// Create new verification
	verification, err := auth_entities.NewEmailVerification(cmd.UserID, cmd.Email, cmd.Type, ttlMinutes)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification: %w", err)
	}

	// Set request info for audit
	verification.SetRequestInfo(cmd.IPAddress, cmd.UserAgent)

	// Save verification
	if err := s.verificationRepo.Save(ctx, verification); err != nil {
		return nil, fmt.Errorf("failed to save verification: %w", err)
	}

	// Send email
	expiresAtFormatted := verification.ExpiresAt.Format("3:04 PM MST")
	if cmd.Type == auth_entities.VerificationTypeMFA {
		if err := s.emailSender.SendMFACode(ctx, cmd.Email, verification.Code, expiresAtFormatted); err != nil {
			slog.ErrorContext(ctx, "failed to send MFA code", "error", err)
			return nil, fmt.Errorf("failed to send verification email: %w", err)
		}
	} else {
		if err := s.emailSender.SendVerificationEmail(ctx, cmd.Email, verification.Token, verification.Code, expiresAtFormatted); err != nil {
			slog.ErrorContext(ctx, "failed to send verification email", "error", err)
			return nil, fmt.Errorf("failed to send verification email: %w", err)
		}
	}

	slog.InfoContext(ctx, "verification email sent",
		"user_id", cmd.UserID,
		"email", cmd.Email,
		"type", cmd.Type,
		"verification_id", verification.ID)

	return verification, nil
}

// VerifyEmail verifies the email with the provided token or code
func (s *EmailVerificationService) VerifyEmail(ctx context.Context, cmd auth_in.VerifyEmailCommand) (*auth_in.VerificationResult, error) {
	// Find verification by token or code
	var verification *auth_entities.EmailVerification
	var err error

	if cmd.Token != "" {
		verification, err = s.verificationRepo.FindByToken(ctx, cmd.Token)
	} else if cmd.Code != "" {
		// For code verification, we need to look up by a different method
		// This is typically done in context of a session
		return nil, fmt.Errorf("code verification requires a session context")
	} else {
		return nil, fmt.Errorf("token or code is required")
	}

	if err != nil {
		return &auth_in.VerificationResult{
			Success: false,
			Message: "Invalid or expired verification token",
		}, nil
	}

	// Check if already verified
	if verification.IsVerified() {
		return &auth_in.VerificationResult{
			Success:  true,
			Message:  "Email already verified",
			Verified: true,
		}, nil
	}

	// Check if expired
	if verification.IsExpired() {
		verification.Status = auth_entities.VerificationStatusExpired
		_ = s.verificationRepo.Update(ctx, verification)

		return &auth_in.VerificationResult{
			Success: false,
			Message: "Verification token has expired",
		}, nil
	}

	// Attempt verification
	codeOrToken := cmd.Token
	if cmd.Code != "" {
		codeOrToken = cmd.Code
	}

	if verification.Verify(codeOrToken) {
		// Update verification status
		if err := s.verificationRepo.Update(ctx, verification); err != nil {
			slog.ErrorContext(ctx, "failed to update verification", "error", err)
		}

		slog.InfoContext(ctx, "email verified successfully",
			"user_id", verification.UserID,
			"email", verification.Email,
			"verification_id", verification.ID)

		return &auth_in.VerificationResult{
			Success:  true,
			Message:  "Email verified successfully",
			Verified: true,
		}, nil
	}

	// Verification failed
	if err := s.verificationRepo.Update(ctx, verification); err != nil {
		slog.ErrorContext(ctx, "failed to update verification attempts", "error", err)
	}

	return &auth_in.VerificationResult{
		Success:           false,
		Message:           "Invalid verification code",
		RemainingAttempts: verification.RemainingAttempts(),
	}, nil
}

// ResendVerification resends the verification email
func (s *EmailVerificationService) ResendVerification(ctx context.Context, cmd auth_in.ResendVerificationCommand) (*auth_entities.EmailVerification, error) {
	// Check for existing pending verification
	existing, err := s.verificationRepo.FindPendingByEmail(ctx, cmd.Email)
	if err != nil || existing == nil {
		// No existing verification, create new one
		return s.SendVerificationEmail(ctx, auth_in.SendVerificationEmailCommand{
			UserID:    cmd.UserID,
			Email:     cmd.Email,
			Type:      auth_entities.VerificationTypeEmail,
			IPAddress: cmd.IPAddress,
			UserAgent: cmd.UserAgent,
		})
	}

	// Check rate limit for resends
	recentAttempts, err := s.verificationRepo.CountRecentAttempts(ctx, cmd.Email, 60)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if recentAttempts >= MaxResendAttempts {
		return nil, fmt.Errorf("too many resend attempts. Please wait before trying again")
	}

	// Create new verification (invalidates previous)
	return s.SendVerificationEmail(ctx, auth_in.SendVerificationEmailCommand{
		UserID:    cmd.UserID,
		Email:     cmd.Email,
		Type:      existing.Type,
		IPAddress: cmd.IPAddress,
		UserAgent: cmd.UserAgent,
	})
}

// GetVerificationStatus gets the status of a user's email verification
func (s *EmailVerificationService) GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.EmailVerification, error) {
	verification, err := s.verificationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("verification not found: %w", err)
	}

	// Check and update expired status
	if verification.IsExpired() && verification.Status == auth_entities.VerificationStatusPending {
		verification.Status = auth_entities.VerificationStatusExpired
		_ = s.verificationRepo.Update(ctx, verification)
	}

	return verification, nil
}

// CancelVerification cancels a pending verification
func (s *EmailVerificationService) CancelVerification(ctx context.Context, verificationID uuid.UUID) error {
	verification, err := s.verificationRepo.FindByID(ctx, verificationID)
	if err != nil {
		return fmt.Errorf("verification not found: %w", err)
	}

	if !verification.IsPending() {
		return fmt.Errorf("verification is not in pending status")
	}

	verification.Cancel()

	if err := s.verificationRepo.Update(ctx, verification); err != nil {
		return fmt.Errorf("failed to cancel verification: %w", err)
	}

	slog.InfoContext(ctx, "verification cancelled",
		"verification_id", verificationID)

	return nil
}

// Ensure EmailVerificationService implements EmailVerificationCommand
var _ auth_in.EmailVerificationCommand = (*EmailVerificationService)(nil)
