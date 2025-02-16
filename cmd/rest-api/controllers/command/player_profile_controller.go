package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	// squad_in
	"github.com/golobby/container/v3"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type PlayerProfileController struct {
	container container.Container
}

func NewPlayerProfileController(container container.Container) *PlayerProfileController {
	return &PlayerProfileController{container: container}
}

func (ctrl *PlayerProfileController) CreatePlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var createPlayerCommand squad_in.CreatePlayerProfileCommand
		err := json.NewDecoder(r.Body).Decode(&createPlayerCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var createPlayerCommandHandler squad_in.CreatePlayerProfileCommandHandler
		err = ctrl.container.Resolve(&createPlayerCommandHandler)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve CreatePlayerProfileCommandHandler", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		player, err := createPlayerCommandHandler.Exec(r.Context(), createPlayerCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create player profile", "err", err)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if strings.Contains(err.Error(), "already exists") {
				w.WriteHeader(http.StatusConflict)
				errorJSON := map[string]string{
					"code":  "CONFLICT",
					"error": err.Error(),
				}

				err = json.NewEncoder(w).Encode(errorJSON)
				if err != nil {
					slog.ErrorContext(r.Context(), "Failed to encode response", "err", err)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		err = json.NewEncoder(w).Encode(player)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to encode response", "err", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
	}
}
