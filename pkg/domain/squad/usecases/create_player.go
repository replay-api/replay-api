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
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type CreatePlayerUseCase struct {
	billableOperationHandler   billing_in.BillableOperationCommandHandler
	PlayerWriter               squad_out.PlayerProfileWriter
	PlayerReader               squad_out.PlayerProfileReader
	GroupWriter                iam_out.GroupWriter
	GroupReader                iam_out.GroupReader
	PlayerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter
	MediaWriter                media_out.MediaWriter
}

func NewCreatePlayerProfileUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	playerWriter squad_out.PlayerProfileWriter,
	playerReader squad_out.PlayerProfileReader,
	groupWriter iam_out.GroupWriter,
	groupReader iam_out.GroupReader,
	playerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter,
	mediaWriter media_out.MediaWriter,
) squad_in.CreatePlayerProfileCommandHandler {
	return &CreatePlayerUseCase{
		billableOperationHandler:   billableOperationHandler,
		PlayerWriter:               playerWriter,
		PlayerReader:               playerReader,
		GroupWriter:                groupWriter,
		GroupReader:                groupReader,
		PlayerProfileHistoryWriter: playerProfileHistoryWriter,
		MediaWriter:                mediaWriter,
	}
}

func (uc *CreatePlayerUseCase) Exec(c context.Context, cmd squad_in.CreatePlayerProfileCommand) (*squad_entities.PlayerProfile, error) {
	isAuthenticated := c.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, shared.NewErrUnauthorized()
	}

	groupSearch := iam_entities.NewGroupAccountSearchByUser(c)

	groups, err := uc.GroupReader.Search(c, groupSearch)

	if err != nil {
		return nil, err
	}

	var group *iam_entities.Group

	if len(groups) == 0 {
		group = iam_entities.NewAccountGroup(uuid.New(), shared.GetResourceOwner(c))
		group, err = uc.GroupWriter.Create(c, group)

		if err != nil {
			return nil, err
		}
	} else {
		group = &groups[0]
	}

	c = context.WithValue(c, shared.GroupIDKey, group.GetID())

	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeCreatePlayerProfile,
		UserID:      shared.GetResourceOwner(c).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"Nickname": cmd.Nickname,
			"SlugURI":  cmd.SlugURI,
			"GameID":   cmd.GameID,
		},
	}

	err = uc.billableOperationHandler.Validate(c, billingCmd)
	if err != nil {
		return nil, err
	}

	roles := squad_common.Unique(cmd.Roles)

	var avatarURI string
	if cmd.Base64Avatar != "" {
		imageName := fmt.Sprintf("%s_%s", cmd.SlugURI, uuid.New().String())
		avatarURI, err = uc.MediaWriter.Create(c, []byte(cmd.Base64Avatar), imageName, cmd.AvatarExtension)
		if err != nil {
			return nil, err
		}
	}

	existingPlayers, err := uc.PlayerReader.Search(c, squad_entities.NewSearchByNickname(c, cmd.Nickname))

	if err != nil {
		return nil, err
	}

	if len(existingPlayers) > 0 {
		return nil, shared.NewErrAlreadyExists(replay_common.ResourceTypePlayerProfile, "Nickname", cmd.Nickname)
	}

	existingPlayers, err = uc.PlayerReader.Search(c, squad_entities.NewSearchBySlugURI(c, cmd.SlugURI))

	if err != nil {
		return nil, err
	}

	if len(existingPlayers) > 0 {
		return nil, shared.NewErrAlreadyExists(replay_common.ResourceTypePlayerProfile, "SlugURI", cmd.SlugURI)
	}

	player := squad_entities.NewPlayerProfile(
		cmd.GameID,
		cmd.Nickname,
		avatarURI,
		cmd.SlugURI,
		cmd.Description,
		roles,
		cmd.VisibilityType,
		shared.GetResourceOwner(c),
	)

	// TODO: Verified Badge if connected with Steam (set networkIDs)
	// TODO: Add PlayerMetadataIDs (due to multiple networks) // queue to reconcile replays etc

	player, err = uc.PlayerWriter.Create(c, player)

	if err != nil {
		slog.ErrorContext(c, "create player profile failed", "err", err)
		return nil, fmt.Errorf("unable to create player profile")
	}

	if _, _, err := uc.billableOperationHandler.Exec(c, billingCmd); err != nil {
		slog.ErrorContext(c, "create player profile failed: unable to account usage", "err", err)
		return nil, err
	}

	history := squad_entities.NewPlayerProfileHistory(player.ID, squad_entities.PlayerHistoryActionCreate, shared.GetResourceOwner(c))

	_, _ = uc.PlayerProfileHistoryWriter.Create(c, history)

	slog.InfoContext(c, "player profile created", "player_id", player.ID, "nickname", player.Nickname, "slug_uri", player.SlugURI)

	return player, nil
}
