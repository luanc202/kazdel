package handlers

import (
	"net/http"

	"github.com/a-h/templ"
)

func Render(component templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
