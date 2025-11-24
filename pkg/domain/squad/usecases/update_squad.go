package squad_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
	media_out "github.com/replay-api/replay-api/pkg/domain/media/ports/out"
	squad_common "github.com/replay-api/replay-api/pkg/domain/squad"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
)

type UpdateSquadUseCase struct {
	SquadWriter         squad_out.SquadWriter
	SquadReader         squad_out.SquadReader
	GroupWriter         iam_out.GroupWriter
	GroupReader         iam_out.GroupReader
	SquadHistoryWriter  squad_out.SquadHistoryWriter
	PlayerProfileReader squad_in.PlayerProfileReader
	MediaWriter         media_out.MediaWriter
}

func NewUpdateSquadUseCase(
	squadWriter squad_out.SquadWriter,
	squadHistoryWriter squad_out.SquadHistoryWriter,
	squadReader squad_out.SquadReader,
	groupWriter iam_out.GroupWriter,
	groupReader iam_out.GroupReader,
	playerProfileReader squad_in.PlayerProfileReader,
	mediaWriter media_out.MediaWriter,
) *UpdateSquadUseCase {
	return &UpdateSquadUseCase{
		SquadWriter:         squadWriter,
		SquadHistoryWriter:  squadHistoryWriter,
		SquadReader:         squadReader,
		GroupWriter:         groupWriter,
		GroupReader:         groupReader,
		PlayerProfileReader: playerProfileReader,
		MediaWriter:         mediaWriter,
	}
}

func (uc *UpdateSquadUseCase) Exec(ctx context.Context, squadID uuid.UUID, cmd squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error) {
	if !uc.isAuthenticated(ctx) {
		return nil, common.NewErrUnauthorized()
	}

	if err := uc.validateCommand(cmd); err != nil {
		return nil, err
	}

	existingSquad, err := uc.getExistingSquad(ctx, squadID, cmd)
	if err != nil {
		return nil, err
	}

	if !uc.isUserAuthorized(ctx, existingSquad) {
		return nil, fmt.Errorf("forbidden: user is not authorized to update this squad")
	}

	memberships, err := uc.ProcessMemberships(ctx, existingSquad, cmd.Members)
	if err != nil {
		return nil, err
	}

	avatarURI, err := uc.processAvatar(ctx, cmd)
	if err != nil {
		return nil, err
	}

	updatedSquad, err := uc.updateSquad(ctx, existingSquad, cmd, memberships, avatarURI)
	if err != nil {
		return nil, err
	}

	// TODO: remove old avatar/images

	return updatedSquad, nil
}

func (uc *UpdateSquadUseCase) isAuthenticated(ctx context.Context) bool {
	isAuthenticated := ctx.Value(common.AuthenticatedKey).(bool)
	return isAuthenticated
}

func (uc *UpdateSquadUseCase) validateCommand(cmd squad_in.CreateOrUpdatedSquadCommand) error {
	if err := ValidateMembershipUUIDs(cmd.Members); err != nil {
		return err
	}
	if err := ValidateSlugURL(cmd.SlugURI); err != nil {
		return err
	}
	return nil
}

func (uc *UpdateSquadUseCase) getExistingSquad(ctx context.Context, squadID uuid.UUID, cmd squad_in.CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error) {
	if err := uc.checkSquadExists(ctx, squad_entities.NewSearchBySlugURI(ctx, cmd.SlugURI), "SlugURI", cmd.SlugURI); err != nil {
		return nil, err
	}
	if err := uc.checkSquadExists(ctx, squad_entities.NewSearchByName(ctx, cmd.Name), "Name", cmd.Name); err != nil {
		return nil, err
	}
	existingSquads, err := uc.SquadReader.Search(ctx, common.NewSearchByID(ctx, squadID, common.ClientApplicationAudienceIDKey))
	if err != nil {
		return nil, err
	}
	if len(existingSquads) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "ID", squadID.String())
	}
	return &existingSquads[0], nil
}

func (uc *UpdateSquadUseCase) checkSquadExists(ctx context.Context, search common.Search, field, value string) error {
	existingSquads, err := uc.SquadReader.Search(ctx, search)
	if err != nil {
		return err
	}
	if len(existingSquads) > 0 {
		return common.NewErrAlreadyExists(common.ResourceTypeSquad, field, value)
	}
	return nil
}

func latestStatusTime(statusMap map[time.Time]squad_value_objects.SquadMembershipStatus) time.Time {
	var latest time.Time
	for t := range statusMap {
		if t.After(latest) {
			latest = t
		}
	}
	return latest
}

