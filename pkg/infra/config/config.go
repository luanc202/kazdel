package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbpool             *pgxpool.Pool
	logger             *slog.Logger
	StandardDateLayout = "2001-12-01"
)

func InitConfigs() error {
	var err error
	env := GetEnvConfig()

	dbpool, err = pgxpool.New(context.Background(), env.DBUrl)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return nil
}

func GetDbConnection() *pgxpool.Pool {
	return dbpool
}

func GetLogger(scope string) *slog.Logger {
	logger = slog.Default()
	return logger
}
