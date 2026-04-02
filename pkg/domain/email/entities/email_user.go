package email_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type EmailUser struct {
	shared.BaseEntity    `json:",inline" bson:",inline"`
	VHash                string     `json:"v_hash" bson:"v_hash"`
	Email                string     `json:"email" bson:"email"`
	PasswordHash         string     `json:"-" bson:"password_hash"`
	EmailVerified        bool       `json:"email_verified" bson:"email_verified"`
	DisplayName          string     `json:"display_name" bson:"display_name"`
	FailedLoginAttempts  int        `json:"failed_login_attempts" bson:"failed_login_attempts"`
	LockedUntil          *time.Time `json:"locked_until,omitempty" bson:"locked_until,omitempty"`
}

func NewEmailUser(vHash, email, passwordHash, displayName string, emailVerified bool, resourceOwner shared.ResourceOwner) *EmailUser {
	entity := shared.NewEntity(resourceOwner)
	return &EmailUser{
		BaseEntity:           entity,
		VHash:                vHash,
		Email:                email,
		PasswordHash:         passwordHash,
		EmailVerified:        emailVerified,
		DisplayName:          displayName,
		FailedLoginAttempts:  0,
		LockedUntil:          nil,
	}
}

func (e EmailUser) GetID() uuid.UUID {
	return e.ID
}

// IsLocked returns true if the account is currently locked
func (e EmailUser) IsLocked() bool {
	if e.LockedUntil == nil {
		return false
	}
	return e.LockedUntil.After(time.Now())
}

// RecordFailedLogin increments failed login attempts and locks account if threshold reached
func (e *EmailUser) RecordFailedLogin(maxAttempts int, lockoutDuration time.Duration) {
	e.FailedLoginAttempts++
	if e.FailedLoginAttempts >= maxAttempts {
		lockoutTime := time.Now().Add(lockoutDuration)
		e.LockedUntil = &lockoutTime
	}
}

// ResetFailedLoginAttempts clears failed login attempts and unlocks account
func (e *EmailUser) ResetFailedLoginAttempts() {
	e.FailedLoginAttempts = 0
	e.LockedUntil = nil
}
