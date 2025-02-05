package squad_usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type CreateSquadUseCase struct {
	SquadWriter squad_out.SquadWriter
	GroupWriter iam_out.GroupWriter
	GroupReader iam_out.GroupReader
}

func NewCreateSquadUseCase(squadWriter squad_out.SquadWriter, groupWriter iam_out.GroupWriter, groupReader iam_out.GroupReader) *CreateSquadUseCase {
	return &CreateSquadUseCase{
		SquadWriter: squadWriter,
		GroupWriter: groupWriter,
		GroupReader: groupReader,
	}
}

func (uc *CreateSquadUseCase) Exec(ctx context.Context, cmd squad_in.CreateSquadCommand) (*squad_entities.Squad, error) {
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	// TODO: validate DUPs (nickname, slug etc)

	groupSearch := iam_entities.NewGroupAccountSearchByUser(ctx)

	groups, err := uc.GroupReader.Search(ctx, groupSearch)

	if err != nil {
		return nil, err
	}

	var group *iam_entities.Group

	if len(groups) == 0 {
		group = iam_entities.NewAccountGroup(uuid.New(), common.GetResourceOwner(ctx))
		group, err = uc.GroupWriter.Create(ctx, group)

		if err != nil {
			return nil, err
		}
	} else {
		group = &groups[0]
	}

	ctx = context.WithValue(ctx, common.GroupIDKey, group.GetID())

	squad := squad_entities.NewSquad(
		group.GetID(),
		cmd.GameID,
		cmd.AvatarURI,
		cmd.Name,
		cmd.Symbol,
		cmd.Description,
		cmd.SlugURI,
		common.GetResourceOwner(ctx),
	)

	err = ValidateMembershipUUIDs(cmd.Members)

	if err != nil {
		return nil, err
	}

	squad.Membership = cmd.Members

	squad, err = uc.SquadWriter.Create(ctx, squad) // TODO: create squadHistory

	if err != nil {
		return nil, err
	}

	return squad, nil
}

func ValidateMembershipUUIDs(members map[string]squad_value_objects.SquadMembership) error {
	for key := range members {
		_, err := uuid.Parse(key)
		if err != nil {
			return fmt.Errorf("invalid UUID in membership map: %s", key)
		}
	}

	return nil
}
