package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
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
		ctx := context.WithValue(r.Context(), shared.TenantIDKey, replay_common.TeamPROTenantID)
		ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())
		ctx = context.WithValue(ctx, shared.AuthenticatedKey, false)

		rid := r.Header.Get(controllers.ResourceOwnerIDHeaderKey)
		if rid == "" {
			// Check if this is a replay upload request - allow guest uploads
			isReplayUpload := strings.Contains(r.URL.Path, "/games/") && strings.Contains(r.URL.Path, "/replays") && r.Method == "POST" && !strings.Contains(r.URL.Path, "/replays/")
			slog.InfoContext(ctx, "checking replay upload", "path", r.URL.Path, "method", r.Method, "isReplayUpload", isReplayUpload)
			if isReplayUpload {
				slog.InfoContext(ctx, "allowing guest upload for replay endpoint", "path", r.URL.Path)
				// Set up guest resource owner context
				guestUserID := uuid.New()
				guestGroupID := uuid.New()
				ctx = context.WithValue(ctx, shared.GroupIDKey, guestGroupID)
				ctx = context.WithValue(ctx, shared.UserIDKey, guestUserID)
				ctx = context.WithValue(ctx, shared.AudienceKey, shared.UserAudienceIDKey)
				ctx = context.WithValue(ctx, shared.AuthenticatedKey, true) // Allow guest uploads
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			
			slog.WarnContext(ctx, "missing resource owner id", "ResourceOwnerIDHeaderKey", controllers.ResourceOwnerIDHeaderKey)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ridUUID, parseErr := uuid.Parse(rid)
		if parseErr != nil {
			slog.ErrorContext(ctx, "invalid resource owner id format", "rid", rid, "err", parseErr)
			http.Error(w, "invalid resource owner id", http.StatusBadRequest)
			return
		}

		reso, aud, err := m.VerifyRID.Exec(ctx, ridUUID)
		if err != nil {
			slog.ErrorContext(ctx, "unable to verify rid", controllers.ResourceOwnerIDHeaderKey, rid, "err", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify we got a valid resource owner (not zero value)
		if reso.UserID == uuid.Nil && reso.GroupID == uuid.Nil {
			slog.ErrorContext(ctx, "empty resource owner returned from verification", "rid", rid)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		slog.InfoContext(ctx, "resource owner verified", "user_id", reso.UserID, "group_id", reso.GroupID, "aud", aud)

		if !reso.IsUser() {
			slog.WarnContext(ctx, "non end user resource owner", "reso", reso)
		}

		// Set ALL resource owner fields in context including TenantID and ClientID
		ctx = context.WithValue(ctx, shared.TenantIDKey, reso.TenantID)
		ctx = context.WithValue(ctx, shared.ClientIDKey, reso.ClientID)
		ctx = context.WithValue(ctx, shared.GroupIDKey, reso.GroupID)
		ctx = context.WithValue(ctx, shared.UserIDKey, reso.UserID)
		ctx = context.WithValue(ctx, shared.AudienceKey, aud)
		ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
