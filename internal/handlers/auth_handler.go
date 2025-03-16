package handler

import (
	"fmt"
	"log"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/internal/session"
	templates "nitinjuyal1610/uptimeMonitor/web"
	"strings"
)

type AuthHandler struct {
	authService     *service.AuthService
	templateManager *templates.TemplateManager
	sessionManager  *session.SessionManager
}

func NewAuthHandler(as *service.AuthService, tm *templates.TemplateManager, sm *session.SessionManager) *AuthHandler {
	return &AuthHandler{as, tm, sm}
}

func (s *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<div class="text-red-400 text-sm mb-2">Email is required</div>`)
		return
	}
	if password == "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<div class="text-red-400 text-sm mb-2">Password is required</div>`)
		return
	}
	userId, err := s.authService.Login(r.Context(), email, password)

	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<div class="text-red-400 text-sm mb-2">Invalid email or password</div>`)
		return
	}

	// Create session
	if err := s.sessionManager.Create(w, r, userId); err != nil {
		log.Printf("Session creation failed for user %d: %v", userId, err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<div class="text-red-400 text-sm mb-2">Unable to create session</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}


func (s *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	//signup
	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")

	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Email:    email,
		Password: password,
		Name:     name,
	}
	id, err := s.authService.SignUp(r.Context(), user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to signup: %v", err), http.StatusBadRequest)
		return
	}

	//create session redirect to /
	if err := s.sessionManager.Create(w, r, id); err != nil {
		log.Printf("Session creation failed for user %d: %v", id, err)
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusCreated)
}

func (s *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	//-------//

	if err := s.sessionManager.Destroy(w, r); err != nil {
		log.Printf("Failed to Logout %v", err)
		http.Error(w, "Unable to Logout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}
