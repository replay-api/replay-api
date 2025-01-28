package iam_query_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_dtos "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
)

type MembershipQueryService struct {
	common.BaseQueryService[iam_entities.Membership]
	GroupReader iam_out.GroupReader
}

func NewMembershipQueryService(MembershipReader common.Searchable[iam_entities.Membership], groupReader iam_out.GroupReader) *MembershipQueryService {
	queryableFields := map[string]bool{
		"ID":            true,
		"Type":          common.ALLOW,
		"ResourceOwner": common.ALLOW,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"Type":          true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &MembershipQueryService{
		BaseQueryService: common.BaseQueryService[iam_entities.Membership]{
			Reader:          MembershipReader,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.UserAudienceIDKey,
		},
		GroupReader: groupReader,
	}
}

func (s *MembershipQueryService) ListMemberGroups(ctx context.Context, search *common.Search) (map[uuid.UUID]iam_dtos.GroupMembershipDTO, error) {
	userID, ok := ctx.Value(common.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, fmt.Errorf("invalid (uuid.Nil) user ID in context")
	}

	searchAggregations := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field:    "ResourceOwner.UserID",
							Operator: common.EqualsOperator,
							Values:   []interface{}{userID},
						},
					},
				},
			},
		},
	}

	resultOptions := common.SearchResultOptions{
		Limit: 100,
	}

	search.SearchParams = append(search.SearchParams, searchAggregations...)
	search.ResultOptions = resultOptions

	if search.VisibilityOptions.IntendedAudience == 0 {
		search.VisibilityOptions = common.SearchVisibilityOptions{
			RequestSource:    common.GetResourceOwner(ctx),
			IntendedAudience: common.UserAudienceIDKey,
		}
	}

	memberships, err := s.Reader.Search(ctx, *search)
	if err != nil {
		return nil, err
	}

	groupMemberships := make(map[uuid.UUID]iam_dtos.GroupMembershipDTO)
	for _, membership := range memberships {
		groupID := membership.ResourceOwner.GroupID
		getGroupByID := common.Search{
			SearchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{
									Field:    "ID",
									Operator: common.EqualsOperator,
									Values:   []interface{}{groupID},
								},
							},
						},
					},
				},
			},
			ResultOptions: common.SearchResultOptions{
				Limit: 1,
			},
			VisibilityOptions: common.SearchVisibilityOptions{
				RequestSource:    common.GetResourceOwner(ctx),
				IntendedAudience: common.ClientApplicationAudienceIDKey,
			},
		}

		groups, err := s.GroupReader.Search(ctx, getGroupByID)
		if err != nil {
			return nil, err
		}

		if len(groups) == 0 {
			slog.WarnContext(ctx, "Group not found for Membership", "groupID", groupID, "membershipID", membership.ID)
			continue
		}

		groupMemberships[groupID] = iam_dtos.GroupMembershipDTO{
			Membership: membership,
			Group:      groups[0],
		}
	}

	return groupMemberships, nil
}
