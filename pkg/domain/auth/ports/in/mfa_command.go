package auth_in

import (
	"context"

	"github.com/google/uuid"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
)

// MFACommand defines the interface for MFA operations
type MFACommand interface {
	// SetupTOTP initiates TOTP setup for a user
	SetupTOTP(ctx context.Context, userID uuid.UUID, email string) (*auth_entities.MFASetupResponse, error)
	
	// VerifyAndActivate verifies the TOTP code and activates MFA
	VerifyAndActivate(ctx context.Context, userID uuid.UUID, code string) error
	
	// Verify verifies a TOTP code during login
	Verify(ctx context.Context, userID uuid.UUID, code string) error
	
	// VerifyBackupCode verifies a backup code during login
	VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error
	
	// Disable disables MFA for a user (requires password verification)
	Disable(ctx context.Context, userID uuid.UUID) error
	
	// GetStatus gets the current MFA status for a user
	GetStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.UserMFA, error)
	
	// RegenerateBackupCodes regenerates backup codes (requires existing MFA verification)
	RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error)
}

