package squad_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
	media_out "github.com/psavelis/team-pro/replay-api/pkg/domain/media/ports/out"
	squad_common "github.com/psavelis/team-pro/replay-api/pkg/domain/squad"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type CreateSquadUseCase struct {
	SquadWriter         squad_out.SquadWriter
	SquadReader         squad_out.SquadReader
	GroupWriter         iam_out.GroupWriter
	GroupReader         iam_out.GroupReader
	SquadHistoryWriter  squad_out.SquadHistoryWriter
	PlayerProfileReader squad_in.PlayerProfileReader
	MediaWriter         media_out.MediaWriter
}

func NewCreateSquadUseCase(squadWriter squad_out.SquadWriter, squadHistoryWriter squad_out.SquadHistoryWriter, squadReader squad_out.SquadReader, groupWriter iam_out.GroupWriter, groupReader iam_out.GroupReader, playerProfileReader squad_in.PlayerProfileReader, mediaWriter media_out.MediaWriter) *CreateSquadUseCase {
	return &CreateSquadUseCase{
		SquadWriter:         squadWriter,
		SquadHistoryWriter:  squadHistoryWriter,
		SquadReader:         squadReader,
		GroupWriter:         groupWriter,
		GroupReader:         groupReader,
		PlayerProfileReader: playerProfileReader,
		MediaWriter:         mediaWriter,
	}
}

func ValidateSlugURL(slugURI string) error {
	if len(slugURI) < 3 {
		return fmt.Errorf("slugURI must be at least 3 characters long")
	}

	for _, char := range slugURI {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-' || char == '_') {
			return fmt.Errorf("slugURI contains invalid characters")
		}
	}

	return nil
}

func (uc *CreateSquadUseCase) Exec(ctx context.Context, cmd squad_in.CreateSquadCommand) (*squad_entities.Squad, error) {
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	// TODO: check plan for QTY of squads

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
		return nil, common.NewErrAlreadyExists(common.ResourceTypeSquad, "SlugURI", cmd.SlugURI)
	}

	existingSquads, err = uc.SquadReader.Search(ctx, squad_entities.NewSearchByName(ctx, cmd.Name))

	if err != nil {
		return nil, err
	}

	if len(existingSquads) > 0 {
		return nil, common.NewErrAlreadyExists(common.ResourceTypeSquad, "Name", cmd.Name)
	}

	groupSearch := iam_entities.NewGroupAccountSearchByUser(ctx)

	groups, err := uc.GroupReader.Search(ctx, groupSearch)

	if err != nil {
		return nil, err
	}

	rxn := common.GetResourceOwner(ctx)

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

	ctx = context.WithValue(ctx, common.GroupIDKey, group.GetID())

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
			return nil, common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", playerProfileID.String())
		}
		slog.InfoContext(ctx, "roles", "roles", v.Roles)
		userID := players[0].ResourceOwner.UserID
		membershipMap[playerProfileID] = struct{}{}
		memberships = append(memberships, *squad_value_objects.NewSquadMembership(
			userID,
			playerProfileID,
			v.Type,
			squad_common.Unique(v.Roles),
			squad_value_objects.SquadMembershipStatusActive,
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

	uc.SquadHistoryWriter.Create(ctx, squadHistory)

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
