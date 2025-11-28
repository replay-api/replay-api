package iam_entities

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
)

// Tenant represents an organization or company using the platform
// Each tenant has a unique VHASH (Verification Hash) used as an API key
type Tenant struct {
	ID          uuid.UUID    `json:"id" bson:"_id"`
	Name        string       `json:"name" bson:"name"`
	Slug        string       `json:"slug" bson:"slug"` // URL-friendly identifier (e.g., "leetgaming-pro")
	Description string       `json:"description" bson:"description"`
	VHash       string       `json:"-" bson:"v_hash"` // Never expose in JSON responses
	VHashSalt   string       `json:"-" bson:"v_hash_salt"`
	Status      TenantStatus `json:"status" bson:"status"`
	Domain      string       `json:"domain" bson:"domain"`             // Primary domain (e.g., "leetgaming.pro")
	AllowedURLs []string     `json:"allowed_urls" bson:"allowed_urls"` // Allowed callback URLs
	CreatedAt   time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" bson:"updated_at"`
}

// NewTenant creates a new tenant with a generated VHASH
func NewTenant(name, slug, description, domain string, allowedURLs []string) (*Tenant, string) {
	tenantID := uuid.New()
	salt := generateSalt()
	vhash := generateVHash(tenantID.String(), salt)

	tenant := &Tenant{
		ID:          tenantID,
		Name:        name,
		Slug:        slug,
		Description: description,
		VHash:       vhash,
		VHashSalt:   salt,
		Status:      TenantStatusActive,
		Domain:      domain,
		AllowedURLs: allowedURLs,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Return tenant and the plain VHASH (only shown once at creation)
	return tenant, vhash
}

// NewTenantWithID creates a tenant with a specific ID (for seeding)
func NewTenantWithID(id uuid.UUID, name, slug, description, domain string, allowedURLs []string) (*Tenant, string) {
	salt := generateSalt()
	vhash := generateVHash(id.String(), salt)

	tenant := &Tenant{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: description,
		VHash:       vhash,
		VHashSalt:   salt,
		Status:      TenantStatusActive,
		Domain:      domain,
		AllowedURLs: allowedURLs,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	return tenant, vhash
}

// NewTenantWithVHash creates a tenant with a specific VHASH (for migration/import)
func NewTenantWithVHash(id uuid.UUID, name, slug, vhash, vhashSalt string) *Tenant {
	return &Tenant{
		ID:        id,
		Name:      name,
		Slug:      slug,
		VHash:     vhash,
		VHashSalt: vhashSalt,
		Status:    TenantStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// GetID returns the tenant ID
func (t *Tenant) GetID() uuid.UUID {
	return t.ID
}

// ValidateVHash validates if the provided VHASH matches the tenant's VHASH
func (t *Tenant) ValidateVHash(providedVHash string) bool {
	return t.VHash == providedVHash
}

// MaskVHash returns a masked version of the VHASH for display purposes
// Shows first 4 and last 4 characters with dots in between
func (t *Tenant) MaskVHash() string {
	return MaskSecret(t.VHash)
}

// RegenerateVHash creates a new VHASH for the tenant
// Returns the new plain VHASH (only available at regeneration time)
func (t *Tenant) RegenerateVHash() string {
	t.VHashSalt = generateSalt()
	t.VHash = generateVHash(t.ID.String(), t.VHashSalt)
	t.UpdatedAt = time.Now().UTC()
	return t.VHash
}

// generateSalt creates a random salt for VHASH generation
func generateSalt() string {
	saltID := uuid.New()
	hash := sha256.Sum256([]byte(saltID.String() + time.Now().UTC().String()))
	return hex.EncodeToString(hash[:16]) // 32 characters
}

// generateVHash creates a VHASH from tenant ID and salt
func generateVHash(tenantID, salt string) string {
	hash := sha256.Sum256([]byte(tenantID + ":" + salt))
	return hex.EncodeToString(hash[:])
}

// MaskSecret masks a secret string showing only first 4 and last 4 characters
// Uses Unicode filled circles (●) for masking
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "●●●●●●●●"
	}
	return secret[:4] + "●●●●●●●●" + secret[len(secret)-4:]
}

// MaskSecretFull returns a fully masked secret (no visible characters)
func MaskSecretFull(secret string) string {
	if secret == "" {
		return ""
	}
	masked := ""
	for i := 0; i < min(len(secret), 16); i++ {
		masked += "●"
	}
	return masked
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
