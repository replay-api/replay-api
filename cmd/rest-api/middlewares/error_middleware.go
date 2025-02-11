package middlewares

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)

		hdr := w.Header()

		if len(hdr.Get("Content-Type")) == 0 {
			w.Header().Set("Content-Type", "application/json")
		}

		err := r.Context().Err()

		if err != nil {
			var statusCode int
			var errorCode string

			switch {
			case rr.statusCode == http.StatusUnauthorized || err.Error() == "Unauthorized":
				statusCode = http.StatusUnauthorized
				errorCode = "UNAUTHORIZED"
			case rr.statusCode == http.StatusForbidden || err.Error() == "Forbidden":
				statusCode = http.StatusForbidden
				errorCode = "FORBIDDEN"
			case rr.statusCode == http.StatusNotFound || err.Error() == "Not Found":
				statusCode = http.StatusNotFound
				errorCode = "NOT_FOUND"
			case rr.statusCode == http.StatusConflict || err.Error() == "Conflict" || strings.Contains(err.Error(), "already exists"):
				statusCode = http.StatusConflict
				errorCode = "CONFLICT"
			case rr.statusCode == http.StatusBadRequest || err.Error() == "Bad Request":
				statusCode = http.StatusBadRequest
				errorCode = "BAD_REQUEST"
			case rr.statusCode == http.StatusInternalServerError || err.Error() == "Internal Server Error":
				statusCode = http.StatusInternalServerError
				errorCode = "INTERNAL_SERVER_ERROR"
			default:
				statusCode = http.StatusInternalServerError
				errorCode = "UNKNOWN_ERROR"
			}

			if errorCode != "" {
				rr.WriteHeader(statusCode)
			}
			response := map[string]string{
				"code":  errorCode,
				"error": err.Error(),
			}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)

			slog.ErrorContext(r.Context(), "ErrorMiddleware", "err", err, "Status", rr.statusCode)
			return
		} else {
			slog.InfoContext(r.Context(), "ErrorMiddleware", "Status", rr.statusCode)
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}
