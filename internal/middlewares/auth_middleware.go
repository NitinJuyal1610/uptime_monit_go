package middleware

import (
	"context"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/session"
)

type AuthMiddleware struct {
	sessionManager *session.SessionManager
}

type userToken string

func NewAuthMiddleware(sessionManager *session.SessionManager) *AuthMiddleware {
	return &AuthMiddleware{sessionManager: sessionManager}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, authenticated := m.sessionManager.GetUserID(r)
		if !authenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var userToken userToken = "userId"
		ctx := context.WithValue(r.Context(), userToken, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
