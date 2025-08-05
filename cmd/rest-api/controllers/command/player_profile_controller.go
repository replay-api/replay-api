package cmd_controllers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type PlayerProfileController struct {
	container container.Container
	helper    *controllers.ControllerHelper
}

func NewPlayerProfileController(container container.Container) *PlayerProfileController {
	return &PlayerProfileController{
		container: container,
		helper:    controllers.NewControllerHelper(),
	}
}

func (ctrl *PlayerProfileController) CreatePlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var createPlayerCommand squad_in.CreatePlayerProfileCommand

		// Decode request using helper
		if err := ctrl.helper.DecodeJSONRequest(w, r, &createPlayerCommand); err != nil {
			return // Error already handled by helper
		}

		var createPlayerCommandHandler squad_in.CreatePlayerProfileCommandHandler
		err := ctrl.container.Resolve(&createPlayerCommandHandler)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve CreatePlayerProfileCommandHandler", "err", err)
			ctrl.helper.HandleError(w, r, common.NewAPIError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "failed to resolve command handler"), "dependency resolution failed")
			return
		}

		player, err := createPlayerCommandHandler.Exec(r.Context(), createPlayerCommand)
		if err != nil {
			// Handle specific error cases
			if err.Error() == "Unauthorized" {
				ctrl.helper.HandleError(w, r, common.NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized access"), "unauthorized access")
			} else if strings.Contains(err.Error(), "already exists") {
				// Handle conflict case with specific error format for backward compatibility
				conflictErr := common.NewAPIError(http.StatusConflict, "CONFLICT", err.Error())
				ctrl.helper.HandleError(w, r, conflictErr, "player profile already exists")
			} else {
				ctrl.helper.HandleError(w, r, err, "Failed to create player profile")
			}
			return
		}

		// Write successful response
		ctrl.helper.WriteCreated(w, r, player)
	}
}
