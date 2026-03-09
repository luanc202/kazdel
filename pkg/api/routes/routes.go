package routes

import (
	"net/http"
	"kazdel/pkg/api/controllers"
	authMiddleware "kazdel/pkg/api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controllers struct {
	ShortenedUrlController *controllers.ShortenedUrlController
	AuthController         *controllers.AuthController
}

func InitializeRoutes(controllers Controllers, c *chi.Mux) {

	// Serve static files (HTMX, CSS, etc.)
	c.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("pkg/ui/static"))))

	c.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))

			// Auth Routes
			r.Group(func(auth chi.Router) {
				auth.Post("/auth/signup", controllers.AuthController.Signup)
				auth.Post("/auth/login", controllers.AuthController.Login)
			})

			// Protected Routes
			r.Group(func(private chi.Router) {
				private.Use(authMiddleware.AuthMiddleware(controllers.AuthController.AuthUseCase))
				private.Post("/auth/logout", controllers.AuthController.Logout)
				private.Post("/shorten", controllers.ShortenedUrlController.CreateShortenedUrl)
				private.Get("/urls", controllers.ShortenedUrlController.ListUserUrls)
				private.Delete("/urls/{id}", controllers.ShortenedUrlController.DeleteShortenedUrl)
			})
		})
	})

	c.Get("/{slug}", controllers.ShortenedUrlController.RedirectToLongUrl)
}
