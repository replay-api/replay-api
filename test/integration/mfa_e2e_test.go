//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shared "github.com/resource-ownership/go-common/pkg/common"
	auth_entities "github.com/replay-api/replay-api/pkg/domain/auth/entities"
)

// TestE2E_MFAEntities tests MFA entity creation and lifecycle
func TestE2E_MFAEntities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Test user setup
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	t.Run("MFAMethod_Constants", func(t *testing.T) {
		assert.Equal(t, auth_entities.MFAMethod("email"), auth_entities.MFAMethodEmail)
		assert.Equal(t, auth_entities.MFAMethod("totp"), auth_entities.MFAMethodTOTP)
		assert.Equal(t, auth_entities.MFAMethod("sms"), auth_entities.MFAMethodSMS)

		t.Log("✓ MFA method constants defined correctly")
	})

	t.Run("MFASessionStatus_Constants", func(t *testing.T) {
		assert.Equal(t, auth_entities.MFASessionStatus("pending"), auth_entities.MFASessionStatusPending)
		assert.Equal(t, auth_entities.MFASessionStatus("verified"), auth_entities.MFASessionStatusVerified)
		assert.Equal(t, auth_entities.MFASessionStatus("expired"), auth_entities.MFASessionStatusExpired)
		assert.Equal(t, auth_entities.MFASessionStatus("canceled"), auth_entities.MFASessionStatusCanceled)

		t.Log("✓ MFA session status constants defined correctly")
	})

	t.Run("NewMFASettings_Creation", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)

		require.NotNil(t, settings)
		assert.NotEqual(t, uuid.Nil, settings.ID)
		assert.Equal(t, userID, settings.UserID)
		assert.False(t, settings.Enabled)
		assert.Equal(t, auth_entities.MFAMethodEmail, settings.PrimaryMethod)
		assert.Contains(t, settings.EnabledMethods, auth_entities.MFAMethodEmail)
		assert.False(t, settings.CreatedAt.IsZero())
		assert.False(t, settings.UpdatedAt.IsZero())

		t.Logf("✓ MFA settings created: ID=%s", settings.ID)
	})

	t.Run("MFASettings_EnableMFA", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)

		assert.False(t, settings.Enabled)

		settings.EnableMFA(auth_entities.MFAMethodTOTP)

		assert.True(t, settings.Enabled)
		assert.Equal(t, auth_entities.MFAMethodTOTP, settings.PrimaryMethod)
		assert.True(t, settings.HasMethod(auth_entities.MFAMethodTOTP))

		t.Log("✓ EnableMFA() works correctly")
	})

	t.Run("MFASettings_DisableMFA", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)
		settings.EnableMFA(auth_entities.MFAMethodTOTP)

		assert.True(t, settings.Enabled)

		settings.DisableMFA()

		assert.False(t, settings.Enabled)

		t.Log("✓ DisableMFA() works correctly")
	})

	t.Run("MFASettings_HasMethod", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)

		// Initially has email method
		assert.True(t, settings.HasMethod(auth_entities.MFAMethodEmail))
		assert.False(t, settings.HasMethod(auth_entities.MFAMethodTOTP))

		// Add TOTP method
		settings.SetTOTPSecret("test-secret")
		assert.True(t, settings.HasMethod(auth_entities.MFAMethodTOTP))

		t.Log("✓ HasMethod() works correctly")
	})

	t.Run("MFASettings_SetTOTPSecret", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)

		assert.Empty(t, settings.TOTPSecret)
		assert.False(t, settings.HasMethod(auth_entities.MFAMethodTOTP))

		settings.SetTOTPSecret("JBSWY3DPEHPK3PXP")

		assert.Equal(t, "JBSWY3DPEHPK3PXP", settings.TOTPSecret)
		assert.True(t, settings.HasMethod(auth_entities.MFAMethodTOTP))

		t.Log("✓ SetTOTPSecret() works correctly")
	})

	t.Run("MFASettings_GenerateRecoveryCodes", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)

		codes, err := settings.GenerateRecoveryCodes()
		require.NoError(t, err)

		assert.Len(t, codes, 10)
		assert.Len(t, settings.RecoveryCodes, 10)

		// Verify codes are unique
		codeSet := make(map[string]bool)
		for _, code := range codes {
			assert.False(t, codeSet[code], "Recovery codes should be unique")
			codeSet[code] = true
		}

		t.Log("✓ GenerateRecoveryCodes() generates 10 unique codes")
	})

	t.Run("MFASettings_UseRecoveryCode", func(t *testing.T) {
		settings := auth_entities.NewMFASettings(userID)
		codes, _ := settings.GenerateRecoveryCodes()

		// Use a code
		codeToUse := codes[0]
		used := settings.UseRecoveryCode(codeToUse)

		assert.True(t, used)
		assert.Len(t, settings.RecoveryCodes, 9) // One less

		// Try to use same code again
		usedAgain := settings.UseRecoveryCode(codeToUse)
		assert.False(t, usedAgain)

		// Try invalid code
		invalidUsed := settings.UseRecoveryCode("invalid-code")
		assert.False(t, invalidUsed)

		t.Log("✓ UseRecoveryCode() handles codes correctly")
	})

	t.Run("NewMFASession_Creation", func(t *testing.T) {
		verificationID := uuid.New()
		ttlMinutes := 10

		session := auth_entities.NewMFASession(userID, verificationID, auth_entities.MFAMethodTOTP, ttlMinutes)

		require.NotNil(t, session)
		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, verificationID, session.VerificationID)
		assert.Equal(t, auth_entities.MFAMethodTOTP, session.Method)
		assert.Equal(t, auth_entities.MFASessionStatusPending, session.Status)
		assert.Equal(t, 0, session.Attempts)
		assert.Equal(t, 5, session.MaxAttempts)
		assert.True(t, session.ExpiresAt.After(time.Now()))
		assert.Nil(t, session.VerifiedAt)

		t.Logf("✓ MFA session created: ID=%s", session.ID)
	})

	t.Run("MFASession_MarkVerified", func(t *testing.T) {
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)

		assert.Equal(t, auth_entities.MFASessionStatusPending, session.Status)
		assert.Nil(t, session.VerifiedAt)

		session.MarkVerified()

		assert.Equal(t, auth_entities.MFASessionStatusVerified, session.Status)
		assert.NotNil(t, session.VerifiedAt)
		assert.True(t, session.IsVerified())

		t.Log("✓ MarkVerified() works correctly")
	})

	t.Run("MFASession_IsExpired", func(t *testing.T) {
		// Create non-expired session
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)
		assert.False(t, session.IsExpired())

		// Create expired session
		session.ExpiresAt = time.Now().Add(-1 * time.Minute)
		assert.True(t, session.IsExpired())

		t.Log("✓ IsExpired() works correctly")
	})

	t.Run("MFASession_IsPending", func(t *testing.T) {
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)

		assert.True(t, session.IsPending())

		// Verify it
		session.MarkVerified()
		assert.False(t, session.IsPending())

		// Test expired session
		expiredSession := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)
		expiredSession.ExpiresAt = time.Now().Add(-1 * time.Minute)
		assert.False(t, expiredSession.IsPending())

		t.Log("✓ IsPending() works correctly")
	})

	t.Run("MFASession_IncrementAttempts", func(t *testing.T) {
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)

		assert.Equal(t, 0, session.Attempts)
		assert.Equal(t, 5, session.RemainingAttempts())

		session.IncrementAttempts()
		assert.Equal(t, 1, session.Attempts)
		assert.Equal(t, 4, session.RemainingAttempts())

		session.IncrementAttempts()
		session.IncrementAttempts()
		session.IncrementAttempts()
		session.IncrementAttempts()
		assert.Equal(t, 5, session.Attempts)
		assert.Equal(t, 0, session.RemainingAttempts())

		t.Log("✓ IncrementAttempts() and RemainingAttempts() work correctly")
	})

	t.Run("MFASession_SetRequestInfo", func(t *testing.T) {
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)

		assert.Empty(t, session.IPAddress)
		assert.Empty(t, session.UserAgent)

		session.SetRequestInfo("192.168.1.1", "Mozilla/5.0")

		assert.Equal(t, "192.168.1.1", session.IPAddress)
		assert.Equal(t, "Mozilla/5.0", session.UserAgent)

		t.Log("✓ SetRequestInfo() works correctly")
	})

	t.Run("MFASession_SetPendingAction", func(t *testing.T) {
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 10)

		assert.Empty(t, session.PendingAction)
		assert.Nil(t, session.PendingActionData)

		session.SetPendingAction("withdraw", map[string]any{
			"amount":   100.00,
			"currency": "USD",
		})

		assert.Equal(t, "withdraw", session.PendingAction)
		assert.Equal(t, 100.00, session.PendingActionData["amount"])
		assert.Equal(t, "USD", session.PendingActionData["currency"])

		t.Log("✓ SetPendingAction() works correctly")
	})

	t.Run("MFASession_TTL", func(t *testing.T) {
		testCases := []struct {
			ttlMinutes int
			expected   time.Duration
		}{
			{5, 5 * time.Minute},
			{10, 10 * time.Minute},
			{30, 30 * time.Minute},
		}

		for _, tc := range testCases {
			session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, tc.ttlMinutes)

			expectedExpiry := time.Now().Add(tc.expected)
			assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, time.Second,
				"TTL of %d minutes should expire ~%v from now", tc.ttlMinutes, tc.expected)
		}

		t.Log("✓ Session TTL calculated correctly")
	})

	t.Log("✓ All MFA E2E tests passed!")
}

