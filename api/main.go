package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"url-shortener/m/api/controllers"
	"url-shortener/m/api/routes"
	"url-shortener/m/infra/config"
	"url-shortener/m/infra/db"
	"url-shortener/m/usecase"
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

	shortenedURLUseCase := usecase.NewShortenedUrlUseCase(shortenedURLsRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, env.JWT_SECRET)

	shortenedURLController := controllers.NewShortenedUrlController(shortenedURLUseCase)
	authController := controllers.NewAuthController(authUseCase)

	controllers := routes.Controllers{
		ShortenedUrlController: shortenedURLController,
		AuthController:         authController,
	}

	router := routes.InitializeRouter(controllers)

	slog.Info(fmt.Sprintf("Server started on port: %v", env.PORT))
	slog.Error(http.ListenAndServe(":"+env.PORT, router).Error())
}
