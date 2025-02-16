package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
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
			slog.ErrorContext(r.Context(), "Failed to create SQUAD profile", "err", err)
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
			} else if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				errorJSON := map[string]string{
					"code":  "NOT_FOUND",
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

		err = json.NewEncoder(w).Encode(squad)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to encode response", "err", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
	}
}
