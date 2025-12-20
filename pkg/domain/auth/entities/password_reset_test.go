package auth_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// NewPasswordReset TESTS
// =============================================================================

// TestNewPasswordReset_CreatesValidInstance validates password reset creation
// Business context: Password reset is a critical account recovery mechanism.
// When users forget their password, they can request a secure, time-limited
// reset link sent to their verified email address.
func TestNewPasswordReset_CreatesValidInstance(t *testing.T) {
	userID := uuid.New()
	email := "player@example.com"
	ttlMinutes := 60

	reset, err := NewPasswordReset(userID, email, ttlMinutes)

	assert.NoError(t, err)
	assert.NotNil(t, reset)
	assert.NotEqual(t, uuid.Nil, reset.ID)
	assert.Equal(t, userID, reset.UserID)
	assert.Equal(t, email, reset.Email)
	assert.NotEmpty(t, reset.Token, "Should generate secure token")
	assert.Equal(t, PasswordResetStatusPending, reset.Status)
	assert.False(t, reset.CreatedAt.IsZero())
	assert.False(t, reset.UpdatedAt.IsZero())
	assert.False(t, reset.ExpiresAt.IsZero())
	assert.Nil(t, reset.UsedAt)
}

// TestNewPasswordReset_SetsCorrectExpiry validates expiration calculation
// Business context: Reset tokens are time-limited (typically 1 hour) to reduce
// the attack window if the email is compromised. Longer TTLs increase risk.
func TestNewPasswordReset_SetsCorrectExpiry(t *testing.T) {
	testCases := []struct {
		ttlMinutes int
	}{
		{15},
		{30},
		{60},
		{1440}, // 24 hours
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			before := time.Now().UTC()
			reset, err := NewPasswordReset(uuid.New(), "test@example.com", tc.ttlMinutes)

			assert.NoError(t, err)
			expectedExpiry := before.Add(time.Duration(tc.ttlMinutes) * time.Minute)
			assert.WithinDuration(t, expectedExpiry, reset.ExpiresAt, time.Second)
		})
	}
}

// TestNewPasswordReset_GeneratesUniqueTokens ensures token uniqueness
// Business context: Each reset request must have unique tokens to prevent
// token collision attacks where an attacker could reuse another user's token.
func TestNewPasswordReset_GeneratesUniqueTokens(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		reset, err := NewPasswordReset(uuid.New(), "test@example.com", 60)
		assert.NoError(t, err)
		assert.False(t, tokens[reset.Token], "Tokens should be unique")
		tokens[reset.Token] = true
	}
}

// =============================================================================
// STATUS METHODS TESTS
// =============================================================================

// TestPasswordReset_IsExpired_DetectsExpiration validates expiration detection
// Business context: Expired reset tokens must be rejected even if the token
// is otherwise valid. This limits the attack window for stolen tokens.
func TestPasswordReset_IsExpired_DetectsExpiration(t *testing.T) {
	tests := []struct {
		name     string
		offset   time.Duration
		expected bool
	}{
		{"future expiry", 1 * time.Hour, false},
		{"past expiry", -1 * time.Hour, true},
		{"just expired", -1 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
			reset.ExpiresAt = time.Now().UTC().Add(tt.offset)

			assert.Equal(t, tt.expected, reset.IsExpired())
		})
	}
}

// TestPasswordReset_IsPending_ReflectsCorrectState validates pending status
// Business context: Only pending (unused, not expired) reset requests should
// be valid for password changes. This prevents token reuse attacks.
func TestPasswordReset_IsPending_ReflectsCorrectState(t *testing.T) {
	// New reset should be pending
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
	assert.True(t, reset.IsPending())

	// Used reset should not be pending
	reset.MarkAsUsed()
	assert.False(t, reset.IsPending())
}

// TestPasswordReset_IsPending_FailsWhenExpired ensures expired resets are not pending
func TestPasswordReset_IsPending_FailsWhenExpired(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
	reset.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)

	assert.False(t, reset.IsPending(), "Expired reset should not be pending")
}

// TestIsUsed_ReflectsCorrectState validates used status detection
func TestIsUsed_ReflectsCorrectState(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)

	assert.False(t, reset.IsUsed())

	reset.MarkAsUsed()

	assert.True(t, reset.IsUsed())
}

// =============================================================================
// LIFECYCLE TESTS
// =============================================================================

// TestMarkAsUsed_SetsCorrectStatus validates marking reset as used
// Business context: Once a reset token is used to change the password,
// it must be immediately invalidated to prevent reuse attacks.
func TestMarkAsUsed_SetsCorrectStatus(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
	beforeUse := time.Now().UTC()

	reset.MarkAsUsed()

	assert.Equal(t, PasswordResetStatusUsed, reset.Status)
	assert.NotNil(t, reset.UsedAt)
	assert.True(t, reset.UsedAt.After(beforeUse) || reset.UsedAt.Equal(beforeUse))
	assert.True(t, reset.UpdatedAt.After(beforeUse) || reset.UpdatedAt.Equal(beforeUse))
	assert.True(t, reset.IsUsed())
	assert.False(t, reset.IsPending())
}

