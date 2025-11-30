// Package auth_entities contains authentication domain entities
package auth_entities

import (
	"time"

	"github.com/google/uuid"
)

// MFAMethod represents the method of MFA
type MFAMethod string

const (
	MFAMethodEmail MFAMethod = "email"
	MFAMethodTOTP  MFAMethod = "totp"
	MFAMethodSMS   MFAMethod = "sms"  // Future support
)

// MFASessionStatus represents the status of an MFA session
type MFASessionStatus string

const (
	MFASessionStatusPending   MFASessionStatus = "pending"
	MFASessionStatusVerified  MFASessionStatus = "verified"
	MFASessionStatusExpired   MFASessionStatus = "expired"
	MFASessionStatusCanceled  MFASessionStatus = "canceled"
)

// MFASettings represents a user's MFA configuration
type MFASettings struct {
	ID              uuid.UUID   `json:"id" bson:"_id"`
	UserID          uuid.UUID   `json:"user_id" bson:"user_id"`
	Enabled         bool        `json:"enabled" bson:"enabled"`
	PrimaryMethod   MFAMethod   `json:"primary_method" bson:"primary_method"`
	EnabledMethods  []MFAMethod `json:"enabled_methods" bson:"enabled_methods"`
	BackupEmail     string      `json:"backup_email,omitempty" bson:"backup_email,omitempty"`
	TOTPSecret      string      `json:"-" bson:"totp_secret,omitempty"` // Never expose in JSON
	RecoveryCodes   []string    `json:"-" bson:"recovery_codes,omitempty"`
	CreatedAt       time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" bson:"updated_at"`
	LastVerifiedAt  *time.Time  `json:"last_verified_at,omitempty" bson:"last_verified_at,omitempty"`
}

// MFASession represents an active MFA session requiring verification
type MFASession struct {
	ID                uuid.UUID        `json:"id" bson:"_id"`
	UserID            uuid.UUID        `json:"user_id" bson:"user_id"`
	VerificationID    uuid.UUID        `json:"verification_id" bson:"verification_id"` // Links to EmailVerification
	Method            MFAMethod        `json:"method" bson:"method"`
	Status            MFASessionStatus `json:"status" bson:"status"`
	Attempts          int              `json:"attempts" bson:"attempts"`
	MaxAttempts       int              `json:"max_attempts" bson:"max_attempts"`
	ExpiresAt         time.Time        `json:"expires_at" bson:"expires_at"`
	VerifiedAt        *time.Time       `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	CreatedAt         time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" bson:"updated_at"`
	IPAddress         string           `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent         string           `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	PendingAction     string           `json:"pending_action,omitempty" bson:"pending_action,omitempty"` // What action requires MFA
	PendingActionData map[string]any   `json:"pending_action_data,omitempty" bson:"pending_action_data,omitempty"`
}

// NewMFASettings creates new MFA settings for a user
func NewMFASettings(userID uuid.UUID) *MFASettings {
	now := time.Now().UTC()
	return &MFASettings{
		ID:             uuid.New(),
		UserID:         userID,
		Enabled:        false,
		PrimaryMethod:  MFAMethodEmail, // Default to email
		EnabledMethods: []MFAMethod{MFAMethodEmail},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewMFASession creates a new MFA session
func NewMFASession(userID, verificationID uuid.UUID, method MFAMethod, ttlMinutes int) *MFASession {
	now := time.Now().UTC()
	return &MFASession{
		ID:             uuid.New(),
		UserID:         userID,
		VerificationID: verificationID,
		Method:         method,
		Status:         MFASessionStatusPending,
		Attempts:       0,
		MaxAttempts:    5,
		ExpiresAt:      now.Add(time.Duration(ttlMinutes) * time.Minute),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// EnableMFA enables MFA for the user with the specified method
func (s *MFASettings) EnableMFA(method MFAMethod) {
	s.Enabled = true
	s.PrimaryMethod = method
	if !s.HasMethod(method) {
		s.EnabledMethods = append(s.EnabledMethods, method)
	}
	s.UpdatedAt = time.Now().UTC()
}

// DisableMFA disables MFA for the user
func (s *MFASettings) DisableMFA() {
	s.Enabled = false
	s.UpdatedAt = time.Now().UTC()
}

// HasMethod checks if a method is enabled
func (s *MFASettings) HasMethod(method MFAMethod) bool {
	for _, m := range s.EnabledMethods {
		if m == method {
			return true
		}
	}
	return false
}

// SetTOTPSecret sets the TOTP secret for the user
func (s *MFASettings) SetTOTPSecret(secret string) {
	s.TOTPSecret = secret
	if !s.HasMethod(MFAMethodTOTP) {
		s.EnabledMethods = append(s.EnabledMethods, MFAMethodTOTP)
	}
	s.UpdatedAt = time.Now().UTC()
}

// GenerateRecoveryCodes generates new recovery codes
func (s *MFASettings) GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := generateSecureToken(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	s.RecoveryCodes = codes
	s.UpdatedAt = time.Now().UTC()
	return codes, nil
}

// UseRecoveryCode uses a recovery code and marks it as used
func (s *MFASettings) UseRecoveryCode(code string) bool {
	for i, c := range s.RecoveryCodes {
		if c == code {
			// Remove used code
			s.RecoveryCodes = append(s.RecoveryCodes[:i], s.RecoveryCodes[i+1:]...)
			s.UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

// MarkVerified marks the MFA session as verified
func (m *MFASession) MarkVerified() {
	now := time.Now().UTC()
	m.Status = MFASessionStatusVerified
	m.VerifiedAt = &now
	m.UpdatedAt = now
}

// IsExpired checks if the session has expired
func (m *MFASession) IsExpired() bool {
	return time.Now().UTC().After(m.ExpiresAt)
}

// IsPending checks if the session is still pending
func (m *MFASession) IsPending() bool {
	return m.Status == MFASessionStatusPending && !m.IsExpired()
}

// IsVerified checks if the session is verified
func (m *MFASession) IsVerified() bool {
	return m.Status == MFASessionStatusVerified
}

// IncrementAttempts increments the attempt counter
func (m *MFASession) IncrementAttempts() {
	m.Attempts++
	m.UpdatedAt = time.Now().UTC()
}

// RemainingAttempts returns remaining verification attempts
func (m *MFASession) RemainingAttempts() int {
	remaining := m.MaxAttempts - m.Attempts
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SetRequestInfo sets audit information
func (m *MFASession) SetRequestInfo(ipAddress, userAgent string) {
	m.IPAddress = ipAddress
	m.UserAgent = userAgent
	m.UpdatedAt = time.Now().UTC()
}

// SetPendingAction sets the action that requires MFA
func (m *MFASession) SetPendingAction(action string, data map[string]any) {
	m.PendingAction = action
	m.PendingActionData = data
	m.UpdatedAt = time.Now().UTC()
}
