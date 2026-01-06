// Package auth_entities contains authentication domain entities
package auth_entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// VerificationStatus represents the status of an email verification
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusExpired  VerificationStatus = "expired"
	VerificationStatusCanceled VerificationStatus = "canceled"
)

// VerificationType represents the type of verification
type VerificationType string

const (
	VerificationTypeEmail VerificationType = "email"
	VerificationTypeMFA   VerificationType = "mfa"
)

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID         uuid.UUID          `json:"id" bson:"_id"`
	UserID     uuid.UUID          `json:"user_id" bson:"user_id"`
	Email      string             `json:"email" bson:"email"`
	Token      string             `json:"token" bson:"token"`
	Code       string             `json:"code" bson:"code"` // 6-digit code for MFA
	Type       VerificationType   `json:"type" bson:"type"`
	Status     VerificationStatus `json:"status" bson:"status"`
	Attempts   int                `json:"attempts" bson:"attempts"`
	MaxAttempts int               `json:"max_attempts" bson:"max_attempts"`
	ExpiresAt  time.Time          `json:"expires_at" bson:"expires_at"`
	VerifiedAt *time.Time         `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
	IPAddress  string             `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent  string             `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
}

// NewEmailVerification creates a new email verification token
func NewEmailVerification(userID uuid.UUID, email string, verificationType VerificationType, ttlMinutes int) (*EmailVerification, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	code, err := generateNumericCode(6)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &EmailVerification{
		ID:         uuid.New(),
		UserID:     userID,
		Email:      email,
		Token:      token,
		Code:       code,
		Type:       verificationType,
		Status:     VerificationStatusPending,
		Attempts:   0,
		MaxAttempts: 5,
		ExpiresAt:  now.Add(time.Duration(ttlMinutes) * time.Minute),
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Verify attempts to verify with the provided code or token
func (v *EmailVerification) Verify(codeOrToken string) bool {
	if v.IsExpired() {
		v.Status = VerificationStatusExpired
		return false
	}

	if v.Attempts >= v.MaxAttempts {
		return false
	}

	v.Attempts++
	v.UpdatedAt = time.Now().UTC()

	// Check both code and token
	if v.Code == codeOrToken || v.Token == codeOrToken {
		now := time.Now().UTC()
		v.Status = VerificationStatusVerified
		v.VerifiedAt = &now
		return true
	}

	return false
}

// IsExpired checks if the verification has expired
func (v *EmailVerification) IsExpired() bool {
	return time.Now().UTC().After(v.ExpiresAt)
}

// IsPending checks if the verification is still pending
func (v *EmailVerification) IsPending() bool {
	return v.Status == VerificationStatusPending && !v.IsExpired()
}

// IsVerified checks if the verification is complete
func (v *EmailVerification) IsVerified() bool {
	return v.Status == VerificationStatusVerified
}

// Cancel cancels the verification
func (v *EmailVerification) Cancel() {
	v.Status = VerificationStatusCanceled
	v.UpdatedAt = time.Now().UTC()
}

// RemainingAttempts returns the number of remaining verification attempts
func (v *EmailVerification) RemainingAttempts() int {
	remaining := v.MaxAttempts - v.Attempts
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SetRequestInfo sets the IP address and user agent for audit purposes
func (v *EmailVerification) SetRequestInfo(ipAddress, userAgent string) {
	v.IPAddress = ipAddress
	v.UserAgent = userAgent
	v.UpdatedAt = time.Now().UTC()
}

// GetID returns the verification ID (implements shared.Entity)
func (v EmailVerification) GetID() uuid.UUID {
	return v.ID
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateNumericCode generates a random numeric code of specified length
func generateNumericCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		code[i] = digits[bytes[i]%10]
	}

	return string(code), nil
}
