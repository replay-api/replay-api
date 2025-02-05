package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
)

type SquadController struct {
	container container.Container
}

func NewSquadController(container container.Container) *SquadController {
	return &SquadController{container: container}
}

func (ctrl *SquadController) CreateSquadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var createSquadCommand squad_in.CreateSquadCommand
		err := json.NewDecoder(r.Body).Decode(&createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var createSquadCommandHandler squad_in.CreateSquadCommandHandler
		err = ctrl.container.Resolve(&createSquadCommandHandler)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve CreateSquadCommandHandler", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		squad, err := createSquadCommandHandler.Exec(r.Context(), createSquadCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create squad", "err", err)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		err = json.NewEncoder(w).Encode(squad)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to encode response", "err", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
	}
}
