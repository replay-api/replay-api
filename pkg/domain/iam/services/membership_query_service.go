package iam_query_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_dtos "github.com/replay-api/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

type MembershipQueryService struct {
	shared.BaseQueryService[iam_entities.Membership]
	GroupReader iam_out.GroupReader
}

func NewMembershipQueryService(MembershipReader shared.Searchable[iam_entities.Membership], groupReader iam_out.GroupReader) *MembershipQueryService {
	queryableFields := map[string]bool{
		"ID":            true,
		"Type":          shared.ALLOW,
		"ResourceOwner": shared.ALLOW,
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
		BaseQueryService: shared.BaseQueryService[iam_entities.Membership]{
			Reader:          MembershipReader,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        shared.UserAudienceIDKey,
		},
		GroupReader: groupReader,
	}
}

func (s *MembershipQueryService) ListMemberGroups(ctx context.Context, search *shared.Search) (map[uuid.UUID]iam_dtos.GroupMembershipDTO, error) {
	userID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, fmt.Errorf("invalid (uuid.Nil) user ID in context")
	}

	searchAggregations := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field:    "ResourceOwner.UserID",
							Operator: shared.EqualsOperator,
							Values:   []interface{}{userID},
						},
					},
				},
			},
		},
	}

	resultOptions := shared.SearchResultOptions{
		Limit: 100,
	}

	search.SearchParams = append(search.SearchParams, searchAggregations...)
	search.ResultOptions = resultOptions

	if search.VisibilityOptions.IntendedAudience == 0 {
		search.VisibilityOptions = shared.SearchVisibilityOptions{
			RequestSource:    shared.GetResourceOwner(ctx),
			IntendedAudience: shared.UserAudienceIDKey,
		}
	}

	memberships, err := s.Reader.Search(ctx, *search)
	if err != nil {
		return nil, err
	}

	groupMemberships := make(map[uuid.UUID]iam_dtos.GroupMembershipDTO)
	for _, membership := range memberships {
		groupID := membership.ResourceOwner.GroupID
		getGroupByID := shared.Search{
			SearchParams: []shared.SearchAggregation{
				{
					Params: []shared.SearchParameter{
						{
							ValueParams: []shared.SearchableValue{
								{
									Field:    "ID",
									Operator: shared.EqualsOperator,
									Values:   []interface{}{groupID},
								},
							},
						},
					},
				},
			},
			ResultOptions: shared.SearchResultOptions{
				Limit: 1,
			},
			VisibilityOptions: shared.SearchVisibilityOptions{
				RequestSource:    shared.GetResourceOwner(ctx),
				IntendedAudience: shared.ClientApplicationAudienceIDKey,
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
