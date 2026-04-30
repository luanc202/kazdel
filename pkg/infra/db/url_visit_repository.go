package db

import (
	"context"
	"kazdel/pkg/entity"
	"kazdel/pkg/entity/dto"
	interfaces "kazdel/pkg/interface"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UrlVisitRepository struct {
	db *pgxpool.Pool
}

func NewUrlVisitRepository(db *pgxpool.Pool) interfaces.UrlVisitRepository {
	return &UrlVisitRepository{db: db}
}

func (r *UrlVisitRepository) Save(ctx context.Context, visit *entity.UrlVisit) error {
	query := `
		INSERT INTO url_visits (url_id, ip_address, referrer, user_agent, browser, os, country)
		VALUES (@url_id, @ip_address, @referrer, @user_agent, @browser, @os, @country)
	`
	args := pgx.NamedArgs{
		"url_id":     visit.UrlId,
		"ip_address": visit.IpAddress,
		"referrer":   visit.Referrer,
		"user_agent": visit.UserAgent,
		"browser":    visit.Browser,
		"os":         visit.Os,
		"country":    visit.Country,
	}

	_, err := r.db.Exec(ctx, query, args)
	return err
}

func (r *UrlVisitRepository) GetStatsByUrlId(ctx context.Context, urlId uint64) (*dto.UrlStats, error) {
	stats := dto.NewUrlStats()

	// Get total clicks
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM url_visits WHERE url_id = $1", urlId).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, err
	}

	if stats.TotalClicks == 0 {
		return stats, nil
	}

	// Fetch aggregations
	err = r.aggregateMap(ctx, urlId, "browser", stats.BrowserStats)
	if err != nil {
		return nil, err
	}

	err = r.aggregateMap(ctx, urlId, "os", stats.OsStats)
	if err != nil {
		return nil, err
	}

	err = r.aggregateMap(ctx, urlId, "country", stats.CountryStats)
	if err != nil {
		return nil, err
	}

	err = r.aggregateMap(ctx, urlId, "referrer", stats.ReferrerStats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *UrlVisitRepository) aggregateMap(ctx context.Context, urlId uint64, column string, targetMap map[string]int) error {
	// Query to group and count
	// Note: We use string formatting for the column name because it's an internal trusted string
	query := "SELECT COALESCE(" + column + ", 'Unknown'), COUNT(*) FROM url_visits WHERE url_id = $1 GROUP BY " + column
	
	rows, err := r.db.Query(ctx, query, urlId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var count int
		if err := rows.Scan(&key, &count); err != nil {
			return err
		}
		if key == "" {
			key = "Unknown"
		}
		targetMap[key] = count
	}

	return rows.Err()
}
