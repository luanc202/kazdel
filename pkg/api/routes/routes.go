package routes

import (
	"net/http"
	"strings"

	"kazdel/pkg/api/controllers"
	authMiddleware "kazdel/pkg/api/middleware"
	"kazdel/pkg/handlers"
	"kazdel/pkg/infra/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controllers struct {
	ShortenedUrlController *controllers.ShortenedUrlController
	AuthController         *controllers.AuthController
}

func InitializeRoutes(controllers Controllers, webHandlers handlers.Handlers, c *chi.Mux) {

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
		web.Get("/", webHandlers.Home.ServeHTTP)
		web.Get("/login", webHandlers.Auth.ShowLoginPage)
		web.Get("/signup", webHandlers.Auth.ShowSignupPage)

		web.Post("/login", webHandlers.Auth.HandleLogin)
		web.Post("/signup", webHandlers.Auth.HandleSignup)
		web.Post("/logout", webHandlers.Auth.HandleLogout)

		// Protected web routes
		web.Group(func(protectedWeb chi.Router) {
			protectedWeb.Use(authMiddleware.AuthMiddleware(controllers.AuthController.AuthUseCase))

			protectedWeb.Get("/dashboard", webHandlers.ShortenedUrl.ShowDashboard)
			protectedWeb.Post("/shorten", webHandlers.ShortenedUrl.HandleCreateUrl)
			protectedWeb.Delete("/urls/{id}", webHandlers.ShortenedUrl.HandleDeleteUrl)
		})
	})

	// JSON API Routes
	c.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))

			// Auth API Routes
			r.Group(func(auth chi.Router) {
				auth.Post("/auth/signup", controllers.AuthController.Signup)
				auth.Post("/auth/login", controllers.AuthController.Login)
			})

			// Protected API Routes
			r.Group(func(private chi.Router) {
				private.Use(authMiddleware.AuthMiddleware(controllers.AuthController.AuthUseCase))
				private.Post("/auth/logout", controllers.AuthController.Logout)
				private.Post("/shorten", controllers.ShortenedUrlController.CreateShortenedUrl)
				private.Get("/urls", controllers.ShortenedUrlController.ListUserUrls)
				private.Delete("/urls/{id}", controllers.ShortenedUrlController.DeleteShortenedUrl)
			})
		})
	})

	// Global Redirect Route for short links
	c.Get("/{slug}", controllers.ShortenedUrlController.RedirectToLongUrl)
}
