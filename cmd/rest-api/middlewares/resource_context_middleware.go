package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
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

		rid := r.Header.Get("x-resource-owner-id")
		if rid == "" {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		reso, err := m.VerifyRID.Exec(ctx, uuid.MustParse(rid))
		if err != nil {
			slog.ErrorContext(ctx, "unable to verify rid", "x-resource-owner-id", rid)
			http.Error(w, "unknown", http.StatusUnauthorized)
		}

		if !reso.IsUser() {
			slog.WarnContext(ctx, "non end user resource owner", "reso", reso)
		}

		ctx = context.WithValue(ctx, common.GroupIDKey, reso.GroupID)
		ctx = context.WithValue(ctx, common.UserIDKey, reso.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
