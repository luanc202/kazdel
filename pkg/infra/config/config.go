package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "modernc.org/sqlite"
)

var (
	dbpool             *pgxpool.Pool
	sqliteDB           *sql.DB
	logger             *slog.Logger
	StandardDateLayout = "2001-12-01"
)

func InitConfigs() error {
	var err error
	env := GetEnvConfig()

	var handler slog.Handler
	if env.ENV == "DEVELOPMENT" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	logger = slog.New(handler)
	slog.SetDefault(logger)

	if env.GetDatabaseType() == "sqlite" {
		sqliteDB, err = sql.Open("sqlite", env.DBUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to open sqlite database: %v\n", err)
			os.Exit(1)
		}
		sqliteDB.SetMaxOpenConns(1)
		_, err = sqliteDB.Exec("PRAGMA busy_timeout = 5000; PRAGMA foreign_keys = ON;")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to execute pragma on sqlite database: %v\n", err)
			os.Exit(1)
		}
	} else {
		dbpool, err = pgxpool.New(context.Background(), env.DBUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
			os.Exit(1)
		}
	}

	return nil
}

func GetDbConnection() any {
	if GetEnvConfig().GetDatabaseType() == "sqlite" {
		return sqliteDB
	}
	return dbpool
}

func GetPostgresPool() *pgxpool.Pool {
	return dbpool
}

func GetSqliteDB() *sql.DB {
	return sqliteDB
}

func GetLogger(scope string) *slog.Logger {
	if logger == nil {
		return slog.Default().With(slog.String("scope", scope))
	}
	return logger.With(slog.String("scope", scope))
}
