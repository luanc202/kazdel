package handlers

import (
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"kazdel/pkg/ui/pages"
)

// Pages handles public page routes that don't require business logic.
type Pages struct{}

func init() {
	Register(new(Pages))
}

func (h *Pages) Init(deps *Dependencies) error {
	return nil
}

func (h *Pages) Routes(r chi.Router) {
	r.Get("/", templ.Handler(pages.Home()).ServeHTTP)
	r.Get("/login", templ.Handler(pages.Login()).ServeHTTP)
	r.Get("/signup", templ.Handler(pages.SignUp()).ServeHTTP)
}
