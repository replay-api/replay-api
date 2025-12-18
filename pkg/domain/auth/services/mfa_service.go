package auth_services

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" // #nosec G505 - TOTP (RFC 6238) requires HMAC-SHA1
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
)

const (
	totpIssuer     = "LeetGaming.PRO"
	backupCodeCount = 10
)

type MFAService struct {
	mfaRepo        auth_out.MFARepository
	passwordHasher email_out.PasswordHasher
}

func NewMFAService(
	mfaRepo auth_out.MFARepository,
	passwordHasher email_out.PasswordHasher,
) auth_in.MFACommand {
	return &MFAService{
		mfaRepo:        mfaRepo,
		passwordHasher: passwordHasher,
	}
}

// SetupTOTP initiates TOTP setup for a user
func (s *MFAService) SetupTOTP(ctx context.Context, userID uuid.UUID, email string) (*auth_entities.MFASetupResponse, error) {
	slog.InfoContext(ctx, "Setting up TOTP MFA", "user_id", userID)
	
	// Check if MFA already exists
	existing, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err == nil && existing != nil && existing.Status == auth_entities.MFAStatusActive {
		return nil, fmt.Errorf("MFA is already enabled for this user")
	}
	
	// Generate TOTP secret
	secret, err := auth_entities.GenerateTOTPSecret()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate TOTP secret", "error", err)
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}
	
	// Generate backup codes
	backupCodes, err := auth_entities.GenerateBackupCodes(backupCodeCount)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate backup codes", "error", err)
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}
	
	// Hash backup codes for storage
	var hashedCodes []auth_entities.BackupCode
	for _, code := range backupCodes {
		hashedCode, err := s.passwordHasher.HashPassword(ctx, strings.ReplaceAll(code, "-", ""))
		if err != nil {
			slog.ErrorContext(ctx, "Failed to hash backup code", "error", err)
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedCodes = append(hashedCodes, auth_entities.BackupCode{
			Code:      hashedCode,
			CreatedAt: time.Now(),
		})
	}
	
	// Create or update MFA record
	rxn := common.GetResourceOwner(ctx)
	mfa := auth_entities.NewUserMFA(userID, auth_entities.MFAMethodTOTP, rxn)
	mfa.SetupTOTP(secret, totpIssuer, email)
	mfa.BackupCodes = hashedCodes
	mfa.BackupCodesLeft = len(hashedCodes)
	
	// If existing pending MFA, update it; otherwise create new
	if existing != nil {
		mfa.ID = existing.ID
		_, err = s.mfaRepo.Update(ctx, mfa)
	} else {
		_, err = s.mfaRepo.Create(ctx, mfa)
	}
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save MFA configuration", "error", err)
		return nil, fmt.Errorf("failed to save MFA configuration: %w", err)
	}
	
	// Generate otpauth URI (frontend will generate QR code)
	uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		totpIssuer, email, secret, totpIssuer)
	
	slog.InfoContext(ctx, "TOTP MFA setup initiated", "user_id", userID)
	
	return &auth_entities.MFASetupResponse{
		Method:      auth_entities.MFAMethodTOTP,
		Secret:      secret,
		URI:         uri,
		BackupCodes: backupCodes,
	}, nil
}

// VerifyAndActivate verifies the TOTP code and activates MFA
func (s *MFAService) VerifyAndActivate(ctx context.Context, userID uuid.UUID, code string) error {
	slog.InfoContext(ctx, "Verifying and activating TOTP MFA", "user_id", userID)
	
	mfa, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("MFA not found")
	}
	
	if mfa.Status == auth_entities.MFAStatusActive {
		return fmt.Errorf("MFA is already active")
	}
	
	if mfa.TOTPConfig == nil {
		return fmt.Errorf("TOTP not configured")
	}
	
	// Verify the code
	if !s.verifyTOTPCode(mfa.TOTPConfig.Secret, code) {
		slog.WarnContext(ctx, "Invalid TOTP code during activation", "user_id", userID)
		return fmt.Errorf("invalid TOTP code")
	}
	
	// Activate MFA
	mfa.Activate()
	_, err = s.mfaRepo.Update(ctx, mfa)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to activate MFA", "error", err)
		return fmt.Errorf("failed to activate MFA: %w", err)
	}
	
	slog.InfoContext(ctx, "TOTP MFA activated successfully", "user_id", userID)
	return nil
}

// Verify verifies a TOTP code during login
func (s *MFAService) Verify(ctx context.Context, userID uuid.UUID, code string) error {
	mfa, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("MFA not found")
	}
	
	if mfa.Status != auth_entities.MFAStatusActive {
		return fmt.Errorf("MFA is not active")
	}
	
	if mfa.TOTPConfig == nil {
		return fmt.Errorf("TOTP not configured")
	}
	
	if !s.verifyTOTPCode(mfa.TOTPConfig.Secret, code) {
		slog.WarnContext(ctx, "Invalid TOTP code during verification", "user_id", userID)
		return fmt.Errorf("invalid TOTP code")
	}
	
	// Record usage
	mfa.RecordUsage()
	_, _ = s.mfaRepo.Update(ctx, mfa)
	
	slog.InfoContext(ctx, "TOTP verification successful", "user_id", userID)
	return nil
}

