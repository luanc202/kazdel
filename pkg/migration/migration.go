package internal

import (
	"database/sql"
	"fmt"
	"kazdel/pkg/infra/config"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
)

const (
	migrationsDir = "file://migrations"
)

func Up() error {

	env, err := config.LoadEnv("../")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v\n", err)
	}
	fmt.Println(env.DBUrl_Migration)
	db, err := sql.Open("postgres", env.DBUrl_Migration)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v\n", err)
		}
	}()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"pgxv5",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize migration: %v", err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run up migrations: %v", err)
	}
	return nil
}

func Down() error {

	env, err := config.LoadEnv("../")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v\n", err)
	}

	db, err := sql.Open("postgres", env.DBUrl_Migration)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v\n", err)
		}
	}()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"postgres",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize migration: %v", err)
	}
	defer migration.Close()

	if err := migration.Down(); err != nil && err != migrate.ErrNoChange {
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

func Force(version int) error {

	env, err := config.LoadEnv("../")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v\n", err)
	}

	db, err := sql.Open("postgres", env.DBUrl_Migration)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v\n", err)
		}
	}()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"postgres",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize migration: %v", err)
	}
	defer migration.Close()

	if err := migration.Force(version); err != nil {
		return fmt.Errorf("failed to run force migration: %v", err)
	}
	return nil
}
