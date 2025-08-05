package iam_use_cases

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_dtos "github.com/replay-api/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

type SetRIDTokenProfileCommandHandler struct {
	RIDWriter     iam_out.RIDTokenWriter
	RIDReader     iam_out.RIDTokenReader
	ProfileReader iam_out.ProfileReader
}

func (h *SetRIDTokenProfileCommandHandler) Exec(ctx context.Context, cmd iam_in.SetRIDTokenProfileCommand) (*iam_entities.RIDToken, iam_dtos.ProfilesDTO, error) {
	res, err := h.RIDReader.Search(ctx, common.NewSearchByID(ctx, cmd.RIDTokenID, common.UserAudienceIDKey))

	if err != nil {
		return nil, iam_dtos.ProfilesDTO{}, err
	}

	if len(res) == 0 {
		return nil, iam_dtos.ProfilesDTO{}, common.NewErrNotFound(common.ResourceTypeTokens, "ID", cmd.RIDTokenID.String())
	}

	token := res[0]

	profiles, err := h.ProfileReader.Search(ctx, common.NewSearchByIDs(ctx, cmd.Profiles, common.UserAudienceIDKey))

	if err != nil {
		return nil, iam_dtos.ProfilesDTO{}, err
	}

	if len(profiles) == 0 {
		return nil, iam_dtos.ProfilesDTO{}, common.NewErrNotFound(common.ResourceTypeProfile, "IDs", cmd.Profiles)
	}

	if len(profiles) != len(cmd.Profiles) {
		notFoundIDs := make([]uuid.UUID, 0)
		for _, profileID := range cmd.Profiles {
			found := false
			for _, profile := range profiles {
				if profile.ID == profileID {
					token.ActiveProfiles[profile.Type] = profileID
					found = true
					break
				}
			}
			if !found {
				notFoundIDs = append(notFoundIDs, profileID)
			}
		}

		if len(notFoundIDs) > 0 {
			return nil, iam_dtos.ProfilesDTO{}, common.NewErrNotFound(common.ResourceTypeProfile, "IDs", notFoundIDs)
		}

		return nil, iam_dtos.ProfilesDTO{}, common.NewErrConflict(common.ResourceTypeProfile, "IDs", cmd.Profiles)
	}

	token.ID = uuid.New()

	profilesDTO := iam_dtos.ProfilesDTO{}

	profilesDTO.ActiveProfiles = token.ActiveProfiles
	profilesDTO.Profiles = profiles

	_, err = h.RIDWriter.Create(ctx, &token)

	if err != nil {
		return nil, iam_dtos.ProfilesDTO{}, err
	}

	return &token, profilesDTO, nil
}
