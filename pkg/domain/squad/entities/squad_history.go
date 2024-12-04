package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type SquadHistoryAction string

const (
	SquadCreated               SquadHistoryAction = "Squad Created"
	SquadUpdated               SquadHistoryAction = "Squad Updated"
	SquadMemberAdded           SquadHistoryAction = "Squad Member Added"
	SquadMemberLeft            SquadHistoryAction = "Squad Member Left"
	SquadMemberRemoved         SquadHistoryAction = "Squad Member Removed"
	SquadMemberPromoted        SquadHistoryAction = "Squad Member Promoted"
	SquadMemberDemoted         SquadHistoryAction = "Squad Member Demoted"
	SquadMemberJoined          SquadHistoryAction = "Squad Member Joined"
	SquadMemberJoinRequest     SquadHistoryAction = "User Request to Join Squad"
	SquadMemberRequestDeclined SquadHistoryAction = "Request to Join Squad Declined"
	SquadMemberRequestAccepted SquadHistoryAction = "Request to Join Squad Accepted"
	SquadDeleted               SquadHistoryAction = "Squad Deleted"
)

type SquadHistory struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	SquadID       uuid.UUID            `json:"squad_id" bson:"squad_id"`
	UserID        uuid.UUID            `json:"user_id" bson:"user_id"`
	Role          SquadRole            `json:"role" bson:"role"`
	Action        SquadHistoryAction   `json:"action" bson:"action"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
}
