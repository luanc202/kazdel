package db

import (
	"context"
	"log/slog"
	"url-shortener/m/entity"
	interfaces "url-shortener/m/interface"

	"github.com/jackc/pgx/v5"
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
	(short_slug, long_url, expires_at, user_id)
	VALUES
	(@shortSlug, @longUrl, @expiresAt, @userId)`

	_, err := pr.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"shortSlug": shortenedUrl.ShortSlug,
			"longUrl":   shortenedUrl.LongUrl,
			"expiresAt": shortenedUrl.ExpiresAt,
			"userId":    shortenedUrl.UserId,
		})

	if err != nil {
		slog.Error("failed to insert product into database", "error", err)
		return err
	}

	return nil
}

func (pr *ShortenedUrlRepository) FindBySlug(slug string) (*entity.ShortenedUrl, error) {

	sql := `SELECT short_slug, long_url, expires_at, user_id FROM shortened_urls WHERE short_slug = @slug
	LIMIT 1`

	var shortenedUrl entity.ShortenedUrl

	err := pr.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"slug": slug,
		}).Scan(
		&shortenedUrl.ShortSlug,
		&shortenedUrl.LongUrl,
		&shortenedUrl.ExpiresAt,
		&shortenedUrl.UserId,
	)

	if err != nil {
		slog.Error("failed to find shortened url by slug", "error", err)
		return nil, err
	}

	return &shortenedUrl, nil
}
