package squad_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
	media_out "github.com/replay-api/replay-api/pkg/domain/media/ports/out"
	squad_common "github.com/replay-api/replay-api/pkg/domain/squad"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// CreateSquadUseCase handles the creation of competitive squads (teams) on the platform.
//
// This is a core use case for team-based competitive gaming, enabling players to:
//   - Form organized teams with designated roles
//   - Create team identity (name, logo, slug URL)
//   - Establish team membership hierarchy (owners, managers, players)
//
// Flow:
//  1. Authentication verification - user must be authenticated
//  2. Input validation - UUID format, slug URL format, uniqueness checks
//  3. Duplicate prevention - checks for existing squads with same slug/name
//  4. Group account management - creates IAM group for squad resource ownership
//  5. Billing validation - verifies subscription allows squad creation
//  6. Membership processing - validates and creates member relationships
//  7. Avatar handling - processes and stores squad logo if provided
//  8. Squad persistence - creates squad entity with history record
//
// Security:
//   - Requires authenticated context (shared.AuthenticatedKey)
//   - Creates isolated IAM group for squad resource ownership
//   - Validates all member player profiles exist and are accessible
//
// Features:
//   - Slug URL validation (alphanumeric, hyphens, underscores only)
//   - Automatic duplicate detection for names and slugs
//   - Avatar upload with media storage integration
//   - Squad history tracking for audit trail
//
// Dependencies:
//   - BillableOperationCommandHandler: Subscription limit validation
//   - SquadWriter/Reader: Squad persistence
//   - GroupWriter/Reader: IAM group management for squad ownership
//   - SquadHistoryWriter: Audit trail
//   - PlayerProfileReader: Member validation
//   - MediaWriter: Avatar storage
type CreateSquadUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	SquadWriter              squad_out.SquadWriter
	SquadReader              squad_out.SquadReader
	GroupWriter              iam_out.GroupWriter
	GroupReader              iam_out.GroupReader
	SquadHistoryWriter       squad_out.SquadHistoryWriter
	PlayerProfileReader      squad_in.PlayerProfileReader
	MediaWriter              media_out.MediaWriter
}

func NewCreateSquadUseCase(squadWriter squad_out.SquadWriter, squadHistoryWriter squad_out.SquadHistoryWriter, squadReader squad_out.SquadReader, groupWriter iam_out.GroupWriter, groupReader iam_out.GroupReader, playerProfileReader squad_in.PlayerProfileReader, mediaWriter media_out.MediaWriter, billableOperationHandler billing_in.BillableOperationCommandHandler) *CreateSquadUseCase {
	return &CreateSquadUseCase{
		billableOperationHandler: billableOperationHandler,
		SquadWriter:              squadWriter,
		SquadHistoryWriter:       squadHistoryWriter,
		SquadReader:              squadReader,
		GroupWriter:              groupWriter,
		GroupReader:              groupReader,
		PlayerProfileReader:      playerProfileReader,
		MediaWriter:              mediaWriter,
	}
}

func ValidateSlugURL(slugURI string) error {
	if len(slugURI) < 3 {
		return fmt.Errorf("slugURI must be at least 3 characters long")
	}

	for _, char := range slugURI {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-' || char == '_' || char == '/') {
			return fmt.Errorf("slugURI contains invalid characters")
		}
	}

	return nil
}

