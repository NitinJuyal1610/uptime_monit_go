package service

import (
	"fmt"
	"nitinjuyal1610/uptimeMonitor/internal/models"
	"nitinjuyal1610/uptimeMonitor/internal/repository"
	"nitinjuyal1610/uptimeMonitor/pkg/utils"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo}
}

func (as *AuthService) Login(email, password string) (int, error) {
	existingUser, err := as.userRepo.GetUserByEmail(email)

	if err != nil {
		fmt.Println(err)
		return 0, fmt.Errorf("user with this email does not exist")
	}

	//compare password
	isCorrect := utils.CheckPasswordHash(password, existingUser.Password)
	if !isCorrect {
		return 0, fmt.Errorf("incorrect password")
	}
	//generate token
	return existingUser.Id, nil
}

func (as *AuthService) SignUp(user *models.User) (int, error) {
	_, err := as.userRepo.GetUserByEmail(user.Email)
	if err == nil {
		return 0, fmt.Errorf("user already exist with this email")
	}
	//hash
	user.Password, _ = utils.HashPassword(user.Password)
	return as.userRepo.CreateUser(user)
}
