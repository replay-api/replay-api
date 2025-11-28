package common

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ResourceOwner represents the owner of a resource.
type ResourceOwner struct {
	TenantID uuid.UUID `json:"tenant_id" bson:"tenant_id"` // TenantID represents the ID of the tenant the resource belongs to.
	ClientID uuid.UUID `json:"client_id" bson:"client_id"` // ClientID represents the ID of the client associated with the resource.
	GroupID  uuid.UUID `json:"group_id" bson:"group_id"`   // GroupID represents the ID of the group the resource is associated with. (redundant with ClientID ?)
	UserID   uuid.UUID `json:"user_id" bson:"user_id"`     // EndUserID represents the ID of the end user who owns the resource.
}

// adminAudienceKeys is an immutable map of audience keys that have admin privileges
// Using a function to return the map ensures thread-safety without global mutable state
var adminAudienceKeys = map[IntendedAudienceKey]bool{
	TenantAudienceIDKey:            true,
	ClientApplicationAudienceIDKey: true,
}

// IsAdmin checks if the current context has admin-level access
// Admin access is granted to TenantAudienceIDKey and ClientApplicationAudienceIDKey
func IsAdmin(userContext context.Context) bool {
	audience, ok := userContext.Value(AudienceKey).(IntendedAudienceKey)
	if !ok {
		return false
	}

	return adminAudienceKeys[audience]
}

// IsAuthenticated checks if the current context represents an authenticated user
func IsAuthenticated(ctx context.Context) bool {
	isAuth, ok := ctx.Value(AuthenticatedKey).(bool)
	return ok && isAuth
}

// CanAccessResource checks if the current user can access a resource based on visibility
func CanAccessResource(ctx context.Context, resourceOwner ResourceOwner, visibilityType VisibilityTypeKey) bool {
	currentOwner := GetResourceOwner(ctx)

	// Public resources are always accessible within the same tenant
	if visibilityType == PublicVisibilityTypeKey && currentOwner.TenantID == resourceOwner.TenantID {
		return true
	}

	// Must be authenticated for non-public resources
	if !IsAuthenticated(ctx) {
		return false
	}

	// User owns the resource
	if currentOwner.UserID != uuid.Nil && currentOwner.UserID == resourceOwner.UserID {
		return true
	}

	// User is in the same group
	if currentOwner.GroupID != uuid.Nil && currentOwner.GroupID == resourceOwner.GroupID {
		return true
	}

	// Restricted visibility - check client access
	if visibilityType == RestrictedVisibilityTypeKey {
		return currentOwner.ClientID != uuid.Nil && currentOwner.ClientID == resourceOwner.ClientID
	}

	return false
}

func GetResourceOwner(userContext context.Context) ResourceOwner {
	res := ResourceOwner{}

	if tenantID, ok := userContext.Value(TenantIDKey).(uuid.UUID); ok {
		res.TenantID = tenantID
	}

	if clientID, ok := userContext.Value(ClientIDKey).(uuid.UUID); ok {
		res.ClientID = clientID
	}

	if groupID, ok := userContext.Value(GroupIDKey).(uuid.UUID); ok {
		res.GroupID = groupID
	}

	if userID, ok := userContext.Value(UserIDKey).(uuid.UUID); ok {
		res.UserID = userID
	}

	if res.IsMissingTenant() {
		panic(fmt.Errorf("GetResourceOwner.IsMissingTenant: tenant_id missing in context %v", userContext))
	}

	return res
}

func (ro ResourceOwner) IsMissingTenant() bool {
	return ro.TenantID == uuid.Nil
}

func (ro ResourceOwner) IsTenant() bool {
	return ro.TenantID != uuid.Nil && ro.ClientID == uuid.Nil && ro.GroupID == uuid.Nil && ro.UserID == uuid.Nil
}

func (ro ResourceOwner) IsClient() bool {
	return ro.TenantID != uuid.Nil && ro.ClientID != uuid.Nil && ro.GroupID == uuid.Nil && ro.UserID == uuid.Nil
}

func (ro ResourceOwner) IsGroup() bool {
	return ro.TenantID != uuid.Nil && ro.GroupID != uuid.Nil && ro.UserID == uuid.Nil
}

func (ro ResourceOwner) IsUser() bool {
	return ro.TenantID != uuid.Nil && ro.UserID != uuid.Nil
}

// NewResourceOwner creates a ResourceOwner from the provided IDs
func NewResourceOwner(tenantID, clientID, groupID, userID uuid.UUID) ResourceOwner {
	return ResourceOwner{
		TenantID: tenantID,
		ClientID: clientID,
		GroupID:  groupID,
		UserID:   userID,
	}
}
