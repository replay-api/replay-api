package iam_use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

// OnboardOpenIDUserUseCase handles user onboarding from OpenID providers (Steam, Google, Email).
//
// This is the central onboarding use case that orchestrates:
//   1. Profile lookup - checks if user already exists by source key
//   2. User creation - creates new User entity if not found
//   3. Group creation - creates system group for user's resources
//   4. Membership creation - establishes owner membership in group
//   5. Profile creation - links external identity to internal user
//   6. RID Token creation - issues authentication token
//
// Flow:
//   - Existing User: Returns existing profile + new RID token
//   - New User: Creates User → Group → Membership → Profile → RID Token
//
// Dependencies:
//   - UserReader/Writer: User entity persistence
//   - ProfileReader/Writer: Profile entity persistence (links to external identity)
//   - GroupWriter: System group creation for resource ownership
//   - MembershipWriter: User-to-group membership
//   - CreateRIDToken: Token issuance for authentication
type OnboardOpenIDUserUseCase struct {
	UserReader       iam_out.UserReader
	UserWriter       iam_out.UserWriter
	ProfileReader    iam_out.ProfileReader
	ProfileWriter    iam_out.ProfileWriter
	GroupWriter      iam_out.GroupWriter
	MembershipWriter iam_out.MembershipWriter
	CreateRIDToken   iam_in.CreateRIDTokenCommand
}

func NewOnboardOpenIDUserUseCase(userReader iam_out.UserReader, userWriter iam_out.UserWriter, profileReader iam_out.ProfileReader, profileWriter iam_out.ProfileWriter, groupWriter iam_out.GroupWriter, membershipWriter iam_out.MembershipWriter, createRIDToken iam_in.CreateRIDTokenCommand) *OnboardOpenIDUserUseCase {
	return &OnboardOpenIDUserUseCase{
		UserReader:       userReader,
		UserWriter:       userWriter,
		ProfileReader:    profileReader,
		ProfileWriter:    profileWriter,
		GroupWriter:      groupWriter,
		MembershipWriter: membershipWriter,
		CreateRIDToken:   createRIDToken,
	}
}

// Exec onboards a user from an OpenID provider.
//
// Parameters:
//   - ctx: Context with ResourceOwner (UserID, GroupID must be pre-populated for new users)
//   - cmd: OnboardOpenIDUserCommand containing:
//     - Name: User's display name
//     - Source: RIDSourceKey (Steam/Google/Email)
//     - Key: External identifier (Steam ID, email address)
//     - ProfileDetails: Provider-specific profile data
//
// Returns:
//   - *Profile: Created or existing user profile
//   - *RIDToken: Fresh authentication token
//   - error: Validation or persistence errors
//
// Errors:
//   - "invalid resource owner: no user id" - UserID not in context for new user
//   - "invalid resource owner: no group id" - GroupID not in context for new user
func (uc *OnboardOpenIDUserUseCase) Exec(ctx context.Context, cmd iam_in.OnboardOpenIDUserCommand) (*iam_entities.Profile, *iam_entities.RIDToken, error) {
	profileSourceKeySearch := uc.newSearchByProfileSourceKey(ctx, cmd.Source, cmd.Key)

	slog.InfoContext(ctx, fmt.Sprintf("profileSourceKeySearch: %v", profileSourceKeySearch))

	profiles, err := uc.ProfileReader.Search(ctx, profileSourceKeySearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting user profile", "err",
			err)
		return nil, nil, err
	}

	slog.InfoContext(ctx, fmt.Sprintf("profileSourceKeySearch: %v, profiles %v, len: %d", profileSourceKeySearch, profiles, len(profiles)))

	if len(profiles) > 0 {
		// TODO: check if profile, user and group is active, try reuse token
		slog.InfoContext(ctx, fmt.Sprintf("attempt to reuse user profile and create RID Token: %v", profiles[0]))
		ridToken, err := uc.CreateRIDToken.Exec(ctx, profiles[0].GetResourceOwner(ctx), cmd.Source, iam_entities.DefaultTokenAudience) // ??? Or group Audience?

		if err != nil {
			slog.ErrorContext(ctx, "error creating rid token", "err",
				err)
			return nil, nil, err
		}

		return &profiles[0], ridToken, nil
	}

	rxn := shared.GetResourceOwner(ctx)

	if rxn.UserID == uuid.Nil {
		return nil, nil, fmt.Errorf("invalid resource owner: no user id")
	}

	if rxn.GroupID == uuid.Nil {
		return nil, nil, fmt.Errorf("invalid resource owner: no group id")
	}

	user := iam_entities.NewUser(rxn.UserID, cmd.Name, rxn)

	slog.InfoContext(ctx, fmt.Sprintf("attempt to create user: %v", user))

	user, err = uc.UserWriter.Create(ctx, user)
	if err != nil {
		slog.ErrorContext(ctx, "error creating user", "err", err)
		return nil, nil, err
	}

	rxn.UserID = user.ID

	group := iam_entities.NewGroup(rxn.GroupID, iam_entities.DefaultUserGroupName, iam_entities.GroupTypeSystem, rxn)

	rxn.GroupID = group.ID

	group, err = uc.GroupWriter.Create(ctx, group)

	if err != nil {
		slog.ErrorContext(ctx, "error creating group", "err",
			err)
		return nil, nil, err
	}

	membership := iam_entities.NewMembership(iam_entities.MembershipTypeOwner, iam_entities.MembershipStatusActive, rxn)

	_, err = uc.MembershipWriter.Create(ctx, membership)

	if err != nil {
		slog.ErrorContext(ctx, "error creating membership", "err",
			err)
		return nil, nil, err
	}

	profile := iam_entities.NewProfile(user.ID, group.ID, cmd.Source, cmd.Key, cmd.ProfileDetails, rxn)

	profile, err = uc.ProfileWriter.Create(ctx, profile)

	if err != nil {
		slog.ErrorContext(ctx, "error creating user profile", "err",
			err)
		return nil, nil, err
	}

	ridToken, err := uc.CreateRIDToken.Exec(ctx, profile.GetResourceOwner(ctx), cmd.Source, iam_entities.DefaultTokenAudience)

	if err != nil {
		slog.ErrorContext(ctx, "error creating rid token", "err",
			err)
		return nil, nil, err
	}

	if ridToken == nil {
		return nil, nil, fmt.Errorf("failed to create rid token: token is nil")
	}

	return profile, ridToken, nil
}

func (uc *OnboardOpenIDUserUseCase) newSearchByProfileSourceKey(ctx context.Context, source iam_entities.RIDSourceKey, key string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					// Operator: shared.EqualsOperator,
					ValueParams: []shared.SearchableValue{
						{
							Field: "RIDSource",
							Values: []interface{}{
								source,
							},
						},
						{
							Field: "SourceKey",
							Values: []interface{}{
								key,
							},
						},
					},
				},
			},
		},
	}

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}
