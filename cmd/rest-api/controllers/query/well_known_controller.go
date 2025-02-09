package query_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

// WellKnownController handles the Open Configuration requests
type WellKnownController struct {
	wellKnownReader iam_in.WellKnownReader
}

// NewWellKnownController returns a new WellKnownController instance
func NewWellKnownController(c *container.Container) *WellKnownController {
	var wellKnownReader iam_in.WellKnownReader
	err := c.Resolve(&wellKnownReader)

	if err != nil {
		slog.Error("Cannot resolve iam_entities.WellKnownReader for WellKnownController", "err", err)
		panic(err)
	}

	return &WellKnownController{wellKnownReader: wellKnownReader}
}

// HandleWellKnown handles the Open Configuration request
func (c *WellKnownController) HandleOpenConfiguration(w http.ResponseWriter, r *http.Request) {

	ncontext := context.WithValue(r.Context(), common.TenantIDKey, common.TeamPROTenantID)

	wellKnown, err := c.wellKnownReader.GetOpenConfiguration(ncontext)

	if err != nil {
		slog.ErrorContext(r.Context(), "Error reading well-known configuration", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wellKnown)
}

// HandleWellKnownJwks handles the Open Configuration JWKS request
func (c *WellKnownController) HandleOpenConfigurationJwks(w http.ResponseWriter, r *http.Request) {
	jwks, err := c.wellKnownReader.GetOpenConfigurationJwks(r.Context())

	if err != nil {
		slog.ErrorContext(r.Context(), "Error reading well-known configuration", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}
