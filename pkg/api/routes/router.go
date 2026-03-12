package routes

import (
	customMiddleware "kazdel/pkg/api/middleware"
	"kazdel/pkg/infra/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitializeRouter(controllers Controllers) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)

	logger := config.GetLogger("http")
	router.Use(customMiddleware.LoggerMiddleware(logger))

	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.Heartbeat("/health"))

	InitializeRoutes(controllers, router)

	return router
}
