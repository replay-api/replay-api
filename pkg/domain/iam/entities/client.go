package iam_entities

import (
	"time"

	"github.com/google/uuid"
)

// ClientType represents the type of client application
type ClientType string

const (
	ClientTypeWeb      ClientType = "web"      // Web frontend application
	ClientTypeMobile   ClientType = "mobile"   // Mobile application
	ClientTypeDesktop  ClientType = "desktop"  // Desktop application
	ClientTypeAPI      ClientType = "api"      // API/Backend service
	ClientTypeBot      ClientType = "bot"      // Bot/Automation
	ClientTypeInternal ClientType = "internal" // Internal service
)

// ClientStatus represents the status of a client
type ClientStatus string

const (
	ClientStatusActive    ClientStatus = "active"
	ClientStatusSuspended ClientStatus = "suspended"
	ClientStatusPending   ClientStatus = "pending"
)

// Client represents an application that belongs to a tenant
// Example: leetgaming-pro-web is a Client of the LeetGaming tenant
type Client struct {
	ID               uuid.UUID    `json:"id" bson:"_id"`
	TenantID         uuid.UUID    `json:"tenant_id" bson:"tenant_id"`
	Name             string       `json:"name" bson:"name"`
	Slug             string       `json:"slug" bson:"slug"`        // URL-friendly identifier (e.g., "leetgaming-pro-web")
	Description      string       `json:"description" bson:"description"`
	Type             ClientType   `json:"type" bson:"type"`
	Status           ClientStatus `json:"status" bson:"status"`
	ClientSecret     string       `json:"-" bson:"client_secret"`                   // Never expose in JSON
	AllowedOrigins   []string     `json:"allowed_origins" bson:"allowed_origins"`   // CORS origins
	AllowedCallbacks []string     `json:"allowed_callbacks" bson:"allowed_callbacks"` // OAuth callback URLs
	CreatedAt        time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" bson:"updated_at"`
}

// NewClient creates a new client application for a tenant
func NewClient(tenantID uuid.UUID, name, slug, description string, clientType ClientType, allowedOrigins, allowedCallbacks []string) (*Client, string) {
	clientID := uuid.New()
	clientSecret := generateClientSecret()

	client := &Client{
		ID:               clientID,
		TenantID:         tenantID,
		Name:             name,
		Slug:             slug,
		Description:      description,
		Type:             clientType,
		Status:           ClientStatusActive,
		ClientSecret:     clientSecret,
		AllowedOrigins:   allowedOrigins,
		AllowedCallbacks: allowedCallbacks,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	return client, clientSecret
}

// NewClientWithID creates a client with a specific ID (for seeding)
func NewClientWithID(id, tenantID uuid.UUID, name, slug, description string, clientType ClientType, allowedOrigins, allowedCallbacks []string) (*Client, string) {
	clientSecret := generateClientSecret()

	client := &Client{
		ID:               id,
		TenantID:         tenantID,
		Name:             name,
		Slug:             slug,
		Description:      description,
		Type:             clientType,
		Status:           ClientStatusActive,
		ClientSecret:     clientSecret,
		AllowedOrigins:   allowedOrigins,
		AllowedCallbacks: allowedCallbacks,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	return client, clientSecret
}

// NewClientWithSecret creates a client with a specific secret (for migration/import)
func NewClientWithSecret(id, tenantID uuid.UUID, name, slug, clientSecret string, clientType ClientType) *Client {
	return &Client{
		ID:           id,
		TenantID:     tenantID,
		Name:         name,
		Slug:         slug,
		Type:         clientType,
		Status:       ClientStatusActive,
		ClientSecret: clientSecret,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

// GetID returns the client ID
func (c *Client) GetID() uuid.UUID {
	return c.ID
}

// ValidateSecret validates if the provided secret matches the client's secret
func (c *Client) ValidateSecret(providedSecret string) bool {
	return c.ClientSecret == providedSecret
}

// MaskSecret returns a masked version of the client secret for display
func (c *Client) MaskSecret() string {
	return MaskSecret(c.ClientSecret)
}

// RegenerateSecret creates a new secret for the client
// Returns the new plain secret (only available at regeneration time)
func (c *Client) RegenerateSecret() string {
	c.ClientSecret = generateClientSecret()
	c.UpdatedAt = time.Now().UTC()
	return c.ClientSecret
}

// generateClientSecret creates a random client secret
func generateClientSecret() string {
	return generateVHash(uuid.New().String(), generateSalt())
}
