// Package auth_out defines outbound repository interfaces for authentication
package auth_out

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
)

// EmailVerificationRepository defines persistence operations for email verifications
type EmailVerificationRepository interface {
	// Save creates a new email verification record
	Save(ctx context.Context, verification *auth_entities.EmailVerification) error

	// FindByID retrieves a verification by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*auth_entities.EmailVerification, error)

	// FindByToken retrieves a verification by its token
	FindByToken(ctx context.Context, token string) (*auth_entities.EmailVerification, error)

	// FindByUserID retrieves the latest verification for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.EmailVerification, error)

	// FindPendingByEmail retrieves pending verifications for an email
	FindPendingByEmail(ctx context.Context, email string) (*auth_entities.EmailVerification, error)

	// Update updates an existing verification record
	Update(ctx context.Context, verification *auth_entities.EmailVerification) error

	// InvalidatePreviousVerifications invalidates all previous verifications for a user
	InvalidatePreviousVerifications(ctx context.Context, userID uuid.UUID, email string) error

	// CountRecentAttempts counts verification attempts in the last N minutes
	CountRecentAttempts(ctx context.Context, email string, minutes int) (int, error)
}

// EmailSender defines the interface for sending emails
type EmailSender interface {
	// SendVerificationEmail sends a verification email
	SendVerificationEmail(ctx context.Context, email, token, code string, expiresAt string) error

	// SendMFACode sends an MFA code email
	SendMFACode(ctx context.Context, email, code string, expiresAt string) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, email, token string, expiresAt string) error
}
