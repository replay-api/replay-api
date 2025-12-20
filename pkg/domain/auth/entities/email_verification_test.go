package auth_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// NewEmailVerification TESTS
// =============================================================================

// TestNewEmailVerification_CreatesValidInstance validates email verification creation
// Business context: Email verification is the first security step in user onboarding.
// When a user signs up, they must verify their email before accessing platform
// features like matchmaking and wallet operations.
func TestNewEmailVerification_CreatesValidInstance(t *testing.T) {
	userID := uuid.New()
	email := "player@example.com"
	ttlMinutes := 30

	verification, err := NewEmailVerification(userID, email, VerificationTypeEmail, ttlMinutes)

	assert.NoError(t, err)
	assert.NotNil(t, verification)
	assert.NotEqual(t, uuid.Nil, verification.ID)
	assert.Equal(t, userID, verification.UserID)
	assert.Equal(t, email, verification.Email)
	assert.NotEmpty(t, verification.Token, "Should generate secure token")
	assert.NotEmpty(t, verification.Code, "Should generate verification code")
	assert.Equal(t, VerificationTypeEmail, verification.Type)
	assert.Equal(t, VerificationStatusPending, verification.Status)
	assert.Equal(t, 0, verification.Attempts)
	assert.Equal(t, 5, verification.MaxAttempts)
	assert.False(t, verification.CreatedAt.IsZero())
	assert.False(t, verification.UpdatedAt.IsZero())
	assert.Nil(t, verification.VerifiedAt)
}

// TestNewEmailVerification_SetsCorrectExpiry validates expiration calculation
// Business context: Verification links/codes have limited validity (typically 24h)
// to prevent stale tokens from being used and limit attack windows.
func TestNewEmailVerification_SetsCorrectExpiry(t *testing.T) {
	testCases := []struct {
		ttlMinutes      int
		expectedMinutes int
	}{
		{15, 15},
		{30, 30},
		{60, 60},
		{1440, 1440}, // 24 hours
	}

	for _, tc := range testCases {
		t.Run("ttl_"+string(rune('0'+tc.ttlMinutes%10)), func(t *testing.T) {
			before := time.Now().UTC()
			verification, err := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, tc.ttlMinutes)
			
			assert.NoError(t, err)
			expectedExpiry := before.Add(time.Duration(tc.ttlMinutes) * time.Minute)
			
			// Allow 1 second tolerance
			assert.WithinDuration(t, expectedExpiry, verification.ExpiresAt, time.Second)
		})
	}
}

// TestNewEmailVerification_GeneratesUniqueTokens ensures token uniqueness
// Business context: Each verification must have unique tokens to prevent
// cross-user token reuse attacks.
func TestNewEmailVerification_GeneratesUniqueTokens(t *testing.T) {
	v1, err1 := NewEmailVerification(uuid.New(), "user1@example.com", VerificationTypeEmail, 30)
	v2, err2 := NewEmailVerification(uuid.New(), "user2@example.com", VerificationTypeEmail, 30)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, v1.Token, v2.Token, "Tokens should be unique")
	assert.NotEqual(t, v1.Code, v2.Code, "Codes should be unique")
}

// TestNewEmailVerification_SupportsAllTypes validates all verification types
func TestNewEmailVerification_SupportsAllTypes(t *testing.T) {
	types := []VerificationType{VerificationTypeEmail, VerificationTypeMFA}

	for _, vType := range types {
		t.Run(string(vType), func(t *testing.T) {
			v, err := NewEmailVerification(uuid.New(), "test@example.com", vType, 30)
			assert.NoError(t, err)
			assert.Equal(t, vType, v.Type)
		})
	}
}

// =============================================================================
// VERIFY TESTS
// =============================================================================

// TestVerify_SucceedsWithValidCode validates code-based verification
// Business context: Users can verify by entering the 6-digit code sent to
// their email, which is more convenient on mobile devices than clicking links.
func TestVerify_SucceedsWithValidCode(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	code := v.Code

	result := v.Verify(code)

	assert.True(t, result)
	assert.Equal(t, VerificationStatusVerified, v.Status)
	assert.NotNil(t, v.VerifiedAt)
	assert.True(t, v.IsVerified())
	assert.False(t, v.IsPending())
}

