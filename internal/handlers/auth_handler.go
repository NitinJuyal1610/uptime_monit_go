package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	service "nitinjuyal1610/uptimeMonitor/internal/services"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
	templates "nitinjuyal1610/uptimeMonitor/web"
	"time"
)

type AuthHandler struct {
	authService     *service.AuthService
	templateManager *templates.TemplateManager
}

func NewAuthHandler(as *service.AuthService, tm *templates.TemplateManager) *AuthHandler {
	return &AuthHandler{as, tm}
}

func (s *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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
	token, err := s.authService.Login(email, password)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to login: %v", err), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"token":   token,
	})
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
