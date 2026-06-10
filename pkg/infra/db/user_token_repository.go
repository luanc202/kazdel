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

type UserTokenRepository struct {
	dbConnection *pgxpool.Pool
}

func NewUserTokenRepository(dbConnection *pgxpool.Pool) interfaces.UserTokenRepository {
	return &UserTokenRepository{dbConnection: dbConnection}
}

func (r *UserTokenRepository) Save(token *entity.UserToken) error {
	sql := `INSERT INTO user_tokens (id, user_id, token, context, expires_at, created_at) 
			VALUES (@id, @user_id, @token, @context, @expires_at, @created_at)`

	_, err := r.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"id":         token.ID,
			"user_id":    token.UserID,
			"token":      token.Token,
			"context":    token.Context,
			"expires_at": token.ExpiresAt,
			"created_at": token.CreatedAt,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to save user token: %w", err)
	}

	return nil
}

func (r *UserTokenRepository) FindByToken(token string) (*entity.UserToken, error) {
	sql := `SELECT id, user_id, token, context, expires_at, created_at 
			FROM user_tokens WHERE token = @token LIMIT 1`

	var t entity.UserToken
	var idStr, userIdStr string

	err := r.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{"token": token},
	).Scan(&idStr, &userIdStr, &t.Token, &t.Context, &t.ExpiresAt, &t.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
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
	sql := `DELETE FROM user_tokens WHERE token = @token`

	_, err := r.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{"token": token},
	)

	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

func (r *UserTokenRepository) DeleteByUserIdAndContext(userId string, tokenContext entity.TokenContext) error {
	ctx := context.Background()
	sql := `DELETE FROM user_tokens WHERE user_id = @user_id AND context = @context`

	_, err := r.dbConnection.Exec(
		ctx,
		sql,
		pgx.NamedArgs{
			"user_id": userId,
			"context": tokenContext,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to delete tokens by user id and context: %w", err)
	}

	return nil
}