// Exec creates a new squad based on the provided command and context.
// It performs several validation checks and operations, including:
// 1. Authentication: Ensures the user is authenticated.
// 2. Membership UUID Validation: Validates the UUIDs of the squad members.
// 3. Slug URL Validation: Validates the slug URL of the squad.
// 4. Duplicate Squad Check: Checks if a squad with the same SlugURI or Name already exists.
// 5. Group Account Handling: Creates or retrieves a group account for the user.
// 6. Billing Operation: Validates and executes a billing operation for creating the squad.
// 7. Membership Handling: Processes and validates squad memberships.
// 8. Avatar Handling: Processes the squad's avatar if provided.
// 9. Squad Creation: Creates the squad and its history record.
//
// Parameters:
// - ctx: The context for the operation, which includes authentication and other metadata.
// - cmd: The command containing the details for creating or updating the squad.
//
// Returns:
// - A pointer to the created squad entity.
// - An error if any validation or operation fails.
func (uc *CreateSquadUseCase) Exec(ctx context.Context, cmd squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error) {
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	err := ValidateMembershipUUIDs(cmd.Members)

	if err != nil {
		return nil, err
	}

	err = ValidateSlugURL(cmd.SlugURI)

	if err != nil {
		return nil, err
	}

	existingSquads, err := uc.SquadReader.Search(ctx, squad_entities.NewSearchBySlugURI(ctx, cmd.SlugURI))

	if err != nil {
		return nil, err
	}

	if len(existingSquads) > 0 {
		return nil, shared.NewErrAlreadyExists(replay_common.ResourceTypeSquad, "SlugURI", cmd.SlugURI)
	}

	existingSquads, err = uc.SquadReader.Search(ctx, squad_entities.NewSearchByName(ctx, cmd.Name))

	if err != nil {
		return nil, err
	}

	if len(existingSquads) > 0 {
		return nil, shared.NewErrAlreadyExists(replay_common.ResourceTypeSquad, "Name", cmd.Name)
	}

	groupSearch := iam_entities.NewGroupAccountSearchByUser(ctx)

	groups, err := uc.GroupReader.Search(ctx, groupSearch)

	if err != nil {
		return nil, err
	}

	rxn := shared.GetResourceOwner(ctx)

	var group *iam_entities.Group

	if len(groups) == 0 {
		group = iam_entities.NewAccountGroup(uuid.New(), rxn)
		group, err = uc.GroupWriter.Create(ctx, group)

		if err != nil {
			return nil, err
		}
	} else {
		group = &groups[0]
	}

	ctx = context.WithValue(ctx, shared.GroupIDKey, group.GetID())

	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeCreateSquadProfile,
		UserID:      rxn.UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"SlugURI": cmd.SlugURI,
			"Name":    cmd.Name,
			"GameID":  cmd.GameID,
		},
	}

	_ = uc.billableOperationHandler.Validate(ctx, billingCmd)

	memberships := make([]squad_value_objects.SquadMembership, 0)
	membershipMap := make(map[uuid.UUID]interface{})

	for k, v := range cmd.Members {
		playerProfileID := uuid.MustParse(k)
		if membershipMap[playerProfileID] != nil {
			continue
		}
		players, err := uc.PlayerProfileReader.Search(ctx, squad_entities.NewSearchByID(ctx, playerProfileID))

		if err != nil {
			return nil, err
		}

		if len(players) == 0 {
			return nil, shared.NewErrNotFound(replay_common.ResourceTypePlayerProfile, "ID", playerProfileID.String())
		}
		slog.InfoContext(ctx, "roles", "roles", v.Roles)
		userID := players[0].ResourceOwner.UserID
		membershipMap[playerProfileID] = struct{}{}
		memberships = append(memberships, *squad_value_objects.NewSquadMembership(
			userID,
			playerProfileID,
			v.Type,
			squad_common.Unique(v.Roles),
			v.Status,
			v.Type,
		))
	}

	var avatarURI string
	if cmd.Base64Logo != "" {
		imageName := fmt.Sprintf("%s_%s", cmd.SlugURI, uuid.New().String())
		avatarURI, err = uc.MediaWriter.Create(ctx, []byte(cmd.Base64Logo), imageName, cmd.LogoExtension)
		if err != nil {
			return nil, err
		}
	}

	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)

	if err != nil {
		slog.ErrorContext(ctx, "create squad failed: unable to execute billing command", "err", err, "rxn", rxn)
		return nil, err
	}

	squad := squad_entities.NewSquad(
		group.GetID(),
		cmd.GameID,
		avatarURI,
		cmd.Name,
		cmd.Symbol,
		cmd.Description,
		cmd.SlugURI,
		memberships,
		rxn,
	)

	squad, err = uc.SquadWriter.Create(ctx, squad)

	if err != nil {
		return nil, err
	}

	squadHistory := squad_entities.NewSquadHistory(
		squad.GetID(),
		rxn.UserID,
		squad_entities.SquadCreated,
		rxn,
	)

	_, _ = uc.SquadHistoryWriter.Create(ctx, squadHistory)

	return squad, nil
}

func ValidateMembershipUUIDs(members map[string]squad_in.CreateSquadMembershipInput) error {
	for key := range members {
		_, err := uuid.Parse(key)
		if err != nil {
			return fmt.Errorf("invalid UUID in membership map: %s", key)
		}
	}

	return nil
}
