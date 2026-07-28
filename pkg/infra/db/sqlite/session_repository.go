package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
	"time"
)

type SqliteSessionRepository struct {
	Db *sql.DB
}

func NewSessionRepository(db *sql.DB) interfaces.SessionRepository {
	return &SqliteSessionRepository{
		Db: db,
	}
}

func (r *SqliteSessionRepository) Create(session *entity.Session) error {
	query := `INSERT INTO sessions (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.Db.ExecContext(context.Background(), query, session.ID.String(), session.UserID.String(), session.Token, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *SqliteSessionRepository) FindByToken(token string) (*entity.Session, error) {
	query := `SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = ?`
	row := r.Db.QueryRowContext(context.Background(), query, token)

	var session entity.Session
	var idStr, userIdStr string
	err := row.Scan(&idStr, &userIdStr, &session.Token, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	session.ID, _ = uniqueEntityId.ParseID(idStr)
	session.UserID, _ = uniqueEntityId.ParseID(userIdStr)

	return &session, nil
}

func (r *SqliteSessionRepository) DeleteByToken(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := r.Db.ExecContext(context.Background(), query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *SqliteSessionRepository) DeleteExpired() error {
	query := `DELETE FROM sessions WHERE expires_at < ?`
	_, err := r.Db.ExecContext(context.Background(), query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}
