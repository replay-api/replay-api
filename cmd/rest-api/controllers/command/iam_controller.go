package cmd_controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

type IAMController struct {
	container container.Container
	helper    *controllers.ControllerHelper
}

func NewIAMController(container container.Container) *IAMController {
	return &IAMController{
		container: container,
		helper:    controllers.NewControllerHelper(),
	}
}

func (ctlr *IAMController) SetRIDTokenProfile(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var setRIDTokenProfileCommand iam_in.SetRIDTokenProfileCommand

		// Decode request using helper
		if err := ctlr.helper.DecodeJSONRequest(w, r, &setRIDTokenProfileCommand); err != nil {
			return // Error already handled by helper
		}

		var setRIDTokenProfileCommandHandler iam_in.SetRIDTokenProfileCommandHandler
		err := ctlr.container.Resolve(&setRIDTokenProfileCommandHandler)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve SetRIDTokenProfileCommandHandler", "err", err)
			ctlr.helper.HandleError(w, r, common.NewAPIError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "failed to resolve command handler"), "dependency resolution failed")
			return
		}

		token, profilesDto, err := setRIDTokenProfileCommandHandler.Exec(r.Context(), setRIDTokenProfileCommand)
		if err != nil {
			// Handle specific error cases
			if err.Error() == "Unauthorized" {
				ctlr.helper.HandleError(w, r, common.NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized access"), "unauthorized access")
			} else {
				ctlr.helper.HandleError(w, r, err, "Failed to set RID token profile")
			}
			return
		}

		// Set additional headers for resource owner information
		w.Header().Set(controllers.ResourceOwnerIDHeaderKey, token.GetID().String())
		w.Header().Set(controllers.ResourceOwnerAudTypeHeaderKey, string(token.IntendedAudience))

		// Write successful response with profile data
		ctlr.helper.WriteCreated(w, r, profilesDto)
	}
}
