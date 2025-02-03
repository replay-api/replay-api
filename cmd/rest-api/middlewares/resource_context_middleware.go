package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
)

type ResourceContextMiddleware struct {
	VerifyRID iam_in.VerifyRIDKeyCommand
}

func NewResourceContextMiddleware(container *container.Container) *ResourceContextMiddleware {
	var verifyRID iam_in.VerifyRIDKeyCommand
	err := container.Resolve(&verifyRID)

	if err != nil {
		slog.Error("unable to resolve VerifyRIDKeyCommand")
	}

	return &ResourceContextMiddleware{
		VerifyRID: verifyRID,
	}
}

func (m *ResourceContextMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), common.TenantIDKey, common.TeamPROTenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, common.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.AuthenticatedKey, false)

		rid := r.Header.Get(controllers.ResourceOwnerIDHeaderKey)
		if rid == "" {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		reso, aud, err := m.VerifyRID.Exec(ctx, uuid.MustParse(rid))
		if err != nil {
			slog.ErrorContext(ctx, "unable to verify rid", controllers.ResourceOwnerIDHeaderKey, rid)
			http.Error(w, "unknown", http.StatusUnauthorized)
		}

		if !reso.IsUser() {
			slog.WarnContext(ctx, "non end user resource owner", "reso", reso)
		}

		ctx = context.WithValue(ctx, common.GroupIDKey, reso.GroupID)
		ctx = context.WithValue(ctx, common.UserIDKey, reso.UserID)
		ctx = context.WithValue(ctx, common.AudienceKey, aud)
		ctx = context.WithValue(ctx, common.AuthenticatedKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
