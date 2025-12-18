// Package auth_entities contains authentication domain entities
package auth_entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// PasswordResetStatus represents the status of a password reset request
type PasswordResetStatus string

const (
	PasswordResetStatusPending  PasswordResetStatus = "pending"
	PasswordResetStatusUsed     PasswordResetStatus = "used"
	PasswordResetStatusExpired  PasswordResetStatus = "expired"
	PasswordResetStatusCanceled PasswordResetStatus = "canceled"
)

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID          uuid.UUID           `json:"id" bson:"_id"`
	UserID      uuid.UUID           `json:"user_id" bson:"user_id"`
	Email       string              `json:"email" bson:"email"`
	Token       string              `json:"token" bson:"token"`
	Status      PasswordResetStatus `json:"status" bson:"status"`
	ExpiresAt   time.Time           `json:"expires_at" bson:"expires_at"`
	UsedAt      *time.Time          `json:"used_at,omitempty" bson:"used_at,omitempty"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
	IPAddress   string              `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent   string              `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
}

// NewPasswordReset creates a new password reset request
func NewPasswordReset(userID uuid.UUID, email string, ttlMinutes int) (*PasswordReset, error) {
	token, err := generateResetToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &PasswordReset{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		Status:    PasswordResetStatusPending,
		ExpiresAt: now.Add(time.Duration(ttlMinutes) * time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// IsExpired checks if the reset request has expired
func (r *PasswordReset) IsExpired() bool {
	return time.Now().UTC().After(r.ExpiresAt)
}

// IsPending checks if the reset request is still pending
func (r *PasswordReset) IsPending() bool {
	return r.Status == PasswordResetStatusPending && !r.IsExpired()
}

// IsUsed checks if the reset request has been used
func (r *PasswordReset) IsUsed() bool {
	return r.Status == PasswordResetStatusUsed
}

// MarkAsUsed marks the reset request as used
func (r *PasswordReset) MarkAsUsed() {
	now := time.Now().UTC()
	r.Status = PasswordResetStatusUsed
	r.UsedAt = &now
	r.UpdatedAt = now
}

// MarkAsExpired marks the reset request as expired
func (r *PasswordReset) MarkAsExpired() {
	r.Status = PasswordResetStatusExpired
	r.UpdatedAt = time.Now().UTC()
}

// Cancel cancels the reset request
func (r *PasswordReset) Cancel() {
	r.Status = PasswordResetStatusCanceled
	r.UpdatedAt = time.Now().UTC()
}

// SetRequestInfo sets the IP address and user agent for audit purposes
func (r *PasswordReset) SetRequestInfo(ipAddress, userAgent string) {
	r.IPAddress = ipAddress
	r.UserAgent = userAgent
	r.UpdatedAt = time.Now().UTC()
}

// GetID returns the reset request ID (implements common.Entity)
func (r PasswordReset) GetID() uuid.UUID {
	return r.ID
}

// generateResetToken generates a cryptographically secure random token
func generateResetToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

