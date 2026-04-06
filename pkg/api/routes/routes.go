package routes

import (
	"net/http"
	"strings"

	"kazdel/pkg/api/controllers"
	authMiddleware "kazdel/pkg/api/middleware"
	"kazdel/pkg/infra/config"
	"kazdel/pkg/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controllers struct {
	ShortenedUrlController *controllers.ShortenedUrlController
	AuthController         *controllers.AuthController
}

func InitializeRoutes(controllers Controllers, c *chi.Mux) {

	// Serve static files (HTMX, CSS, etc.)
	c.Handle("/static/*", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := config.GetEnvConfig()
		if env != nil && strings.ToLower(env.ENV) != "production" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		http.FileServer(http.Dir("pkg/ui/static")).ServeHTTP(w, r)
	})))

	// Web UI Routes (Templ + HTMX)
	c.Group(func(web chi.Router) {
		// Public web routes
		web.Get("/", templ.Handler(pages.Home()).ServeHTTP)
		web.Get("/login", templ.Handler(pages.Login()).ServeHTTP)
		web.Get("/signup", templ.Handler(pages.SignUp()).ServeHTTP)

		// Protected web routes
		web.Group(func(protectedWeb chi.Router) {
			protectedWeb.Use(authMiddleware.AuthMiddleware(controllers.AuthController.AuthUseCase))

			protectedWeb.Get("/dashboard", controllers.ShortenedUrlController.ServeDashboardPage)
			protectedWeb.Post("/dashboard/urls/shorten", controllers.ShortenedUrlController.CreateShortenedUrl)
			protectedWeb.Delete("/dashboard/urls/{slug}", controllers.ShortenedUrlController.DeleteShortenedUrl)
		})
	})

	// JSON API Routes
	c.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/x-www-form-urlencoded"))

			// Auth API Routes
			r.Group(func(auth chi.Router) {
				auth.Post("/auth/signup", controllers.AuthController.Signup)
				auth.Post("/auth/login", controllers.AuthController.Login)
			})

			// Protected API Routes
			r.Group(func(private chi.Router) {
				private.Use(authMiddleware.AuthMiddleware(controllers.AuthController.AuthUseCase))
				// We allow standard GET for logout matching the anchor tag on dashboard
				private.Get("/auth/logout", controllers.AuthController.Logout)
				private.Post("/auth/logout", controllers.AuthController.Logout)
			})
		})
	})

	// Global Redirect Route for short links
	c.Get("/{slug}", controllers.ShortenedUrlController.RedirectToLongUrl)
}
