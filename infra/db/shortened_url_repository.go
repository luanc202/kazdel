package db

import (
	"context"
	"log/slog"
	"url-shortener/m/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShortenedUrlRepository struct {
	dbConnection *pgxpool.Pool
}

func NewShortenedUrlRepository(dbConn *pgxpool.Pool) *ShortenedUrlRepository {
	return &ShortenedUrlRepository{
		dbConnection: dbConn,
	}
}

func (pr *ShortenedUrlRepository) Save(shortenedUrl *entity.ShortenedUrl) error {

	sql := `INSERT INTO shortened_urls
	(short_slug, original_url, expires_at, user_id)
	VALUES
	(@shortSlug, @originalUrl, @expiresAt, @userId)`

	_, err := pr.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"shortSlug":   shortenedUrl.ShortSlug,
			"originalUrl": shortenedUrl.OriginalUrl,
			"expiresAt":   shortenedUrl.ExpiresAt,
			"userId":      shortenedUrl.UserId,
		})

	if err != nil {
		slog.Error("failed to insert product into database", "error", err)
		return err
	}

	return nil
}

func (pr *ShortenedUrlRepository) FindByID(id uint64) *entity.ShortenedUrl {
	return nil
}

func (pr *ShortenedUrlRepository) Update(shortenedUrlId uint64, shortenedUrlToUpdate *entity.ShortenedUrl) error {
	return nil
}
