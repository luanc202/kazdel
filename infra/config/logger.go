package config

import (
	"log"
	"log/slog"
	"os"

	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/traceid"
	"github.com/golang-cz/devslog"
)

func configLogger() {
	env, err := LoadEnv("../")
	if err != nil {
		log.Fatalf("Failed to load environment file: %v\n", err)
	}
	isLocalhost := env.ENV == "localhost" || env.ENV == "development" || env.ENV == "test" || env.ENV == "staging" || env.ENV == ""

	logFormat := httplog.SchemaECS.Concise(isLocalhost)

	logger := slog.New(logHandler(isLocalhost, &slog.HandlerOptions{
		AddSource:   !isLocalhost,
		ReplaceAttr: logFormat.ReplaceAttr,
	}))

	if !isLocalhost {
		logger = logger.With(
			slog.String("app", APP_NAME),
			slog.String("version", VERSION),
			slog.String("env", env.ENV),
		)
	}

	// Set as a default logger for both slog and log.
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelError)
}

func logHandler(isLocalhost bool, handlerOpts *slog.HandlerOptions) slog.Handler {
	if isLocalhost {
		// Pretty logs for localhost development.
		return devslog.NewHandler(os.Stdout, &devslog.Options{
			SortKeys:           true,
			MaxErrorStackTrace: 5,
			MaxSlicePrintSize:  20,
			HandlerOptions:     handlerOpts,
		})
	}

	// JSON logs for production with "traceId".
	return traceid.LogHandler(
		slog.NewJSONHandler(os.Stdout, handlerOpts),
	)
}
