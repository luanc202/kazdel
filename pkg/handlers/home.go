package handlers

import (
	"context"
	"net/http"

	"kazdel/pkg/ui/pages"
)

type HomePageHandler struct {
	ctx context.Context
	w   http.ResponseWriter
	r   *http.Request
}

func NewHomePageHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) *HomePageHandler {
	return &HomePageHandler{
		ctx: ctx,
		w:   w,
		r:   r,
	}
}

func (h *HomePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := pages.Home().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
