package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
)

type ShortenedUrlRepository struct {
	dbConnection *sql.DB
}

func NewShortenedUrlRepository(dbConn *sql.DB) interfaces.ShortenedUrlRepository {
	return &ShortenedUrlRepository{
		dbConnection: dbConn,
	}
}

func (pr *ShortenedUrlRepository) Save(shortenedUrl *entity.ShortenedUrl) error {
	query := `INSERT INTO shortened_urls
	(short_slug, long_url, description, password_hash, expires_at, user_id)
	VALUES
	(?, ?, ?, ?, ?, ?)`

	_, err := pr.dbConnection.ExecContext(
		context.Background(),
		query,
		shortenedUrl.ShortSlug,
		shortenedUrl.LongUrl,
		shortenedUrl.Description,
		shortenedUrl.PasswordHash,
		shortenedUrl.ExpiresAt,
		shortenedUrl.UserId.String(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return entity.ErrCustomSlugAlreadyExists
		}
		slog.Error("failed to insert shortened url into database", "error", err)
		return err
	}

	return nil
}

func (pr *ShortenedUrlRepository) FindBySlug(slug string) (*entity.ShortenedUrl, error) {
	query := `SELECT id, short_slug, long_url, description, password_hash, expires_at, user_id FROM shortened_urls WHERE short_slug = ? LIMIT 1`

	var shortenedUrl entity.ShortenedUrl

	err := pr.dbConnection.QueryRowContext(
		context.Background(),
		query,
		slug,
	).Scan(
		&shortenedUrl.ID,
		&shortenedUrl.ShortSlug,
		&shortenedUrl.LongUrl,
		&shortenedUrl.Description,
		&shortenedUrl.PasswordHash,
		&shortenedUrl.ExpiresAt,
		&shortenedUrl.UserId,
	)

	if err != nil {
		slog.Error("failed to find shortened url by slug", "error", err)
		return nil, err
	}

	return &shortenedUrl, nil
}

func (pr *ShortenedUrlRepository) FindByUserIdPaginated(userId uniqueEntityId.ID, search string, page, limit int) ([]*entity.ShortenedUrl, int, error) {
	offset := (page - 1) * limit

	baseQuery := `
		FROM shortened_urls s
		LEFT JOIN url_visits v ON s.id = v.url_id
		WHERE s.user_id = ?
	`
	args := []any{userId.String()}

	if search != "" {
		baseQuery += ` AND (s.short_slug LIKE ? OR s.long_url LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	countSql := `SELECT COUNT(DISTINCT s.id) ` + baseQuery
	var total int
	err := pr.dbConnection.QueryRowContext(context.Background(), countSql, args...).Scan(&total)
	if err != nil {
		slog.Error("failed to count shortened urls", "error", err)
		return nil, 0, err
	}

	query := `
		SELECT s.short_slug, s.long_url, s.description, s.password_hash, s.expires_at, s.user_id, s.created_at, COUNT(v.id) as views
		` + baseQuery + `
		GROUP BY s.id
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`
	queryArgs := append(args, limit, offset)

	rows, err := pr.dbConnection.QueryContext(context.Background(), query, queryArgs...)
	if err != nil {
		slog.Error("failed to find shortened urls by user id", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var shortenedUrls []*entity.ShortenedUrl

	for rows.Next() {
		var shortenedUrl entity.ShortenedUrl
		err := rows.Scan(
			&shortenedUrl.ShortSlug,
			&shortenedUrl.LongUrl,
			&shortenedUrl.Description,
			&shortenedUrl.PasswordHash,
			&shortenedUrl.ExpiresAt,
			&shortenedUrl.UserId,
			&shortenedUrl.CreatedAt,
			&shortenedUrl.Views,
		)
		if err != nil {
			slog.Error("failed to scan shortened url", "error", err)
			return nil, 0, err
		}
		shortenedUrls = append(shortenedUrls, &shortenedUrl)
	}

	return shortenedUrls, total, nil
}

func (pr *ShortenedUrlRepository) DeleteExpired(now time.Time) error {
	query := `DELETE FROM shortened_urls WHERE expires_at < ?`
	slog.Info("Deleting expired shortened urls", "now", now)
	_, err := pr.dbConnection.ExecContext(
		context.Background(),
		query,
		now,
	)

	if err != nil {
		slog.Error("failed to delete expired shortened urls from database", "error", err)
		return err
	}

	return nil
}

func (pr *ShortenedUrlRepository) Delete(id uint64, userId uniqueEntityId.ID) error {
	query := `DELETE FROM shortened_urls WHERE id = ? AND user_id = ?`
	slog.Info("Deleting shortened url", "id", id, "userId", userId)
	result, err := pr.dbConnection.ExecContext(
		context.Background(),
		query,
		id,
		userId.String(),
	)

	if err != nil {
		slog.Error("failed to delete shortened url from database", "error", err)
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("shortened url not found")
	}

	return nil
}

func (pr *ShortenedUrlRepository) Update(shortenedUrl *entity.ShortenedUrl) error {
	query := `UPDATE shortened_urls 
	SET long_url = ?, 
	    description = ?, 
	    password_hash = ?, 
	    expires_at = ?,
	    updated_at = CURRENT_TIMESTAMP
	WHERE id = ? AND user_id = ?`

	result, err := pr.dbConnection.ExecContext(
		context.Background(),
		query,
		shortenedUrl.LongUrl,
		shortenedUrl.Description,
		shortenedUrl.PasswordHash,
		shortenedUrl.ExpiresAt,
		shortenedUrl.ID,
		shortenedUrl.UserId.String(),
	)

	if err != nil {
		slog.Error("failed to update shortened url in database", "error", err)
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("shortened url not found or not owned by user")
	}

	return nil
}
