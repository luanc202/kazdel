package controllers

import (
	"fmt"
	"net/http"
	"time"

	authMiddleware "kazdel/pkg/api/middleware"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/ui/pages"
	"kazdel/pkg/uniqueEntityId"
	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form"
)

type ShortenedUrlController struct {
	Usecase *usecase.ShortenedUrlUsecase
}

func NewShortenedUrlController(usecase *usecase.ShortenedUrlUsecase) *ShortenedUrlController {
	return &ShortenedUrlController{
		Usecase: usecase,
	}
}

// RedirectToLongUrl simply redirects public users who click a shortened link
func (cntrl *ShortenedUrlController) RedirectToLongUrl(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	shortenedUrl, err := cntrl.Usecase.FindBySlug(slug)
	if err != nil {
		http.Error(w, "Shortened URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, shortenedUrl.LongUrl, http.StatusFound)
}

// ServeDashboardPage renders the main UI dashboard fully SSR
func (cntrl *ShortenedUrlController) ServeDashboardPage(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	urls, err := cntrl.Usecase.ListByUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response []dto.ShortenedUrlView
	for _, u := range urls {
		response = append(response, *dto.NewShortenedUrlView(u.ShortSlug, u.LongUrl, u.ExpiresAt.Format("2006-01-02 15:04:05")))
	}

	pages.Dashboard(response).Render(r.Context(), w)
}

// CreateShortenedUrl handles form submissions to create URLs, returning HTMX fragments
func (cntrl *ShortenedUrlController) CreateShortenedUrl(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		pages.CreateUrlForm("Invalid form submission", "", time.Now().AddDate(0, 0, 7).Format("2006-01-02T15:04")).Render(r.Context(), w)
		return
	}

	var shortenedUrlInsert dto.ShortenedUrlInsert
	decoder := form.NewDecoder()
	if err := decoder.Decode(&shortenedUrlInsert, r.Form); err != nil {
		originalUrl := r.FormValue("originalUrl")
		expiresAt := r.FormValue("expiresAt")
		pages.CreateUrlForm("Form decoding error", originalUrl, expiresAt).Render(r.Context(), w)
		return
	}

	err = shortenedUrlInsert.Validate()
	if err != nil {
		pages.CreateUrlForm(err.Error(), shortenedUrlInsert.OriginalUrl, shortenedUrlInsert.ExpiresAt).Render(r.Context(), w)
		return
	}

	err = cntrl.Usecase.Save(shortenedUrlInsert, userId)
	if err != nil {
		pages.CreateUrlForm(fmt.Sprintf("Failed to save: %s", err.Error()), shortenedUrlInsert.OriginalUrl, shortenedUrlInsert.ExpiresAt).Render(r.Context(), w)
		return
	}

	// Tell HTMX to do a full page transition back to dashboard to grab the fresh list
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// DeleteShortenedUrl removes a URL and returns 200 OK allowing HTMX to sweep the DOM
func (cntrl *ShortenedUrlController) DeleteShortenedUrl(w http.ResponseWriter, r *http.Request) {
	shortSlug := chi.URLParam(r, "slug")
	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	url, err := cntrl.Usecase.FindBySlug(shortSlug)
	if err != nil || url.UserId != userId {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Delete takes ID, but chi.URLParam gives us "slug". We passed ID to usecase Delete!
	// Oh! Usecase.Delete requires the uint64 ID! We have it now at `url.ID`.
	err = cntrl.Usecase.Delete(url.ID, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