// TestVerify_SucceedsWithValidToken validates token-based verification
// Business context: Users can verify by clicking the email link, which
// includes the full token for one-click verification.
func TestVerify_SucceedsWithValidToken(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	token := v.Token

	result := v.Verify(token)

	assert.True(t, result)
	assert.Equal(t, VerificationStatusVerified, v.Status)
	assert.True(t, v.IsVerified())
}

// TestVerify_FailsWithInvalidCode validates rejection of wrong codes
// Business context: Incorrect codes must be rejected and attempts tracked
// to prevent brute-force attacks on verification codes.
func TestVerify_FailsWithInvalidCode(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	initialAttempts := v.Attempts

	result := v.Verify("000000")

	assert.False(t, result)
	assert.Equal(t, VerificationStatusPending, v.Status)
	assert.Nil(t, v.VerifiedAt)
	assert.Equal(t, initialAttempts+1, v.Attempts)
}

// TestVerify_FailsAfterMaxAttempts validates brute-force protection
// Business context: After 5 failed attempts, verification is locked to
// prevent automated attacks. User must request a new verification.
func TestVerify_FailsAfterMaxAttempts(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	correctCode := v.Code

	// Exhaust all attempts with wrong codes
	for i := 0; i < 5; i++ {
		v.Verify("000000")
	}

	assert.Equal(t, 0, v.RemainingAttempts())

	// Even correct code should now fail
	result := v.Verify(correctCode)
	assert.False(t, result, "Should fail after max attempts exhausted")
}

// TestVerify_FailsWhenExpired validates expiration enforcement
// Business context: Expired verifications must be rejected regardless of
// code correctness to maintain security window constraints.
func TestVerify_FailsWhenExpired(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	v.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour) // Expired 1 hour ago
	code := v.Code

	result := v.Verify(code)

	assert.False(t, result)
	assert.Equal(t, VerificationStatusExpired, v.Status)
}

// =============================================================================
// STATUS METHODS TESTS
// =============================================================================

// TestIsExpired_DetectsExpiration validates expiration detection
func TestIsExpired_DetectsExpiration(t *testing.T) {
	tests := []struct {
		name     string
		offset   time.Duration
		expected bool
	}{
		{"future expiry", 1 * time.Hour, false},
		{"past expiry", -1 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
			v.ExpiresAt = time.Now().UTC().Add(tt.offset)

			assert.Equal(t, tt.expected, v.IsExpired())
		})
	}
}

// TestIsPending_ReflectsCorrectState validates pending status detection
func TestIsPending_ReflectsCorrectState(t *testing.T) {
	// New verification should be pending
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	assert.True(t, v.IsPending())

	// Verified should not be pending
	v.Verify(v.Code)
	assert.False(t, v.IsPending())
}

// TestIsPending_FailsWhenExpired ensures expired verifications are not pending
func TestIsPending_FailsWhenExpired(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	v.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)

	assert.False(t, v.IsPending(), "Expired verification should not be pending")
}

// TestIsVerified_ReflectsCorrectState validates verified status detection
func TestIsVerified_ReflectsCorrectState(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)

	assert.False(t, v.IsVerified())

	v.Verify(v.Code)

	assert.True(t, v.IsVerified())
}

// =============================================================================
// CANCEL TESTS
// =============================================================================

// TestCancel_SetsCorrectStatus validates verification cancellation
// Business context: Users or admins may cancel pending verifications,
// e.g., when the user changes their email address during signup.
func TestCancel_SetsCorrectStatus(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	beforeCancel := time.Now().UTC()

	v.Cancel()

	assert.Equal(t, VerificationStatusCanceled, v.Status)
	assert.True(t, v.UpdatedAt.After(beforeCancel) || v.UpdatedAt.Equal(beforeCancel))
	assert.False(t, v.IsPending())
	assert.False(t, v.IsVerified())
}

