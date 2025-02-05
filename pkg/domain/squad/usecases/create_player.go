package squad_usecases

import (
	"context"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
)

type CreatePlayerUseCase struct {
	PlayerWriter squad_out.PlayerProfileWriter
	GroupWriter  iam_out.GroupWriter
	GroupReader  iam_out.GroupReader
}

func (uc *CreatePlayerUseCase) Exec(c context.Context, cmd squad_in.CreatePlayerCommand) (*squad_entities.PlayerProfile, error) {
	isAuthenticated := c.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	groupSearch := iam_entities.NewGroupAccountSearchByUser(c)

	groups, err := uc.GroupReader.Search(c, groupSearch)

	if err != nil {
		return nil, err
	}

	var group *iam_entities.Group

	if len(groups) == 0 {
		group = iam_entities.NewAccountGroup(uuid.New(), common.GetResourceOwner(c))
		group, err = uc.GroupWriter.Create(c, group)

		if err != nil {
			return nil, err
		}
	} else {
		group = &groups[0]
	}

	c = context.WithValue(c, common.GroupIDKey, group.GetID())

	player := squad_entities.NewPlayerProfile(
		cmd.GameID,
		cmd.Nickname,
		cmd.AvatarURI,
		"",
		cmd.VisibilityType,
		common.GetResourceOwner(c),
	)
	// TODO: Verified Badge if connected wwith Steam
	// TODO: Design link wwith playerMetadata

	player, err = uc.PlayerWriter.Create(c, player)

	if err != nil {
		return nil, err
	}

	return player, nil
}
