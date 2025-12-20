package auth_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func testResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		ClientID: uuid.New(),
	}
}

// =============================================================================
// NewUserMFA TESTS
// =============================================================================

// TestNewUserMFA_CreatesValidInstance validates MFA entity creation
// Business context: MFA setup is critical for account security. When a user
// initiates MFA enrollment, the system must create a properly initialized
// entity with pending status, ready for the verification workflow.
func TestNewUserMFA_CreatesValidInstance(t *testing.T) {
	userID := uuid.New()
	rxn := testResourceOwner()

	mfa := NewUserMFA(userID, MFAMethodTOTP, rxn)

	assert.NotNil(t, mfa)
	assert.NotEqual(t, uuid.Nil, mfa.ID, "Should generate a unique ID")
	assert.Equal(t, userID, mfa.UserID, "Should set the user ID")
	assert.Equal(t, MFAMethodTOTP, mfa.Method, "Should set the MFA method")
	assert.Equal(t, MFAStatusPending, mfa.Status, "New MFA should start in pending status")
	assert.Equal(t, rxn, mfa.ResourceOwner, "Should set resource owner")
	assert.False(t, mfa.CreatedAt.IsZero(), "Should set creation timestamp")
	assert.False(t, mfa.UpdatedAt.IsZero(), "Should set update timestamp")
	assert.Nil(t, mfa.VerifiedAt, "Should not be verified yet")
	assert.Nil(t, mfa.LastUsedAt, "Should not have usage yet")
}

// TestNewUserMFA_AllMethods verifies all supported MFA methods can be used
// Business context: Platform supports multiple MFA methods (TOTP, SMS, email, 
// WebAuthn, backup codes) to accommodate different user security preferences
// and compliance requirements.
func TestNewUserMFA_AllMethods(t *testing.T) {
	methods := []MFAMethod{
		MFAMethodTOTP,
		MFAMethodSMS,
		MFAMethodEmail,
		MFAMethodBackup,
		MFAMethodWebAuthn,
	}

	for _, method := range methods {
		t.Run(string(method), func(t *testing.T) {
			mfa := NewUserMFA(uuid.New(), method, testResourceOwner())
			assert.Equal(t, method, mfa.Method)
		})
	}
}

// =============================================================================
// TOTP TESTS
// =============================================================================

// TestGenerateTOTPSecret_GeneratesSecureSecret validates TOTP secret generation
// Business context: TOTP secrets must be cryptographically secure (160 bits)
// and properly encoded in Base32 for compatibility with authenticator apps
// like Google Authenticator, Authy, and 1Password.
func TestGenerateTOTPSecret_GeneratesSecureSecret(t *testing.T) {
	secret1, err := GenerateTOTPSecret()
	assert.NoError(t, err)
	assert.NotEmpty(t, secret1)

	// Base32 encoded 160-bit secret should be 32 characters
	assert.GreaterOrEqual(t, len(secret1), 32, "Secret should be at least 32 chars")

	// Each call should generate unique secrets
	secret2, err := GenerateTOTPSecret()
	assert.NoError(t, err)
	assert.NotEqual(t, secret1, secret2, "Each secret should be unique")
}

// TestSetupTOTP_InitializesConfig validates TOTP configuration setup
// Business context: TOTP setup requires standard configuration parameters
// (SHA1 algorithm, 6 digits, 30-second period) for compatibility with
// RFC 6238 compliant authenticator applications.
func TestSetupTOTP_InitializesConfig(t *testing.T) {
	mfa := NewUserMFA(uuid.New(), MFAMethodTOTP, testResourceOwner())
	secret := "JBSWY3DPEHPK3PXP" // Example secret
	issuer := "LeetGaming"
	userEmail := "player@example.com"

	mfa.SetupTOTP(secret, issuer, userEmail)

	assert.NotNil(t, mfa.TOTPConfig)
	assert.Equal(t, secret, mfa.TOTPConfig.Secret)
	assert.Equal(t, "SHA1", mfa.TOTPConfig.Algorithm)
	assert.Equal(t, 6, mfa.TOTPConfig.Digits)
	assert.Equal(t, 30, mfa.TOTPConfig.Period)
	assert.Equal(t, issuer, mfa.TOTPConfig.Issuer)
	assert.False(t, mfa.TOTPConfig.CreatedAt.IsZero())
}

