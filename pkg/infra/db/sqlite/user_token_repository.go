package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
)

type UserTokenRepository struct {
	dbConnection *sql.DB
}

func NewUserTokenRepository(dbConnection *sql.DB) interfaces.UserTokenRepository {
	return &UserTokenRepository{dbConnection: dbConnection}
}

func (r *UserTokenRepository) Save(token *entity.UserToken) error {
	query := `INSERT INTO user_tokens (id, user_id, token, context, expires_at, created_at) 
			VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.dbConnection.ExecContext(
		context.Background(),
		query,
		token.ID.String(),
		token.UserID.String(),
		token.Token,
		token.Context,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save user token: %w", err)
	}

	return nil
}

func (r *UserTokenRepository) FindByToken(token string) (*entity.UserToken, error) {
	query := `SELECT id, user_id, token, context, expires_at, created_at 
			FROM user_tokens WHERE token = ? LIMIT 1`

	var t entity.UserToken
	var idStr, userIdStr string

	err := r.dbConnection.QueryRowContext(
		context.Background(),
		query,
		token,
	).Scan(&idStr, &userIdStr, &t.Token, &t.Context, &t.ExpiresAt, &t.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to find token: %w", err)
	}

	id, err := uniqueEntityId.ParseID(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id: %w", err)
	}
	t.ID = id

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user_id: %w", err)
	}
	t.UserID = userId

	return &t, nil
}

func (r *UserTokenRepository) DeleteByToken(token string) error {
	query := `DELETE FROM user_tokens WHERE token = ?`

	_, err := r.dbConnection.ExecContext(
		context.Background(),
		query,
		token,
	)

	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

func (r *UserTokenRepository) DeleteByUserIdAndContext(userId string, tokenContext entity.TokenContext) error {
	ctx := context.Background()
	query := `DELETE FROM user_tokens WHERE user_id = ? AND context = ?`

	_, err := r.dbConnection.ExecContext(
		ctx,
		query,
		userId,
		tokenContext,
	)

	if err != nil {
		return fmt.Errorf("failed to delete tokens by user id and context: %w", err)
	}

	return nil
}
