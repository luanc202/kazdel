package handlers

import (
	"net/http"
	"strings"

	"kazdel/pkg/infra/config"
	customMiddleware "kazdel/pkg/middleware"
	"kazdel/pkg/ui/pages"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// BuildRouter builds and returns the configured chi.Mux router.
// All handlers registered via Register() are initialized and their routes mounted.
func BuildRouter(deps *Dependencies) (*chi.Mux, error) {
	rootRouter := chi.NewRouter()

	// Global middleware
	rootRouter.Use(middleware.Recoverer)

	logger := config.GetLogger("http")
	rootRouter.Use(customMiddleware.LoggerMiddleware(logger))

	// Create sub-router for the app
	// this enables a chi router to be mounted on a specific path
	// and support subpath
	appRouter := chi.NewRouter()

	rootRouter.Use(middleware.Heartbeat("/health"))

	// Serve static files (HTMX, CSS, etc.)
	appRouter.Handle("/static/*", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := config.GetEnvConfig()
		disableCacheOnLocalEnvironment(env, w)
		http.FileServer(http.Dir("pkg/ui/static")).ServeHTTP(w, r)
	})))

	// Initialize and register all handlers
	for _, h := range GetHandlers() {
		if err := h.Init(deps); err != nil {
			return nil, err
		}

		h.Routes(appRouter)
	}

	// Mount under BASE_PATH if configured (e.g., "/url")
	basePath := config.GetEnvConfig().BASE_PATH
	if basePath != "" && basePath != "/" {
		rootRouter.Mount(basePath, appRouter)
	} else {
		rootRouter.Mount("/", appRouter)
	}

	// Custom error pages
	rootRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		pages.ErrorPage(
			http.StatusNotFound,
			"Page Not Found",
			"The page you're looking for doesn't exist or has been moved.",
		).Render(r.Context(), w)
	})

	rootRouter.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		pages.ErrorPage(
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"This action is not supported.",
		).Render(r.Context(), w)
	})

	return rootRouter, nil
}

func disableCacheOnLocalEnvironment(env *config.EnvConfig, w http.ResponseWriter) {
	if env != nil && strings.ToLower(env.ENV) != "production" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}
}