// =============================================================================
// BACKUP CODES TESTS
// =============================================================================

// TestGenerateBackupCodes_GeneratesRequestedCount validates backup code generation
// Business context: Backup codes are emergency one-time recovery codes that
// allow users to access their account if they lose access to their primary
// MFA device. Platform generates 10 codes by default, each usable once.
func TestGenerateBackupCodes_GeneratesRequestedCount(t *testing.T) {
	testCases := []struct {
		count int
	}{
		{count: 1},
		{count: 5},
		{count: 10},
		{count: 16},
	}

	for _, tc := range testCases {
		t.Run("count_"+string(rune('0'+tc.count)), func(t *testing.T) {
			codes, err := GenerateBackupCodes(tc.count)
			assert.NoError(t, err)
			assert.Len(t, codes, tc.count)
		})
	}
}

// TestGenerateBackupCodes_FormatsCorrectly validates backup code format
// Business context: Backup codes use XXXX-XXXX format for better readability
// and reduced transcription errors when users need to enter them manually.
func TestGenerateBackupCodes_FormatsCorrectly(t *testing.T) {
	codes, err := GenerateBackupCodes(10)
	assert.NoError(t, err)

	for _, code := range codes {
		// Format: XXXX-XXXX (9 characters)
		assert.Len(t, code, 9, "Code should be 9 characters (XXXX-XXXX)")
		assert.Equal(t, "-", string(code[4]), "Code should have hyphen at position 4")
	}
}

// TestGenerateBackupCodes_GeneratesUniqueCodes ensures all codes are unique
// Business context: Each backup code must be unique to prevent reuse attacks.
// Statistical collisions in 10 codes from a proper random source are
// effectively impossible but should be verified.
func TestGenerateBackupCodes_GeneratesUniqueCodes(t *testing.T) {
	codes, err := GenerateBackupCodes(10)
	assert.NoError(t, err)

	seen := make(map[string]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "All codes should be unique")
		seen[code] = true
	}
}

// =============================================================================
// MFA LIFECYCLE TESTS
// =============================================================================

// TestMFA_Activate_TransitionsToActive validates MFA activation workflow
// Business context: After user verifies their MFA setup (e.g., entering a 
// correct TOTP code), the system activates MFA protection on their account.
// This is a security-critical transition that enables 2FA enforcement.
func TestMFA_Activate_TransitionsToActive(t *testing.T) {
	mfa := NewUserMFA(uuid.New(), MFAMethodTOTP, testResourceOwner())
	beforeActivation := time.Now()

	// Verify initial pending state
	assert.Equal(t, MFAStatusPending, mfa.Status)
	assert.False(t, mfa.IsActive())
	assert.True(t, mfa.IsPending())

	mfa.Activate()

	// Verify active state
	assert.Equal(t, MFAStatusActive, mfa.Status)
	assert.True(t, mfa.IsActive())
	assert.False(t, mfa.IsPending())
	assert.NotNil(t, mfa.VerifiedAt)
	assert.True(t, mfa.VerifiedAt.After(beforeActivation) || mfa.VerifiedAt.Equal(beforeActivation))
	assert.True(t, mfa.UpdatedAt.After(beforeActivation) || mfa.UpdatedAt.Equal(beforeActivation))
}

