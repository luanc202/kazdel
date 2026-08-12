package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
)

type SettingRepository struct {
	dbConnection *sql.DB
}

func NewSettingRepository(dbConnection *sql.DB) interfaces.SettingRepository {
	return &SettingRepository{dbConnection: dbConnection}
}

func (r *SettingRepository) Save(setting *entity.Setting) error {
	query := `INSERT INTO app_settings (key, value, updated_at) 
	VALUES (?, ?, ?)
	ON CONFLICT (key) DO UPDATE SET
		value = EXCLUDED.value,
		updated_at = EXCLUDED.updated_at`

	_, err := r.dbConnection.ExecContext(
		context.Background(),
		query,
		setting.Key,
		setting.Value,
		setting.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save setting: %w", err)
	}
	return nil
}

func (r *SettingRepository) FindByKey(key string) (*entity.Setting, error) {
	query := `SELECT key, value, updated_at FROM app_settings WHERE key = ? LIMIT 1`

	var setting entity.Setting
	err := r.dbConnection.QueryRowContext(
		context.Background(),
		query,
		key,
	).Scan(
		&setting.Key,
		&setting.Value,
		&setting.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("setting not found")
		}
		return nil, fmt.Errorf("failed to find setting by key: %w", err)
	}

	return &setting, nil
}

func (r *SettingRepository) FindAll() ([]*entity.Setting, error) {
	query := `SELECT key, value, updated_at FROM app_settings`

	rows, err := r.dbConnection.QueryContext(context.Background(), query)
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
