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
	if logger == nil {
		return slog.Default().With(slog.String("scope", scope))
	}
	return logger.With(slog.String("scope", scope))
}
