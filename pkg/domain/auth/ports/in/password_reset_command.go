// Package auth_in defines inbound port interfaces for authentication operations
package auth_in

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
)

// RequestPasswordResetCommand represents a request to initiate password reset
type RequestPasswordResetCommand struct {
	Email     string `json:"email"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// ConfirmPasswordResetCommand represents a request to complete password reset
type ConfirmPasswordResetCommand struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// PasswordResetResult represents the result of a password reset operation
type PasswordResetResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PasswordResetCommand defines the inbound port for password reset operations
type PasswordResetCommand interface {
	// RequestPasswordReset initiates a password reset and sends email
	RequestPasswordReset(ctx context.Context, cmd RequestPasswordResetCommand) (*auth_entities.PasswordReset, error)

	// ConfirmPasswordReset completes the password reset with new password
	ConfirmPasswordReset(ctx context.Context, cmd ConfirmPasswordResetCommand) (*PasswordResetResult, error)

	// ValidateResetToken validates a reset token without using it
	ValidateResetToken(ctx context.Context, token string) (*auth_entities.PasswordReset, error)

	// GetResetStatus gets the status of a password reset by user ID
	GetResetStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.PasswordReset, error)

	// CancelPasswordReset cancels a pending password reset
	CancelPasswordReset(ctx context.Context, resetID uuid.UUID) error
}

