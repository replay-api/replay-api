package squad_usecases

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
	squad_common "github.com/replay-api/replay-api/pkg/domain/squad"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

type CreatePlayerUseCase struct {
	PlayerWriter               squad_out.PlayerProfileWriter
	PlayerReader               squad_out.PlayerProfileReader
	GroupWriter                iam_out.GroupWriter
	GroupReader                iam_out.GroupReader
	PlayerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter
}

func NewCreatePlayerProfileUseCase(
	playerWriter squad_out.PlayerProfileWriter,
	playerReader squad_out.PlayerProfileReader,
	groupWriter iam_out.GroupWriter,
	groupReader iam_out.GroupReader,
	playerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter,
) squad_in.CreatePlayerProfileCommandHandler {
	return &CreatePlayerUseCase{
		PlayerWriter:               playerWriter,
		PlayerReader:               playerReader,
		GroupWriter:                groupWriter,
		GroupReader:                groupReader,
		PlayerProfileHistoryWriter: playerProfileHistoryWriter,
	}
}

func (uc *CreatePlayerUseCase) Exec(c context.Context, cmd squad_in.CreatePlayerProfileCommand) (*squad_entities.PlayerProfile, error) {
	isAuthenticated := c.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	// TODO: fix roles, avatar, description
	// TODO: check if token has SteamAccountID, if same name as player, connect them. (add verified bool to player, for badge)
	// TODO: design way to connect player metadata

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

	// Remove duplicate roles
	roles := squad_common.Unique(cmd.Roles)

	player := squad_entities.NewPlayerProfile(
		cmd.GameID,
		cmd.Nickname,
		cmd.AvatarURI,
		cmd.SlugURI,
		cmd.Description,
		roles,
		cmd.VisibilityType,
		common.GetResourceOwner(c),
	)
	// TODO: Verified Badge if connected wwith Steam
	// TODO: Design link wwith playerMetadata

	existingPlayers, err := uc.PlayerReader.Search(c, squad_entities.NewSearchByNickname(c, player.Nickname))

	if err != nil {
		return nil, err
	}

	if len(existingPlayers) > 0 {
		return nil, common.NewErrAlreadyExists(common.ResourceTypePlayerProfile, "Nickname", player.Nickname)
	}

	existingPlayers, err = uc.PlayerReader.Search(c, squad_entities.NewSearchBySlugURI(c, player.SlugURI))

	if err != nil {
		return nil, err
	}

	if len(existingPlayers) > 0 {
		return nil, common.NewErrAlreadyExists(common.ResourceTypePlayerProfile, "SlugURI", player.SlugURI)
	}

	player, err = uc.PlayerWriter.Create(c, player)

	if err != nil {
		return nil, err
	}

	history := squad_entities.NewPlayerProfileHistory(player.ID, squad_entities.PlayerHistoryActionCreate, common.GetResourceOwner(c))

	uc.PlayerProfileHistoryWriter.Create(c, history)

	return player, nil
}
