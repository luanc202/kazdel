package internal

import (
	"database/sql"
	"fmt"
	"kazdel/pkg/infra/config"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func getMigrationsDir(env *config.EnvConfig) string {
	base := "migrations"
	if env != nil && env.MIGRATIONS_PATH != "" {
		base = env.MIGRATIONS_PATH
	}
	if strings.HasSuffix(base, "/postgres") || strings.HasSuffix(base, "/sqlite") {
		return "file://" + base
	}
	if env.GetDatabaseType() == "sqlite" {
		return "file://" + base + "/sqlite"
	}
	return "file://" + base + "/postgres"
}

func getMigrationDriver(env *config.EnvConfig) (database.Driver, *sql.DB, string, error) {
	dbType := env.GetDatabaseType()
	migrationsDir := getMigrationsDir(env)

	if dbType == "sqlite" {
		db, err := sql.Open("sqlite", env.DBUrl_Migration)
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to connect to sqlite database: %v", err)
		}
		driver, err := sqlite.WithInstance(db, &sqlite.Config{
			MigrationsTable: "schema_migrations",
		})
		if err != nil {
			db.Close()
			return nil, nil, "", fmt.Errorf("failed to create sqlite driver: %v", err)
		}
		return driver, db, migrationsDir, nil
	}

	db, err := sql.Open("postgres", env.DBUrl_Migration)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to connect to postgres database: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		db.Close()
		return nil, nil, "", fmt.Errorf("failed to create postgres driver: %v", err)
	}
	return driver, db, migrationsDir, nil
}

func Up() error {
	env, err := config.LoadEnv("../")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v\n", err)
	}
	driver, db, migrationsDir, err := getMigrationDriver(env)
	if err != nil {
		log.Fatalf("Failed to get migration driver: %v\n", err)
	}
	defer db.Close()
	dbName := env.GetDatabaseType()
	migration, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		dbName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %v", err)
	}
	defer migration.Close()
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

	driver, db, migrationsDir, err := getMigrationDriver(env)
	if err != nil {
		log.Fatalf("Failed to get migration driver: %v\n", err)
	}

	defer db.Close()

	dbName := env.GetDatabaseType()
	migration, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		dbName,
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

	env, _ := config.LoadEnv("../")
	dbType := "sqlite"
	if env != nil {
		dbType = env.GetDatabaseType()
	}
	dir := "migrations/" + dbType
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %v", err)
	}
	timestamp := time.Now().Format("20060102150405") // e.g., 202504090001
	upFile := fmt.Sprintf("%s/%s_%s.up.sql", dir, timestamp, name)
	downFile := fmt.Sprintf("%s/%s_%s.down.sql", dir, timestamp, name)
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

	driver, db, migrationsDir, err := getMigrationDriver(env)
	if err != nil {

		log.Fatalf("Failed to get migration driver: %v\n", err)
	}

	defer db.Close()

	dbName := env.GetDatabaseType()
	migration, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		dbName,
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
