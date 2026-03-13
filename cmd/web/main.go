package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"kazdel/pkg/api/controllers"
	"kazdel/pkg/api/routes"
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

	shortenedURLsRepo := db.NewShortenedUrlRepository(dbConn)
	userRepo := db.NewUserRepository(dbConn)
	sessionRepo := db.NewPostgresSessionRepository(dbConn)

	shortenedURLUseCase := usecase.NewShortenedUrlUseCase(shortenedURLsRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo)

	shortenedURLController := controllers.NewShortenedUrlController(shortenedURLUseCase)
	authController := controllers.NewAuthController(authUseCase)

	apiControllers := routes.Controllers{
		ShortenedUrlController: shortenedURLController,
		AuthController:         authController,
	}

	webHandlers := handlers.Handlers{
		Home:         handlers.NewHomePageHandler(context.Background(), nil, nil), // Dummy for now, actual implementation likely needs http context injected dynamically or refactoring
		Auth:         handlers.NewAuthHandler(authUseCase),
		ShortenedUrl: handlers.NewShortenedUrlHandler(shortenedURLUseCase),
	}

	router := routes.InitializeRouter(apiControllers, webHandlers)

	slog.Info(fmt.Sprintf("Server started on port: %v", env.PORT))
	slog.Error(http.ListenAndServe(":"+env.PORT, router).Error())
}
