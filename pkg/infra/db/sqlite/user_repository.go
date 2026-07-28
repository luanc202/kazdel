package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
)

type UserRepository struct {
	dbConnection *sql.DB
}

func NewUserRepository(dbConnection *sql.DB) interfaces.UserRepository {
	return &UserRepository{dbConnection: dbConnection}
}

func (ur *UserRepository) Save(user *entity.User) error {
	query := `INSERT INTO users (id, name, username, role, email, email_verified, password_hash, created_at, updated_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		username = EXCLUDED.username,
		role = EXCLUDED.role,
		email = EXCLUDED.email,
		email_verified = EXCLUDED.email_verified,
		password_hash = EXCLUDED.password_hash,
		updated_at = EXCLUDED.updated_at`

	_, err := ur.dbConnection.ExecContext(
		context.Background(),
		query,
		user.ID.String(),
		user.Name,
		user.Username,
		user.Role,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (ur *UserRepository) FindByEmail(email string) (*entity.User, error) {
	query := `SELECT id, name, username, role, email, email_verified, password_hash, created_at, updated_at FROM users WHERE email = ? LIMIT 1`

	var user entity.User
	var idStr string

	err := ur.dbConnection.QueryRowContext(
		context.Background(),
		query,
		email,
	).Scan(
		&idStr,
		&user.Name,
		&user.Username,
		&user.Role,
		&user.Email,
		&user.EmailVerified,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
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
	query := `SELECT id, name, username, role, email, email_verified, password_hash, created_at, updated_at FROM users WHERE username = ? LIMIT 1`

	var user entity.User
	var idStr string

	err := ur.dbConnection.QueryRowContext(
		context.Background(),
		query,
		username,
	).Scan(
		&idStr,
		&user.Name,
		&user.Username,
		&user.Role,
		&user.Email,
		&user.EmailVerified,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
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
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	var exists bool

	err := ur.dbConnection.QueryRowContext(
		context.Background(),
		query,
		username,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check if username exists: %w", err)
	}

	return exists, nil
}

func (ur *UserRepository) ExistsByEmail(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`

	var exists bool

	err := ur.dbConnection.QueryRowContext(
		context.Background(),
		query,
		email,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check if email exists: %w", err)
	}

	return exists, nil
}

func (ur *UserRepository) FindById(id string) (*entity.User, error) {
	query := `SELECT id, name, username, role, email, email_verified, password_hash, created_at, updated_at FROM users WHERE id = ? LIMIT 1`

	var user entity.User
	var idStr string

	err := ur.dbConnection.QueryRowContext(
		context.Background(),
		query,
		id,
	).Scan(
		&idStr,
		&user.Name,
		&user.Username,
		&user.Role,
		&user.Email,
		&user.EmailVerified,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	parsedId, err := uniqueEntityId.ParseID(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}
	user.ID = parsedId

	return &user, nil
}