// TestE2E_MFAWorkflow tests the complete MFA workflow
func TestE2E_MFAWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	userID := uuid.New()

	t.Run("Complete_MFA_Setup_Workflow", func(t *testing.T) {
		// Step 1: Create MFA settings for user
		settings := auth_entities.NewMFASettings(userID)
		assert.False(t, settings.Enabled)
		t.Log("Step 1: MFA settings created (disabled)")

		// Step 2: Set TOTP secret (simulating QR code setup)
		totpSecret := "JBSWY3DPEHPK3PXP"
		settings.SetTOTPSecret(totpSecret)
		assert.Equal(t, totpSecret, settings.TOTPSecret)
		t.Log("Step 2: TOTP secret set")

		// Step 3: Generate recovery codes
		codes, err := settings.GenerateRecoveryCodes()
		require.NoError(t, err)
		assert.Len(t, codes, 10)
		t.Log("Step 3: Recovery codes generated")

		// Step 4: Enable MFA with TOTP
		settings.EnableMFA(auth_entities.MFAMethodTOTP)
		assert.True(t, settings.Enabled)
		assert.Equal(t, auth_entities.MFAMethodTOTP, settings.PrimaryMethod)
		t.Log("Step 4: MFA enabled with TOTP")

		// Step 5: Create MFA session for verification
		session := auth_entities.NewMFASession(userID, uuid.New(), auth_entities.MFAMethodTOTP, 5)
		assert.True(t, session.IsPending())
		t.Log("Step 5: MFA session created for verification")

		// Step 6: Simulate verification
		session.MarkVerified()
		assert.True(t, session.IsVerified())
		now := time.Now()
		settings.LastVerifiedAt = &now
		t.Log("Step 6: MFA verified successfully")

		t.Log("✓ Complete MFA setup workflow passed!")
	})

	t.Run("MFA_Recovery_Workflow", func(t *testing.T) {
		// Setup user with MFA
		settings := auth_entities.NewMFASettings(userID)
		settings.SetTOTPSecret("SECRET")
		codes, _ := settings.GenerateRecoveryCodes()
		settings.EnableMFA(auth_entities.MFAMethodTOTP)

		// User loses device, needs to use recovery code
		recoveryCode := codes[0]
		usedSuccessfully := settings.UseRecoveryCode(recoveryCode)
		assert.True(t, usedSuccessfully)
		assert.Len(t, settings.RecoveryCodes, 9)

		// Cannot reuse same code
		usedAgain := settings.UseRecoveryCode(recoveryCode)
		assert.False(t, usedAgain)

		t.Log("✓ MFA recovery workflow passed!")
	})

	t.Log("✓ All MFA workflow tests passed!")
}

