package auth_entities

import (
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// MFAMethod represents the type of MFA method
type MFAMethod string

const (
	MFAMethodTOTP     MFAMethod = "totp"     // Time-based One-Time Password (Google Authenticator, etc.)
	MFAMethodSMS      MFAMethod = "sms"      // SMS-based OTP
	MFAMethodEmail    MFAMethod = "email"    // Email-based OTP
	MFAMethodBackup   MFAMethod = "backup"   // Backup codes
	MFAMethodWebAuthn MFAMethod = "webauthn" // Hardware security keys
)

// MFAStatus represents the status of MFA setup
type MFAStatus string

const (
	MFAStatusPending   MFAStatus = "pending"   // Setup initiated but not verified
	MFAStatusActive    MFAStatus = "active"    // MFA is enabled and verified
	MFAStatusDisabled  MFAStatus = "disabled"  // MFA was disabled
	MFAStatusRecovery  MFAStatus = "recovery"  // User is in recovery mode
)

// TOTPConfig stores TOTP-specific configuration
type TOTPConfig struct {
	Secret    string    `json:"-" bson:"secret"` // Base32 encoded secret - never expose in JSON
	Algorithm string    `json:"algorithm" bson:"algorithm"` // SHA1, SHA256, SHA512
	Digits    int       `json:"digits" bson:"digits"` // Number of digits (usually 6)
	Period    int       `json:"period" bson:"period"` // Time step in seconds (usually 30)
	Issuer    string    `json:"issuer" bson:"issuer"` // App name for authenticator display
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// BackupCode represents a one-time backup code
type BackupCode struct {
	Code      string    `json:"-" bson:"code"` // Hashed backup code
	UsedAt    *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// UserMFA stores MFA configuration for a user
type UserMFA struct {
	ID              uuid.UUID            `json:"id" bson:"_id"`
	UserID          uuid.UUID            `json:"user_id" bson:"user_id"`
	Method          MFAMethod            `json:"method" bson:"method"`
	Status          MFAStatus            `json:"status" bson:"status"`
	TOTPConfig      *TOTPConfig          `json:"totp_config,omitempty" bson:"totp_config,omitempty"`
	BackupCodes     []BackupCode         `json:"-" bson:"backup_codes"` // Never expose backup codes
	BackupCodesLeft int                  `json:"backup_codes_left" bson:"backup_codes_left"`
	VerifiedAt      *time.Time           `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	LastUsedAt      *time.Time           `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
	ResourceOwner   common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt       time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" bson:"updated_at"`
}

// MFASetupResponse contains the data needed to set up MFA on user's device
type MFASetupResponse struct {
	Method      MFAMethod `json:"method"`
	Secret      string    `json:"secret,omitempty"`       // Only for TOTP setup
	QRCode      string    `json:"qr_code,omitempty"`      // Base64 encoded QR code image
	URI         string    `json:"uri,omitempty"`          // otpauth:// URI
	BackupCodes []string  `json:"backup_codes,omitempty"` // Only shown once during setup
}

// MFAVerifyRequest contains the data for MFA verification
type MFAVerifyRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
	Method MFAMethod `json:"method"`
}

// NewUserMFA creates a new MFA configuration for a user
func NewUserMFA(userID uuid.UUID, method MFAMethod, rxn common.ResourceOwner) *UserMFA {
	now := time.Now()
	return &UserMFA{
		ID:            uuid.New(),
		UserID:        userID,
		Method:        method,
		Status:        MFAStatusPending,
		ResourceOwner: rxn,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GenerateTOTPSecret generates a new TOTP secret
func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, 20) // 160 bits
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// GenerateBackupCodes generates a set of backup codes
func GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code := make([]byte, 6) // 8 character codes
		if _, err := rand.Read(code); err != nil {
			return nil, err
		}
		// Format as XXXX-XXXX for readability
		encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(code)
		codes[i] = encoded[:4] + "-" + encoded[4:8]
	}
	return codes, nil
}

// SetupTOTP initializes TOTP configuration for the user
func (m *UserMFA) SetupTOTP(secret, issuer, userEmail string) {
	m.TOTPConfig = &TOTPConfig{
		Secret:    secret,
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Issuer:    issuer,
		CreatedAt: time.Now(),
	}
}

// Activate marks the MFA as active after successful verification
func (m *UserMFA) Activate() {
	now := time.Now()
	m.Status = MFAStatusActive
	m.VerifiedAt = &now
	m.UpdatedAt = now
}

// Disable disables MFA for the user
func (m *UserMFA) Disable() {
	m.Status = MFAStatusDisabled
	m.UpdatedAt = time.Now()
}

// RecordUsage records that MFA was used
func (m *UserMFA) RecordUsage() {
	now := time.Now()
	m.LastUsedAt = &now
	m.UpdatedAt = now
}

// GetID returns the entity ID
func (m UserMFA) GetID() uuid.UUID {
	return m.ID
}

// IsActive returns whether MFA is enabled
func (m UserMFA) IsActive() bool {
	return m.Status == MFAStatusActive
}

// IsPending returns whether MFA setup is pending verification
func (m UserMFA) IsPending() bool {
	return m.Status == MFAStatusPending
}

