package repository

import (
	"database/sql"
	"errors"
	"nitinjuyal1610/uptimeMonitor/internal/models"
)

type UserRepository interface {
	CreateUser(user *models.User) (int, error)
	GetUserByEmail(email string) (*models.User, error)
}

type UserRepositoryPg struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryPg{db: db}
}

func (repo *UserRepositoryPg) CreateUser(user *models.User) (int, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`
	var userID int
	err := repo.db.QueryRow(query, user.Email, user.Password).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (repo *UserRepositoryPg) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password FROM users WHERE email = $1`
	user := &models.User{}
	err := repo.db.QueryRow(query, email).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
