package server

import (
	"context"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	"strings"
)

type userToken string

const UserTokenKey userToken = "userToken"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		//verify token
		tokenData, err := utils.VerifyToken(token)

		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserTokenKey, tokenData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
