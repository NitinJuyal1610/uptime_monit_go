package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type SessionManager struct {
	store *sessions.CookieStore
}

func NewSessionManager(secretKey string) *SessionManager {
	return &SessionManager{
		store: sessions.NewCookieStore([]byte(secretKey)),
	}
}

func (sm *SessionManager) Create(w http.ResponseWriter, r *http.Request, userId int) error {
	session, err := sm.store.Get(r, "auth-session")
	if err != nil {
		return err
	}
	session.Values["authenticated"] = true
	session.Values["userId"] = userId
	return session.Save(r, w)
}

func (sm *SessionManager) Destroy(w http.ResponseWriter, r *http.Request) error {
	session, err := sm.store.Get(r, "auth-session")
	if err != nil {
		return err
	}
	session.Values["authenticated"] = false
	delete(session.Values, "userId")
	return session.Save(r, w)
}

func (sm *SessionManager) GetUserID(r *http.Request) (int, bool) {
	session, _ := sm.store.Get(r, "auth-session")

	userID, ok := session.Values["userId"].(int)
	return userID, ok
}
