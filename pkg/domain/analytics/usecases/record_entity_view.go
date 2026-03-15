package analytics_usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

const viewDeduplicationWindow = 5 * time.Minute

// ensure interface compliance
var _ analytics_in.RecordViewCommandHandler = (*RecordEntityViewUseCase)(nil)

type RecordEntityViewUseCase struct {
	viewWriter     analytics_out.EntityViewWriter
	viewReader     analytics_out.EntityViewReader
	eventPublisher analytics_out.ViewEventPublisher
	privacyReader  analytics_out.ViewPrivacyReader
}

func NewRecordEntityViewUseCase(
	viewWriter analytics_out.EntityViewWriter,
	viewReader analytics_out.EntityViewReader,
	eventPublisher analytics_out.ViewEventPublisher,
	privacyReader analytics_out.ViewPrivacyReader,
) analytics_in.RecordViewCommandHandler {
	return &RecordEntityViewUseCase{
		viewWriter:     viewWriter,
		viewReader:     viewReader,
		eventPublisher: eventPublisher,
		privacyReader:  privacyReader,
	}
}

func (uc *RecordEntityViewUseCase) Exec(ctx context.Context, cmd analytics_in.RecordViewCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	viewerType := analytics_entities.ViewerTypeAnonymous
	var effectiveViewerID *uuid.UUID

	if cmd.ViewerID != nil {
		viewerType = analytics_entities.ViewerTypeAuthenticated

		// Check if viewer has anonymous mode enabled
		privacySettings, err := uc.privacyReader.GetByUserID(ctx, *cmd.ViewerID)
		if err == nil && privacySettings != nil && privacySettings.AnonymousMode {
			effectiveViewerID = nil
			viewerType = analytics_entities.ViewerTypeAnonymous
		} else {
			effectiveViewerID = cmd.ViewerID

			// Deduplicate: skip if same viewer viewed this entity within the window
			since := time.Now().Add(-viewDeduplicationWindow)
			existing, err := uc.viewReader.GetLastViewByViewer(ctx, cmd.EntityID, *cmd.ViewerID, since)
			if err == nil && existing != nil {
				slog.DebugContext(ctx, "view deduplicated",
					"entity_id", cmd.EntityID,
					"viewer_id", cmd.ViewerID,
					"window", viewDeduplicationWindow)
				return nil
			}
		}
	}

	referrerType := cmd.ReferrerType
	if referrerType == "" {
		referrerType = analytics_entities.ReferrerTypeDirect
	}

	deviceType := cmd.DeviceType
	if deviceType == "" {
		deviceType = analytics_entities.DeviceTypeUnknown
	}

	rxn := shared.GetResourceOwner(ctx)

	view := analytics_entities.NewEntityView(
		cmd.EntityID,
		cmd.EntityType,
		effectiveViewerID,
		viewerType,
		cmd.SessionID,
		referrerType,
		cmd.GeoRegion,
		deviceType,
		rxn,
	)

	_, err := uc.viewWriter.Create(ctx, view)
	if err != nil {
		slog.ErrorContext(ctx, "failed to persist entity view", "err", err, "entity_id", cmd.EntityID)
		return err
	}

	// Publish event non-blocking
	go func() {
		if pubErr := uc.eventPublisher.PublishEntityViewed(context.Background(), view); pubErr != nil {
			slog.Error("failed to publish entity viewed event", "err", pubErr, "entity_id", view.EntityID)
		}
	}()

	slog.InfoContext(ctx, "entity view recorded",
		"entity_id", cmd.EntityID,
		"entity_type", cmd.EntityType,
		"viewer_type", viewerType)

	return nil
}
