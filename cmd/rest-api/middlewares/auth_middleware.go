// implement a go mux middleware that receives the Bearer Token of a steam account, and validate it against: https://api.steampowered.com/ISteamUserOAuth/GetTokenDetails/v1/?access_token=token in and add the steamID to context

package middlewares

import (
	"context"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "no-auth", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authorizationHeader, "Bearer ")
		if len(bearerToken) != 2 {
			http.Error(w, "no-auth", http.StatusUnauthorized)
			return
		}

		// TODO: remover bypass
		// TODO review!!
		// steamToken := bearerToken[1]

		// steamID, err := getSteamID(steamToken)
		// if err != nil {
		// 	http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		// 	return
		// }

		// ctx := context.WithValue(r.Context(), "steamID", steamID)
		next.ServeHTTP(w, r.WithContext(context.Background()))
	})
}