// TestMFA_Disable_TransitionsToDisabled validates MFA disabling workflow
// Business context: Users may need to disable MFA (e.g., changing phone numbers,
// security key replacement). This must properly update status while maintaining
// audit trail. Re-enabling MFA requires fresh setup for security.
func TestMFA_Disable_TransitionsToDisabled(t *testing.T) {
	mfa := NewUserMFA(uuid.New(), MFAMethodTOTP, testResourceOwner())
	mfa.Activate()
	beforeDisable := time.Now()

	assert.True(t, mfa.IsActive())

	mfa.Disable()

	assert.Equal(t, MFAStatusDisabled, mfa.Status)
	assert.False(t, mfa.IsActive())
	assert.False(t, mfa.IsPending())
	assert.True(t, mfa.UpdatedAt.After(beforeDisable) || mfa.UpdatedAt.Equal(beforeDisable))
}

// TestMFA_RecordUsage_TracksLastUsed validates usage tracking
// Business context: Tracking MFA usage helps detect suspicious account activity
// and supports compliance auditing (e.g., when was 2FA last used for login).
func TestMFA_RecordUsage_TracksLastUsed(t *testing.T) {
	mfa := NewUserMFA(uuid.New(), MFAMethodTOTP, testResourceOwner())
	mfa.Activate()

	assert.Nil(t, mfa.LastUsedAt)

	beforeUsage := time.Now()
	mfa.RecordUsage()

	assert.NotNil(t, mfa.LastUsedAt)
	assert.True(t, mfa.LastUsedAt.After(beforeUsage) || mfa.LastUsedAt.Equal(beforeUsage))
	assert.True(t, mfa.UpdatedAt.After(beforeUsage) || mfa.UpdatedAt.Equal(beforeUsage))
}

// =============================================================================
// ACCESSOR TESTS
// =============================================================================

// TestMFA_GetID_ReturnsCorrectID validates ID accessor
func TestMFA_GetID_ReturnsCorrectID(t *testing.T) {
	mfa := NewUserMFA(uuid.New(), MFAMethodTOTP, testResourceOwner())

	assert.Equal(t, mfa.ID, mfa.GetID())
}

// TestMFA_StatusMethods_ReflectCurrentState validates status accessors
func TestMFA_StatusMethods_ReflectCurrentState(t *testing.T) {
	tests := []struct {
		name          string
		status        MFAStatus
		expectActive  bool
		expectPending bool
	}{
		{"pending", MFAStatusPending, false, true},
		{"active", MFAStatusActive, true, false},
		{"disabled", MFAStatusDisabled, false, false},
		{"recovery", MFAStatusRecovery, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mfa := NewUserMFA(uuid.New(), MFAMethodTOTP, testResourceOwner())
			mfa.Status = tt.status

			assert.Equal(t, tt.expectActive, mfa.IsActive())
			assert.Equal(t, tt.expectPending, mfa.IsPending())
		})
	}
}

// =============================================================================
// CONSTANT TESTS
// =============================================================================

// TestMFAMethod_Constants verifies MFA method constants
func TestMFAMethod_Constants(t *testing.T) {
	// Verify constants haven't changed (API contract)
	assert.Equal(t, MFAMethod("totp"), MFAMethodTOTP)
	assert.Equal(t, MFAMethod("sms"), MFAMethodSMS)
	assert.Equal(t, MFAMethod("email"), MFAMethodEmail)
	assert.Equal(t, MFAMethod("backup"), MFAMethodBackup)
	assert.Equal(t, MFAMethod("webauthn"), MFAMethodWebAuthn)
}

// TestMFAStatus_Constants verifies MFA status constants
func TestMFAStatus_Constants(t *testing.T) {
	// Verify constants haven't changed (API contract)
	assert.Equal(t, MFAStatus("pending"), MFAStatusPending)
	assert.Equal(t, MFAStatus("active"), MFAStatusActive)
	assert.Equal(t, MFAStatus("disabled"), MFAStatusDisabled)
	assert.Equal(t, MFAStatus("recovery"), MFAStatusRecovery)
}

