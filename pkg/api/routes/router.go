package routes

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitializeRouter(controllers Controllers) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.Heartbeat("/health"))

	if os.Getenv("ENVIRONMENT") != "DEVELOPMENT" {
		router.Use(middleware.Logger)
	}

	InitializeRoutes(controllers, router)

	return router
}
