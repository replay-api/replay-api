package auth_out

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
)

// MFARepository defines the interface for MFA persistence
type MFARepository interface {
	// Create creates a new MFA configuration
	Create(ctx context.Context, mfa *auth_entities.UserMFA) (*auth_entities.UserMFA, error)
	
	// GetByUserID gets the MFA configuration for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) (*auth_entities.UserMFA, error)
	
	// Update updates an existing MFA configuration
	Update(ctx context.Context, mfa *auth_entities.UserMFA) (*auth_entities.UserMFA, error)
	
	// Delete deletes an MFA configuration
	Delete(ctx context.Context, id uuid.UUID) error
}

