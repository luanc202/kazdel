package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"kazdel/pkg/handlers"
	"kazdel/pkg/infra/config"
	"kazdel/pkg/infra/db"
	"kazdel/pkg/infra/mail"
	"kazdel/pkg/usecase"

	"github.com/oschwald/geoip2-golang"
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

	// Initialize GeoIP2 Reader
	var geoipDb *geoip2.Reader
	geoipPath := "data/GeoLite2-Country.mmdb"
	dbReader, err := geoip2.Open(geoipPath)
	if err == nil {
		geoipDb = dbReader
		defer geoipDb.Close()
	} else {
		slog.Warn("Failed to open GeoIP database, geolocation tracking will be disabled", "error", err)
	}

	// Create repositories
	shortenedURLsRepo := db.NewShortenedUrlRepository(dbConn)
	urlVisitRepo := db.NewUrlVisitRepository(dbConn)
	userRepo := db.NewUserRepository(dbConn)
	sessionRepo := db.NewSessionRepository(dbConn)
	userTokenRepo := db.NewUserTokenRepository(dbConn)

	// Create services
	emailService := mail.NewSMTPMailService()

	// Create use cases
	shortenedURLUseCase := usecase.NewShortenedUrlUseCase(shortenedURLsRepo, urlVisitRepo, geoipDb, emailService)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, userTokenRepo, emailService)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := shortenedURLUseCase.CleanExpiredURLs(); err != nil {
				slog.Error("Failed to clean expired URLs", "error", err)
			}
		}
	}()

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
