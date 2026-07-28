package sqlite

import (
	"context"
	"database/sql"
	"kazdel/pkg/entity"
	"kazdel/pkg/entity/dto"
	interfaces "kazdel/pkg/interface"
)

type UrlVisitRepository struct {
	db *sql.DB
}

func NewUrlVisitRepository(db *sql.DB) interfaces.UrlVisitRepository {
	return &UrlVisitRepository{db: db}
}

func (r *UrlVisitRepository) Save(ctx context.Context, visit *entity.UrlVisit) error {
	query := `
		INSERT INTO url_visits (url_id, ip_address, referrer, user_agent, browser, os, country)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		visit.UrlId,
		visit.IpAddress,
		visit.Referrer,
		visit.UserAgent,
		visit.Browser,
		visit.Os,
		visit.Country,
	)
	return err
}

func (r *UrlVisitRepository) GetStatsByUrlId(ctx context.Context, urlId uint64) (*dto.UrlStats, error) {
	stats := dto.NewUrlStats()

	// Get total clicks
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM url_visits WHERE url_id = ?", urlId).Scan(&stats.TotalClicks)
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
	query := "SELECT COALESCE(" + column + ", 'Unknown'), COUNT(*) FROM url_visits WHERE url_id = ? GROUP BY " + column

	rows, err := r.db.QueryContext(ctx, query, urlId)
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
