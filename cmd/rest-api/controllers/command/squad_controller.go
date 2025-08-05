package cmd_controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type SquadController struct {
	container                 container.Container
	createSquadCommandHandler squad_in.CreateSquadCommandHandler
	helper                    *controllers.ControllerHelper
}

func NewSquadController(container container.Container) *SquadController {
	var createSquadCommandHandler squad_in.CreateSquadCommandHandler
	var err = container.Resolve(&createSquadCommandHandler)
	if err != nil {
		slog.Error("Failed to resolve CreateSquadCommandHandler", "err", err)
		return nil
	}

	return &SquadController{
		container:                 container,
		createSquadCommandHandler: createSquadCommandHandler,
		helper:                    controllers.NewControllerHelper(),
	}
}

func (ctrl *SquadController) CreateSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var createSquadCommand squad_in.CreateOrUpdatedSquadCommand

		// Decode request using helper
		if err := ctrl.helper.DecodeJSONRequest(w, r, &createSquadCommand); err != nil {
			return // Error already handled by helper
		}

		// Execute command
		squad, err := ctrl.createSquadCommandHandler.Exec(r.Context(), createSquadCommand)
		if ctrl.helper.HandleError(w, r, err, "Failed to create SQUAD profile") {
			return // Error handled by helper
		}

		// Write successful response
		ctrl.helper.WriteCreated(w, r, squad)
	}
}

// CreateSquadHandlerWithContext demonstrates using context-based error handling
func (ctrl *SquadController) CreateSquadHandlerWithContext(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var createSquadCommand squad_in.CreateOrUpdatedSquadCommand

		// Decode request using helper
		if err := ctrl.helper.DecodeJSONRequest(w, r, &createSquadCommand); err != nil {
			// Store error in context for middleware to handle
			ctx := common.SetError(r.Context(), common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
			*r = *r.WithContext(ctx)
			return
		}

		squad, err := ctrl.createSquadCommandHandler.Exec(r.Context(), createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create SQUAD profile", "err", err)
			// Store error in context for middleware to handle
			ctx := common.SetError(r.Context(), err)
			*r = *r.WithContext(ctx)
			return
		}

		// Write successful response
		ctrl.helper.WriteCreated(w, r, squad)
	}
}
