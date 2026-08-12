package postgres

import (
	"context"
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingRepository struct {
	dbConnection *pgxpool.Pool
}

func NewSettingRepository(dbConnection *pgxpool.Pool) interfaces.SettingRepository {
	return &SettingRepository{dbConnection: dbConnection}
}

func (r *SettingRepository) Save(setting *entity.Setting) error {
	sql := `INSERT INTO app_settings (key, value, updated_at) 
	VALUES (@key, @value, @updated_at)
	ON CONFLICT (key) DO UPDATE SET
		value = EXCLUDED.value,
		updated_at = EXCLUDED.updated_at`

	_, err := r.dbConnection.Exec(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"key":        setting.Key,
			"value":      setting.Value,
			"updated_at": setting.UpdatedAt,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to save setting: %w", err)
	}
	return nil
}

func (r *SettingRepository) FindByKey(key string) (*entity.Setting, error) {
	sql := `SELECT key, value, updated_at FROM app_settings WHERE key = @key LIMIT 1`

	var setting entity.Setting
	err := r.dbConnection.QueryRow(
		context.Background(),
		sql,
		pgx.NamedArgs{
			"key": key,
		},
	).Scan(
		&setting.Key,
		&setting.Value,
		&setting.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("setting not found")
		}
		return nil, fmt.Errorf("failed to find setting by key: %w", err)
	}

	return &setting, nil
}

func (r *SettingRepository) FindAll() ([]*entity.Setting, error) {
	sql := `SELECT key, value, updated_at FROM app_settings`

	rows, err := r.dbConnection.Query(context.Background(), sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	var settings []*entity.Setting
	for rows.Next() {
		var setting entity.Setting
		if err := rows.Scan(&setting.Key, &setting.Value, &setting.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings = append(settings, &setting)
	}

	return settings, nil
}
