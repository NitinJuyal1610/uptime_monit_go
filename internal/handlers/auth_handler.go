package handler

import (
	"fmt"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/internal/session"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	templates "nitinjuyal1610/uptimeMonitor/web"
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
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}
	err := s.authService.Login(email, password)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to login: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("HX-Redirect", "/")
}

func (s *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	//signup
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Email:    email,
		Password: password,
	}
	id, err := s.authService.SignUp(user)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to signup: %v", err), http.StatusBadRequest)
		return
	}

	res := utils.JSONResponse{
		Message: "Signed up successfully",
		Data: map[string]any{
			"id": id,
		},
	}
	utils.SendJSONResponse(w, http.StatusCreated, res)

}
