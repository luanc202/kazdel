package db

import (
	"context"
	"fmt"
	"time"
	"url-shortener/m/entity"
	"url-shortener/m/internal/uniqueEntityId"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSessionRepository struct {
	Db *pgxpool.Pool
}

func NewPostgresSessionRepository(db *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{
		Db: db,
	}
}

func (r *PostgresSessionRepository) Create(session *entity.Session) error {
	query := `INSERT INTO sessions (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.Db.Exec(context.Background(), query, session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepository) FindByToken(token string) (*entity.Session, error) {
	query := `SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = $1`
	row := r.Db.QueryRow(context.Background(), query, token)

	var session entity.Session
	var idStr, userIdStr string
	err := row.Scan(&idStr, &userIdStr, &session.Token, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	session.ID, _ = uniqueEntityId.ParseID(idStr)
	session.UserID, _ = uniqueEntityId.ParseID(userIdStr)

	return &session, nil
}

func (r *PostgresSessionRepository) DeleteByToken(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.Db.Exec(context.Background(), query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepository) DeleteExpired() error {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	_, err := r.Db.Exec(context.Background(), query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}
