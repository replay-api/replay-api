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

var AdminAudienceIDKeys = make(map[IntendedAudienceKey]bool, 0)

func IsAdmin(userContext context.Context) bool {
	if len(AdminAudienceIDKeys) == 0 {
		AdminAudienceIDKeys[TenantAudienceIDKey] = true
		AdminAudienceIDKeys[ClientApplicationAudienceIDKey] = true
	}

	audience, ok := userContext.Value(AudienceKey).(IntendedAudienceKey)
	if !ok {
		return false
	}

	if _, ok := AdminAudienceIDKeys[audience]; !ok {
		return false
	}

	return AdminAudienceIDKeys[audience]

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
