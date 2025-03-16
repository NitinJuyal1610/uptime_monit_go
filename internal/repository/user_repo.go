package repository

import (
	"context"
	"database/sql"
	"errors"
	"nitinjuyal1610/uptimeMonitor/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepositoryPg struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &UserRepositoryPg{db: db}
}

func (repo *UserRepositoryPg) CreateUser(ctx context.Context, user *models.User) (int, error) {
	query := `INSERT INTO users (email, password, name) VALUES ($1, $2, $3) RETURNING id`
	var userID int
	err := repo.db.QueryRow(ctx, query, user.Email, user.Password, user.Name).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (repo *UserRepositoryPg) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email,name, password FROM users WHERE email = $1`
	user := &models.User{}
	err := repo.db.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email, &user.Name, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
