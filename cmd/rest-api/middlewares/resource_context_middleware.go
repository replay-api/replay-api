package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
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
		slog.InfoContext(r.Context(), "resource context middleware", "path", r.URL.Path, "method", r.Method, "rid", r.Header.Get(controllers.ResourceOwnerIDHeaderKey))
		ctx := context.WithValue(r.Context(), common.TenantIDKey, common.TeamPROTenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, common.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.AuthenticatedKey, false)

		rid := r.Header.Get(controllers.ResourceOwnerIDHeaderKey)
		if rid == "" {
			slog.WarnContext(ctx, "missing resource owner id", "ResourceOwnerIDHeaderKey", controllers.ResourceOwnerIDHeaderKey)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		reso, aud, err := m.VerifyRID.Exec(ctx, uuid.MustParse(rid))
		if err != nil {
			slog.ErrorContext(ctx, "unable to verify rid", controllers.ResourceOwnerIDHeaderKey, rid)
			// Store error in context instead of writing directly
			ctx = common.SetError(ctx, common.ErrUnauthorized)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		slog.InfoContext(ctx, "resource owner verified", "reso", reso, "aud", aud)

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
