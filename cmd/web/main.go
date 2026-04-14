package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"kazdel/pkg/handlers"
	"kazdel/pkg/infra/config"
	"kazdel/pkg/infra/db"
	"kazdel/pkg/usecase"
)

func main() {
	env, err := config.LoadEnv(".")
	if err != nil {
		panic(err)
	}

	err = config.InitConfigs()
	if err != nil {
		panic(err)
	}

	dbConn := config.GetDbConnection()

	// Create repositories
	shortenedURLsRepo := db.NewShortenedUrlRepository(dbConn)
	userRepo := db.NewUserRepository(dbConn)
	sessionRepo := db.NewPostgresSessionRepository(dbConn)

	// Create use cases
	shortenedURLUseCase := usecase.NewShortenedUrlUseCase(shortenedURLsRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo)

	// Create handler dependencies
	deps := &handlers.Dependencies{
		ShortenedUrlUseCase: shortenedURLUseCase,
		AuthUseCase:         authUseCase,
	}

	// Build router (handlers auto-register via init())
	router, err := handlers.BuildRouter(deps)
	if err != nil {
		panic(fmt.Sprintf("failed to build router: %v", err))
	}

	slog.Info(fmt.Sprintf("Server started on port: %v", env.PORT))
	slog.Error(http.ListenAndServe(":"+env.PORT, router).Error())
}
