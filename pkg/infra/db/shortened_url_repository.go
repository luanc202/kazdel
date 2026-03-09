package db

import (
	"context"
	"errors"
	"log/slog"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"

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

func (pr *ShortenedUrlRepository) FindByUserId(userId uniqueEntityId.ID) ([]*entity.ShortenedUrl, error) {
	sql := `SELECT short_slug, long_url, expires_at, user_id FROM shortened_urls WHERE user_id = @userId`

	rows, err := pr.dbConnection.Query(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"userId": userId,
		},
	)

	if err != nil {
		slog.Error("failed to find shortened urls by user id", "error", err)
		return nil, err
	}
	defer rows.Close()

	var shortenedUrls []*entity.ShortenedUrl

	for rows.Next() {
		var shortenedUrl entity.ShortenedUrl
		err := rows.Scan(
			&shortenedUrl.ShortSlug,
			&shortenedUrl.LongUrl,
			&shortenedUrl.ExpiresAt,
			&shortenedUrl.UserId,
		)
		if err != nil {
			slog.Error("failed to scan shortened url", "error", err)
			return nil, err
		}
		shortenedUrls = append(shortenedUrls, &shortenedUrl)
	}

	return shortenedUrls, nil
}

func (pr *ShortenedUrlRepository) Delete(id uint64, userId uniqueEntityId.ID) error {

	sql := `DELETE FROM shortened_urls WHERE id = @id AND user_id = @userId`

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