// VerifyBackupCode verifies a backup code during login
func (s *MFAService) VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
	mfa, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("MFA not found")
	}
	
	if mfa.Status != auth_entities.MFAStatusActive {
		return fmt.Errorf("MFA is not active")
	}
	
	// Normalize code
	normalizedCode := strings.ReplaceAll(strings.ToUpper(code), "-", "")
	
	// Check each unused backup code
	for i, bc := range mfa.BackupCodes {
		if bc.UsedAt != nil {
			continue // Skip used codes
		}
		
		err := s.passwordHasher.ComparePassword(ctx, bc.Code, normalizedCode)
		if err == nil {
			// Code matches - mark as used
			now := time.Now()
			mfa.BackupCodes[i].UsedAt = &now
			mfa.BackupCodesLeft--
			mfa.RecordUsage()
			
			_, err = s.mfaRepo.Update(ctx, mfa)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to update backup code status", "error", err)
			}
			
			slog.InfoContext(ctx, "Backup code verification successful", "user_id", userID, "codes_remaining", mfa.BackupCodesLeft)
			return nil
		}
	}
	
	slog.WarnContext(ctx, "Invalid backup code", "user_id", userID)
	return fmt.Errorf("invalid backup code")
}

// Disable disables MFA for a user
func (s *MFAService) Disable(ctx context.Context, userID uuid.UUID) error {
	mfa, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("MFA not found")
	}
	
	mfa.Disable()
	_, err = s.mfaRepo.Update(ctx, mfa)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to disable MFA", "error", err)
		return fmt.Errorf("failed to disable MFA: %w", err)
	}
	
	slog.InfoContext(ctx, "MFA disabled successfully", "user_id", userID)
	return nil
}

// GetStatus gets the current MFA status for a user
func (s *MFAService) GetStatus(ctx context.Context, userID uuid.UUID) (*auth_entities.UserMFA, error) {
	mfa, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Return nil status for users without MFA
		return nil, nil
	}
	
	return mfa, nil
}

// RegenerateBackupCodes regenerates backup codes
func (s *MFAService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	mfa, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("MFA not found")
	}
	
	if mfa.Status != auth_entities.MFAStatusActive {
		return nil, fmt.Errorf("MFA is not active")
	}
	
	// Verify current TOTP code before regenerating
	if !s.verifyTOTPCode(mfa.TOTPConfig.Secret, code) {
		return nil, fmt.Errorf("invalid TOTP code")
	}
	
	// Generate new backup codes
	newCodes, err := auth_entities.GenerateBackupCodes(backupCodeCount)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}
	
	// Hash and store new codes
	var hashedCodes []auth_entities.BackupCode
	for _, newCode := range newCodes {
		hashedCode, err := s.passwordHasher.HashPassword(ctx, strings.ReplaceAll(newCode, "-", ""))
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedCodes = append(hashedCodes, auth_entities.BackupCode{
			Code:      hashedCode,
			CreatedAt: time.Now(),
		})
	}
	
	mfa.BackupCodes = hashedCodes
	mfa.BackupCodesLeft = len(hashedCodes)
	mfa.UpdatedAt = time.Now()
	
	_, err = s.mfaRepo.Update(ctx, mfa)
	if err != nil {
		return nil, fmt.Errorf("failed to save backup codes: %w", err)
	}
	
	slog.InfoContext(ctx, "Backup codes regenerated", "user_id", userID)
	return newCodes, nil
}

// verifyTOTPCode verifies a TOTP code
func (s *MFAService) verifyTOTPCode(secret, code string) bool {
	// Allow for time drift (1 step before and after current)
	currentTime := time.Now().Unix()
	timeSteps := []int64{-1, 0, 1}
	
	for _, step := range timeSteps {
		expectedCode := generateTOTP(secret, currentTime/30+step)
		if code == expectedCode {
			return true
		}
	}
	
	return false
}

// generateTOTP generates a TOTP code for a given time step
func generateTOTP(secret string, counter int64) string {
	// Decode base32 secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}
	
	// Convert counter to bytes (big-endian)
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(counter))
	
	// HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hash := h.Sum(nil)
	
	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := int64(hash[offset]&0x7f)<<24 |
		int64(hash[offset+1]&0xff)<<16 |
		int64(hash[offset+2]&0xff)<<8 |
		int64(hash[offset+3]&0xff)
	
	// Get 6 digits
	code = code % 1000000
	
	return fmt.Sprintf("%06d", code)
}