// =============================================================================
// REMAINING ATTEMPTS TESTS
// =============================================================================

// TestRemainingAttempts_TracksCorrectly validates attempt counting
func TestRemainingAttempts_TracksCorrectly(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)

	assert.Equal(t, 5, v.RemainingAttempts())

	v.Verify("wrong")
	assert.Equal(t, 4, v.RemainingAttempts())

	v.Verify("wrong")
	v.Verify("wrong")
	assert.Equal(t, 2, v.RemainingAttempts())
}

// TestRemainingAttempts_NeverNegative ensures count never goes negative
func TestRemainingAttempts_NeverNegative(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	v.Attempts = 10 // More than max

	assert.Equal(t, 0, v.RemainingAttempts())
}

// =============================================================================
// REQUEST INFO TESTS
// =============================================================================

// TestSetRequestInfo_StoresAuditData validates audit data storage
// Business context: IP and user agent are recorded for security auditing
// and fraud detection (e.g., detecting verification from unusual locations).
func TestSetRequestInfo_StoresAuditData(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)
	ip := "192.168.1.1"
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
	beforeSet := time.Now().UTC()

	v.SetRequestInfo(ip, ua)

	assert.Equal(t, ip, v.IPAddress)
	assert.Equal(t, ua, v.UserAgent)
	assert.True(t, v.UpdatedAt.After(beforeSet) || v.UpdatedAt.Equal(beforeSet))
}

// =============================================================================
// ACCESSOR TESTS
// =============================================================================

// TestGetID_ReturnsCorrectID validates ID accessor
func TestGetID_ReturnsCorrectID(t *testing.T) {
	v, _ := NewEmailVerification(uuid.New(), "test@example.com", VerificationTypeEmail, 30)

	assert.Equal(t, v.ID, v.GetID())
}

// =============================================================================
// TOKEN GENERATION TESTS
// =============================================================================

// TestGenerateSecureToken_GeneratesCorrectLength validates token generation
func TestGenerateSecureToken_GeneratesCorrectLength(t *testing.T) {
	token, err := generateSecureToken(32)

	assert.NoError(t, err)
	assert.Len(t, token, 64) // Hex encoding doubles length
}

// TestGenerateSecureToken_GeneratesUniqueTokens validates randomness
func TestGenerateSecureToken_GeneratesUniqueTokens(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, _ := generateSecureToken(32)
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
	}
}

// TestGenerateNumericCode_GeneratesCorrectFormat validates code format
func TestGenerateNumericCode_GeneratesCorrectFormat(t *testing.T) {
	code, err := generateNumericCode(6)

	assert.NoError(t, err)
	assert.Len(t, code, 6)

	// Should only contain digits
	for _, c := range code {
		assert.True(t, c >= '0' && c <= '9', "Code should only contain digits")
	}
}

// TestGenerateNumericCode_GeneratesUniqueCodes validates randomness
func TestGenerateNumericCode_GeneratesUniqueCodes(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, _ := generateNumericCode(6)
		codes[code] = true
	}
	// With 6 digits and 100 attempts, we expect very high uniqueness
	assert.Greater(t, len(codes), 90, "Most codes should be unique")
}

// =============================================================================
// CONSTANT TESTS
// =============================================================================

// TestVerificationStatus_Constants verifies status constants
func TestVerificationStatus_Constants(t *testing.T) {
	assert.Equal(t, VerificationStatus("pending"), VerificationStatusPending)
	assert.Equal(t, VerificationStatus("verified"), VerificationStatusVerified)
	assert.Equal(t, VerificationStatus("expired"), VerificationStatusExpired)
	assert.Equal(t, VerificationStatus("canceled"), VerificationStatusCanceled)
}

// TestVerificationType_Constants verifies type constants
func TestVerificationType_Constants(t *testing.T) {
	assert.Equal(t, VerificationType("email"), VerificationTypeEmail)
	assert.Equal(t, VerificationType("mfa"), VerificationTypeMFA)
}