// TestMarkAsExpired_SetsCorrectStatus validates explicit expiration
// Business context: System can proactively expire reset tokens (e.g., when
// user requests a new reset, previous tokens should be expired).
func TestMarkAsExpired_SetsCorrectStatus(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
	beforeExpire := time.Now().UTC()

	reset.MarkAsExpired()

	assert.Equal(t, PasswordResetStatusExpired, reset.Status)
	assert.True(t, reset.UpdatedAt.After(beforeExpire) || reset.UpdatedAt.Equal(beforeExpire))
	assert.False(t, reset.IsPending())
}

// TestPasswordReset_Cancel_SetsCorrectStatus validates reset cancellation
// Business context: Users or admins may cancel pending reset requests
// (e.g., if user remembers password or suspects unauthorized request).
func TestPasswordReset_Cancel_SetsCorrectStatus(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
	beforeCancel := time.Now().UTC()

	reset.Cancel()

	assert.Equal(t, PasswordResetStatusCanceled, reset.Status)
	assert.True(t, reset.UpdatedAt.After(beforeCancel) || reset.UpdatedAt.Equal(beforeCancel))
	assert.False(t, reset.IsPending())
}

// =============================================================================
// AUDIT TESTS
// =============================================================================

// TestPasswordReset_SetRequestInfo_StoresAuditData validates audit data storage
// Business context: IP and user agent are recorded for security auditing
// and fraud detection. Unusual patterns (e.g., reset from foreign IP)
// can trigger additional security measures.
func TestPasswordReset_SetRequestInfo_StoresAuditData(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"
	beforeSet := time.Now().UTC()

	reset.SetRequestInfo(ip, ua)

	assert.Equal(t, ip, reset.IPAddress)
	assert.Equal(t, ua, reset.UserAgent)
	assert.True(t, reset.UpdatedAt.After(beforeSet) || reset.UpdatedAt.Equal(beforeSet))
}

// =============================================================================
// ACCESSOR TESTS
// =============================================================================

// TestPasswordReset_GetID_ReturnsCorrectID validates ID accessor
func TestPasswordReset_GetID_ReturnsCorrectID(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)

	assert.Equal(t, reset.ID, reset.GetID())
}

// =============================================================================
// TOKEN GENERATION TESTS
// =============================================================================

// TestGenerateResetToken_GeneratesCorrectLength validates token length
func TestGenerateResetToken_GeneratesCorrectLength(t *testing.T) {
	token, err := generateResetToken(32)

	assert.NoError(t, err)
	assert.Len(t, token, 64) // Hex encoding doubles length
}

// TestGenerateResetToken_GeneratesUniqueTokens validates randomness
func TestGenerateResetToken_GeneratesUniqueTokens(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, _ := generateResetToken(32)
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
	}
}

// =============================================================================
// CONSTANT TESTS
// =============================================================================

// TestPasswordResetStatus_Constants verifies status constants
func TestPasswordResetStatus_Constants(t *testing.T) {
	assert.Equal(t, PasswordResetStatus("pending"), PasswordResetStatusPending)
	assert.Equal(t, PasswordResetStatus("used"), PasswordResetStatusUsed)
	assert.Equal(t, PasswordResetStatus("expired"), PasswordResetStatusExpired)
	assert.Equal(t, PasswordResetStatus("canceled"), PasswordResetStatusCanceled)
}

// =============================================================================
// SECURITY SCENARIO TESTS
// =============================================================================

// TestPasswordReset_SecurityScenario_TokenReuse validates token cannot be reused
// Business context: After password reset, the token must be invalidated.
// Attempting to reuse a token (even within TTL) must fail.
func TestPasswordReset_SecurityScenario_TokenReuse(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)

	// First use - should work
	assert.True(t, reset.IsPending())
	reset.MarkAsUsed()
	assert.True(t, reset.IsUsed())

	// Attempt reuse - should fail (not pending anymore)
	assert.False(t, reset.IsPending())
}

// TestPasswordReset_SecurityScenario_ExpiredToken validates expired tokens are rejected
// Business context: Even with valid token format, expired tokens must be rejected
// to limit the attack window for compromised email accounts.
func TestPasswordReset_SecurityScenario_ExpiredToken(t *testing.T) {
	reset, _ := NewPasswordReset(uuid.New(), "test@example.com", 60)

	// Simulate expiration
	reset.ExpiresAt = time.Now().UTC().Add(-10 * time.Minute)

	// Token should be expired and not pending
	assert.True(t, reset.IsExpired())
	assert.False(t, reset.IsPending())
}

