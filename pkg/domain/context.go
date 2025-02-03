package common

type ContextKey string

const (
	// Tenancy (internal)
	TenantIDKey ContextKey = "tenant_id"
	ClientIDKey ContextKey = "client_id"
	GroupIDKey  ContextKey = "group_id"
	UserIDKey   ContextKey = "user_id"

	// Parameters
	GameIDParamKey  ContextKey = "game_id"
	MatchIDParamKey ContextKey = "match_id"

	AudienceKey      ContextKey = "aud"
	AuthenticatedKey ContextKey = "auth"
)
