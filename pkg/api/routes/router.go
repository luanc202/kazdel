package routes

import (
	customMiddleware "kazdel/pkg/api/middleware"
	"kazdel/pkg/handlers"
	"kazdel/pkg/infra/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitializeRouter(controllers Controllers, webHandlers handlers.Handlers) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)

	logger := config.GetLogger("http")
	router.Use(customMiddleware.LoggerMiddleware(logger))

	router.Use(middleware.Heartbeat("/health"))

	InitializeRoutes(controllers, webHandlers, router)

	return router
}
