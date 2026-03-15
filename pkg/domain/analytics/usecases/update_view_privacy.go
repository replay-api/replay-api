package analytics_usecases

import (
	"context"
	"log/slog"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
)

var _ analytics_in.UpdateViewPrivacyCommandHandler = (*UpdateViewPrivacyUseCase)(nil)

type UpdateViewPrivacyUseCase struct {
	privacyReader analytics_out.ViewPrivacyReader
	privacyWriter analytics_out.ViewPrivacyWriter
}

func NewUpdateViewPrivacyUseCase(
	privacyReader analytics_out.ViewPrivacyReader,
	privacyWriter analytics_out.ViewPrivacyWriter,
) analytics_in.UpdateViewPrivacyCommandHandler {
	return &UpdateViewPrivacyUseCase{
		privacyReader: privacyReader,
		privacyWriter: privacyWriter,
	}
}

func (uc *UpdateViewPrivacyUseCase) Exec(ctx context.Context, cmd analytics_in.UpdateViewPrivacyCommand) (*analytics_entities.ViewPrivacySettings, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	settings, err := uc.privacyReader.GetByUserID(ctx, cmd.UserID)
	if err != nil || settings == nil {
		settings = analytics_entities.NewViewPrivacySettings(cmd.UserID)
	}

	if cmd.ShowProfileViews != nil {
		settings.ShowProfileViews = *cmd.ShowProfileViews
	}
	if cmd.AllowViewerIdentification != nil {
		settings.AllowViewerIdentification = *cmd.AllowViewerIdentification
	}
	if cmd.AnonymousMode != nil {
		settings.AnonymousMode = *cmd.AnonymousMode
	}

	if err := uc.privacyWriter.Upsert(ctx, settings); err != nil {
		slog.ErrorContext(ctx, "failed to update view privacy settings", "err", err, "user_id", cmd.UserID)
		return nil, err
	}

	return settings, nil
}
