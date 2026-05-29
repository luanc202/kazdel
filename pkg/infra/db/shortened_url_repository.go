package db

import (
	"context"
	"errors"
	"time"

	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShortenedUrlRepository struct {
	dbConnection *pgxpool.Pool
}

func NewShortenedUrlRepository(dbConn *pgxpool.Pool) interfaces.ShortenedUrlRepository {
	return &ShortenedUrlRepository{
		dbConnection: dbConn,
	}
}

func (pr *ShortenedUrlRepository) Save(shortenedUrl *entity.ShortenedUrl) error {

	sql := `INSERT INTO shortened_urls
	(short_slug, long_url, description, password_hash, expires_at, user_id)
	VALUES
	(@shortSlug, @longUrl, @description, @passwordHash, @expiresAt, @userId)`

	_, err := pr.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"shortSlug":    shortenedUrl.ShortSlug,
			"longUrl":      shortenedUrl.LongUrl,
			"description":  shortenedUrl.Description,
			"passwordHash": shortenedUrl.PasswordHash,
			"expiresAt":    shortenedUrl.ExpiresAt,
			"userId":       shortenedUrl.UserId,
		})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrCustomSlugAlreadyExists
		}
		slog.Error("failed to insert product into database", "error", err)
		return err
	}

	return nil
}

func (pr *ShortenedUrlRepository) FindBySlug(slug string) (*entity.ShortenedUrl, error) {

	sql := `SELECT id, short_slug, long_url, description, password_hash, expires_at, user_id FROM shortened_urls WHERE short_slug = @slug
	LIMIT 1`

	var shortenedUrl entity.ShortenedUrl

	err := pr.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"slug": slug,
		}).Scan(
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
		WHERE s.user_id = @userId
	`
	args := pgx.NamedArgs{
		"userId": userId,
		"limit":  limit,
		"offset": offset,
	}

	if search != "" {
		baseQuery += ` AND (s.short_slug ILIKE @search OR s.long_url ILIKE @search)`
		args["search"] = "%" + search + "%"
	}

	countSql := `SELECT COUNT(DISTINCT s.id) ` + baseQuery
	var total int
	err := pr.dbConnection.QueryRow(context.Background(), countSql, args).Scan(&total)
	if err != nil {
		slog.Error("failed to count shortened urls", "error", err)
		return nil, 0, err
	}

	sql := `
		SELECT s.short_slug, s.long_url, s.description, s.password_hash, s.expires_at, s.user_id, s.created_at, COUNT(v.id) as views
		` + baseQuery + `
		GROUP BY s.id
		ORDER BY s.created_at DESC
		LIMIT @limit OFFSET @offset
	`

	rows, err := pr.dbConnection.Query(context.Background(), sql, args)
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
	sql := `DELETE FROM shortened_urls WHERE expires_at < @now`
	slog.Info("Deleting expired shortened urls", "now", now)
	_, err := pr.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"now": now,
		})

	if err != nil {
		slog.Error("failed to delete expired shortened urls from database", "error", err)
		return err
	}

	return nil
}

func (pr *ShortenedUrlRepository) Delete(id uint64, userId uniqueEntityId.ID) error {

	sql := `DELETE FROM shortened_urls WHERE id = @id AND user_id = @userId`
	slog.Info("Deleting shortened url", "id", id, "userId", userId)
	result, err := pr.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"id":     id,
			"userId": userId,
		})

	if err != nil {
		slog.Error("failed to delete shortened url from database", "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("shortened url not found")
	}

	return nil
}

func (pr *ShortenedUrlRepository) Update(shortenedUrl *entity.ShortenedUrl) error {
	sql := `UPDATE shortened_urls 
	SET long_url = @longUrl, 
	    description = @description, 
	    password_hash = @passwordHash, 
	    expires_at = @expiresAt,
	    updated_at = CURRENT_TIMESTAMP
	WHERE id = @id AND user_id = @userId`

	result, err := pr.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"id":           shortenedUrl.ID,
			"userId":       shortenedUrl.UserId,
			"longUrl":      shortenedUrl.LongUrl,
			"description":  shortenedUrl.Description,
			"passwordHash": shortenedUrl.PasswordHash,
			"expiresAt":    shortenedUrl.ExpiresAt,
		})

	if err != nil {
		slog.Error("failed to update shortened url in database", "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("shortened url not found or not owned by user")
	}

	return nil
}
