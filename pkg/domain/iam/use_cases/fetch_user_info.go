package iam_use_cases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

// FetchUserInfoUseCase retrieves user profile information for authenticated users.
//
// This use case provides:
//   - Retrieval of user profile by ID
//   - Validation of user authentication
//   - Resource ownership verification
//
// Use this when:
//   - Loading user profile for display
//   - Fetching user details for account management
//   - Getting current user information after authentication
type FetchUserInfoUseCase struct {
	UserReader    iam_out.UserReader
	ProfileReader iam_out.ProfileReader
}

// NewFetchUserInfoUseCase creates a new FetchUserInfoUseCase instance.
func NewFetchUserInfoUseCase(
	userReader iam_out.UserReader,
	profileReader iam_out.ProfileReader,
) *FetchUserInfoUseCase {
	return &FetchUserInfoUseCase{
		UserReader:    userReader,
		ProfileReader: profileReader,
	}
}

// UserInfo contains the combined user and profile information.
type UserInfo struct {
	User    *iam_entities.User
	Profile *iam_entities.Profile
}

// Exec retrieves user information for the specified user ID.
//
// Parameters:
//   - ctx: Context containing authentication information
//   - userID: UUID of the user to fetch
//
// Returns:
//   - *UserInfo: Combined user and profile data
//   - error: ErrUnauthorized if not authenticated, ErrNotFound if user doesn't exist
func (uc *FetchUserInfoUseCase) Exec(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	// Authentication check
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		slog.WarnContext(ctx, "unauthorized fetch user info attempt", "target_user_id", userID)
		return nil, shared.NewErrUnauthorized()
	}

	// Resource ownership check - users can only fetch their own info unless admin
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID != userID && !shared.IsAdmin(ctx) {
		slog.WarnContext(ctx, "forbidden fetch user info attempt",
			"requesting_user", resourceOwner.UserID,
			"target_user", userID,
		)
		return nil, shared.NewErrForbidden("cannot fetch another user's information")
	}

	// Fetch user via search
	userSearch := uc.newSearchByUserID(ctx, userID)
	users, err := uc.UserReader.Search(ctx, userSearch)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch user", "user_id", userID, "error", err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, shared.NewErrNotFound(shared.ResourceTypeUser, "id", userID.String())
	}

	user := &users[0]

	// Fetch profile via search
	profileSearch := uc.newSearchByUserID(ctx, userID)
	profiles, err := uc.ProfileReader.Search(ctx, profileSearch)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch user profile", "user_id", userID, "error", err)
		return nil, err
	}

	var profile *iam_entities.Profile
	if len(profiles) > 0 {
		profile = &profiles[0]
	}

	slog.InfoContext(ctx, "user info fetched successfully", "user_id", userID)

	return &UserInfo{
		User:    user,
		Profile: profile,
	}, nil
}

// newSearchByUserID creates a search query for finding by user ID.
func (uc *FetchUserInfoUseCase) newSearchByUserID(ctx context.Context, userID uuid.UUID) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field: "ID",
							Values: []interface{}{
								userID,
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