func (uc *UpdateSquadUseCase) ProcessMemberships(ctx context.Context, squad *squad_entities.Squad, members map[string]squad_in.CreateSquadMembershipInput) (map[uuid.UUID]*squad_value_objects.SquadMembership, error) {
	memberships := make(map[uuid.UUID]*squad_value_objects.SquadMembership)
	membershipMap := make(map[uuid.UUID]interface{})
	for k, v := range members {
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
		slog.InfoContext(ctx, "ProcessMemberships:roles", "roles", v.Roles)
		userID := players[0].ResourceOwner.UserID
		existingMembership, exists := memberships[playerProfileID]
		if exists {
			action := uc.GetPromotedOrDemotedStatus(ctx, squad, existingMembership)
			if action != squad_entities.SquadMemberAdded {
				uc.SquadHistoryWriter.Create(ctx, squad_entities.NewSquadHistory(squad.ID, userID, action, common.GetResourceOwner(ctx)))
			}
			existingMembership.Type = v.Type
			existingMembership.Roles = squad_common.Unique(v.Roles)
			if len(existingMembership.Status) == 0 || existingMembership.Status[latestStatusTime(existingMembership.Status)] != v.Status {
				existingMembership.Status[time.Now()] = v.Status
				uc.SquadHistoryWriter.Create(ctx, squad_entities.NewSquadHistory(squad.ID, userID, squad_entities.SquadMemberAdded, common.GetResourceOwner(ctx)))
			}
		} else {
			memberships[playerProfileID] = squad_value_objects.NewSquadMembership(userID, playerProfileID, v.Type, squad_common.Unique(v.Roles), v.Status, v.Type)
			uc.SquadHistoryWriter.Create(ctx, squad_entities.NewSquadHistory(squad.ID, userID, squad_entities.SquadMemberAdded, common.GetResourceOwner(ctx)))
		}
	}
	return memberships, nil
}

func (uc *UpdateSquadUseCase) processAvatar(ctx context.Context, cmd squad_in.CreateOrUpdatedSquadCommand) (string, error) {
	if cmd.Base64Logo == "" {
		return "", nil
	}
	imageName := fmt.Sprintf("%s_%s", cmd.SlugURI, uuid.New().String())
	return uc.MediaWriter.Create(ctx, []byte(cmd.Base64Logo), imageName, cmd.LogoExtension)
}

func (uc *UpdateSquadUseCase) updateSquad(ctx context.Context, squad *squad_entities.Squad, cmd squad_in.CreateOrUpdatedSquadCommand, memberships map[uuid.UUID]*squad_value_objects.SquadMembership, avatarURI string) (*squad_entities.Squad, error) {
	squad.Name = cmd.Name
	squad.Symbol = cmd.Symbol
	squad.Description = cmd.Description
	squad.GameID = cmd.GameID
	squad.LogoURI = avatarURI
	squad.Membership = uc.convertMembershipMapToSlice(memberships)
	updatedSquad, err := uc.SquadWriter.Update(ctx, squad)
	if err != nil {
		return nil, err
	}
	uc.SquadHistoryWriter.Create(ctx, squad_entities.NewSquadHistory(updatedSquad.GetID(), common.GetResourceOwner(ctx).UserID, squad_entities.SquadUpdated, common.GetResourceOwner(ctx)))
	return updatedSquad, nil
}

func (uc *UpdateSquadUseCase) convertMembershipMapToSlice(memberships map[uuid.UUID]*squad_value_objects.SquadMembership) []squad_value_objects.SquadMembership {
	membershipSlice := make([]squad_value_objects.SquadMembership, 0, len(memberships))
	for _, membership := range memberships {
		membershipSlice = append(membershipSlice, *membership)
	}
	return membershipSlice
}

func (uc *UpdateSquadUseCase) isUserAuthorized(ctx context.Context, squad *squad_entities.Squad) bool {
	userID := ctx.Value(common.UserIDKey).(uuid.UUID)
	if userID == squad.ResourceOwner.UserID {
		return true
	}
	for _, v := range squad.Membership {
		if v.UserID == userID && (v.Type == squad_value_objects.SquadMembershipTypeOwner || v.Type == squad_value_objects.SquadMembershipTypeAdmin) {
			return true
		}
	}
	fmt.Printf("User %v is not authorized to update squad %v\n", userID, squad.ID)
	return false
}

func (uc *UpdateSquadUseCase) GetPromotedOrDemotedStatus(ctx context.Context, squad *squad_entities.Squad, membership *squad_value_objects.SquadMembership) squad_entities.SquadHistoryAction {
	for _, v := range squad.Membership {
		if v.UserID == membership.UserID {
			if v.Type != membership.Type {
				if v.Type == squad_value_objects.SquadMembershipTypeOwner || v.Type == squad_value_objects.SquadMembershipTypeAdmin {
					return squad_entities.SquadMemberDemoted
				}
				if membership.Type == squad_value_objects.SquadMembershipTypeOwner || membership.Type == squad_value_objects.SquadMembershipTypeAdmin {
					return squad_entities.SquadMemberPromoted
				}
			}
		}
	}
	return squad_entities.SquadMemberAdded
}
