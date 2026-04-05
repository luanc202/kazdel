package db

import (
	"context"
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	dbConnection *pgxpool.Pool
}

func NewUserRepository(dbConnection *pgxpool.Pool) interfaces.UserRepository {
	return &UserRepository{dbConnection: dbConnection}
}

func (ur *UserRepository) Save(user *entity.User) error {
	sql := `INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES (@id, @name, @email, @password_hash, @created_at, @updated_at)`

	_, err := ur.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"id":            user.ID,
			"name":          user.Name,
			"email":         user.Email,
			"password_hash": user.PasswordHash,
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (ur *UserRepository) FindByEmail(email string) (*entity.User, error) {
	sql := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = @email LIMIT 1`

	var user entity.User
	var idStr string

	err := ur.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"email": email,
		},
	).Scan(
		&idStr,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	id, err := uniqueEntityId.ParseID(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}
	user.ID = id

	return &user, nil
}

func (ur *UserRepository) FindByUsername(username string) (*entity.User, error) {
	sql := `SELECT id, name, username, email, password_hash, created_at, updated_at FROM users WHERE username = @username LIMIT 1`

	var user entity.User
	var idStr string

	err := ur.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"username": username,
		},
	).Scan(
		&idStr,
		&user.Name,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	id, err := uniqueEntityId.ParseID(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}
	user.ID = id

	return &user, nil
}

func (ur *UserRepository) ExistsByUsername(username string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE username = @username)`

	var exists bool

	err := ur.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"username": username,
		},
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check if username exists: %w", err)
	}

	return exists, nil
}

func (ur *UserRepository) ExistsByEmail(email string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE email = @email)`

	var exists bool

	err := ur.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"email": email,
		},
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check if email exists: %w", err)
	}

	return exists, nil
}
