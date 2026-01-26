package routes

import (
	"url-shortener/m/api/controllers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controllers struct {
	ShortenedUrlController *controllers.ShortenedUrlController
}

func InitializeRoutes(controllers Controllers, c *chi.Mux) {

	c.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))

		r.Group(func(private chi.Router) {
			private.Post("/shorten", controllers.ShortenedUrlController.CreateShortenedUrl)
		})

		r.Group(func(public chi.Router) {
			public.Get("/{slug}", controllers.ShortenedUrlController.FindShortenedUrl)
		})
	})
}
