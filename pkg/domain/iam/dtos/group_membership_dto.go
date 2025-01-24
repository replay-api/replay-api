package iam_dtos

import (
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type GroupMembershipDTO struct {
	Group      iam_entities.Group      `json:"group"`
	Membership iam_entities.Membership `json:"membership"`
}
