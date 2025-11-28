// Package auth_in defines inbound port interfaces for authentication operations
package auth_in

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
)

// SendVerificationEmailCommand represents a request to send verification email
type SendVerificationEmailCommand struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Type      auth_entities.VerificationType `json:"type"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// VerifyEmailCommand represents a request to verify an email
type VerifyEmailCommand struct {
	Token     string `json:"token,omitempty"`
	Code      string `json:"code,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// ResendVerificationCommand represents a request to resend verification
type ResendVerificationCommand struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// VerificationResult represents the result of a verification attempt
type VerificationResult struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	RemainingAttempts int    `json:"remaining_attempts,omitempty"`
	Verified          bool   `json:"verified"`
}

// EmailVerificationCommand defines the inbound port for email verification operations
type EmailVerificationCommand interface {
	// SendVerificationEmail sends a verification email to the user
	SendVerificationEmail(ctx context.Context, cmd SendVerificationEmailCommand) (*auth_entities.EmailVerification, error)

	// VerifyEmail verifies the email with the provided token or code
	VerifyEmail(ctx context.Context, cmd VerifyEmailCommand) (*VerificationResult, error)

	// ResendVerification resends the verification email
	ResendVerification(ctx context.Context, cmd ResendVerificationCommand) (*auth_entities.EmailVerification, error)

	// GetVerificationStatus gets the status of a user's email verification
	GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.EmailVerification, error)

	// CancelVerification cancels a pending verification
	CancelVerification(ctx context.Context, verificationID uuid.UUID) error
}
