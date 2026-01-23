package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	migrationsDir = "file://migrations"
)

func Up() error {

	m, err := migrate.New(migrationsDir, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run up migrations: %v", err)
	}
	return nil
}

func Down() error {

	m, err := migrate.New(migrationsDir, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %v", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run down migrations: %v", err)
	}
	return nil
}

func Create(name string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}

	timestamp := time.Now().Format("20060102150405") // e.g., 202504090001
	upFile := fmt.Sprintf("migrations/%s_%s.up.sql", timestamp, name)
	downFile := fmt.Sprintf("migrations/%s_%s.down.sql", timestamp, name)

	if err := os.MkdirAll("migrations", 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %v", err)
	}

	if err := os.WriteFile(upFile, []byte("-- Up migration\n"), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %v", err)
	}

	if err := os.WriteFile(downFile, []byte("-- Down migration\n"), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %v", err)
	}

	fmt.Printf("Created: %s and %s\n", upFile, downFile)
	return nil
}
