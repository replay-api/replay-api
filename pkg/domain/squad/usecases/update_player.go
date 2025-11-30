package squad_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	common "github.com/replay-api/replay-api/pkg/domain"
	media_out "github.com/replay-api/replay-api/pkg/domain/media/ports/out"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

type UpdatePlayerUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	PlayerProfileReader      squad_in.PlayerProfileReader
	PlayerProfileWriter      squad_out.PlayerProfileWriter
	PlayerHistoryWriter      squad_out.PlayerProfileHistoryWriter
	MediaWriter              media_out.MediaWriter
}

func NewUpdatePlayerUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	playerProfileReader squad_in.PlayerProfileReader,
	playerProfileWriter squad_out.PlayerProfileWriter,
	playerHistoryWriter squad_out.PlayerProfileHistoryWriter,
	mediaWriter media_out.MediaWriter,
) *UpdatePlayerUseCase {
	return &UpdatePlayerUseCase{
		billableOperationHandler: billableOperationHandler,
		PlayerProfileReader:      playerProfileReader,
		PlayerProfileWriter:      playerProfileWriter,
		PlayerHistoryWriter:      playerHistoryWriter,
		MediaWriter:              mediaWriter,
	}
}

func (uc *UpdatePlayerUseCase) Exec(ctx context.Context, cmd squad_in.UpdatePlayerCommand) (*squad_entities.PlayerProfile, error) {
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	getByIdSearch := squad_entities.NewSearchByID(ctx, cmd.PlayerID)

	playerProfile, err := uc.PlayerProfileReader.Search(ctx, getByIdSearch)
	if err != nil {
		return nil, err
	}

	if len(playerProfile) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", cmd.PlayerID.String())
	}

	if playerProfile[0].ResourceOwner.UserID != common.GetResourceOwner(ctx).UserID {
		return nil, common.NewErrUnauthorized()
	}

	// Billing validation
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeUpdatePlayerProfile,
		UserID:      common.GetResourceOwner(ctx).UserID,
		Amount:      1,
	}
	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed for update player profile", "error", err, "player_id", cmd.PlayerID)
		return nil, err
	}

	checkSearch := squad_entities.NewNicknameAndSlugExistenceCheck(ctx, cmd.PlayerID, cmd.Nickname, cmd.SlugURI)

	exists, err := uc.PlayerProfileReader.Search(ctx, checkSearch)
	if err != nil {
		return nil, err
	}

	if len(exists) > 0 {
		for _, player := range exists {
			if player.ID != cmd.PlayerID {
				return nil, common.NewErrAlreadyExists(common.ResourceTypePlayerProfile, "Nickname or SlugURI", cmd.Nickname)
			}
		}
	}

	var avatarURI string
	if cmd.Base64Avatar != "" {
		imageName := fmt.Sprintf("%s_%s", cmd.SlugURI, uuid.New().String())
		avatarURI, err = uc.MediaWriter.Create(ctx, []byte(cmd.Base64Avatar), imageName, cmd.AvatarExtension)
		if err != nil {
			return nil, err
		}
	}

	if avatarURI != "" {
		playerProfile[0].Avatar = avatarURI
	}

	playerProfile[0].Nickname = cmd.Nickname

	playerProfile[0].SlugURI = cmd.SlugURI
	playerProfile[0].Roles = cmd.Roles
	playerProfile[0].Description = cmd.Description

	updatedPlayer, err := uc.PlayerProfileWriter.Update(ctx, &playerProfile[0])
	if err != nil {
		return nil, err
	}

	// Billing execution
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "Failed to execute billing for update player profile", "error", err, "player_id", cmd.PlayerID)
	}

	history := squad_entities.NewPlayerProfileHistory(updatedPlayer.ID, squad_entities.PlayerHistoryActionUpdate, common.GetResourceOwner(ctx))

	_, err = uc.PlayerHistoryWriter.Create(ctx, history)

	if err != nil {
		slog.WarnContext(ctx, "Failed to create player history", "err", err)
	}

	// TODO: remove old avatar/images

	return updatedPlayer, nil
}
