package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	appctx "kazdel/pkg/context"
	"kazdel/pkg/entity"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/middleware"
	"kazdel/pkg/ui/components"
	"kazdel/pkg/ui/pages"
	"kazdel/pkg/uniqueEntityId"
	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form"
	"golang.org/x/crypto/bcrypt"
)

// ShortenedUrl handles URL shortening HTTP routes.
type ShortenedUrl struct {
	usecase     *usecase.ShortenedUrlUsecase
	authUseCase *usecase.AuthUseCase
}

func init() {
	Register(new(ShortenedUrl))
}

func (h *ShortenedUrl) Init(deps *Dependencies) error {
	h.usecase = deps.ShortenedUrlUseCase
	h.authUseCase = deps.AuthUseCase
	return nil
}

func (h *ShortenedUrl) Routes(r chi.Router) {
	// Protected dashboard routes
	r.Group(func(protectedWeb chi.Router) {
		protectedWeb.Use(middleware.AuthMiddleware(h.authUseCase))

		protectedWeb.Get("/dashboard", h.DashboardPage)
		protectedWeb.Post("/dashboard/urls/shorten", h.CreateUrl)
		protectedWeb.Delete("/dashboard/urls/{slug}", h.DeleteUrl)
		protectedWeb.Get("/dashboard/urls/{slug}/edit", h.EditUrlForm)
		protectedWeb.Post("/dashboard/urls/{slug}/edit", h.UpdateUrl)
		protectedWeb.Get("/dashboard/urls/{slug}/stats", h.GetUrlStatsPage)
		protectedWeb.Get("/dashboard/urls/{slug}", h.GetUrlRow)
	})

	// Public redirect route for short links (must be registered last to avoid
	// catching other routes)
	r.Get("/{slug}", h.RedirectToLongUrl)
	r.Post("/{slug}/password", h.HandlePasswordSubmission)
}

// RedirectToLongUrl handles GET /{slug}
// @Summary Redirect to the original long URL
// @Param slug path string true "Short URL slug"
// @Success 302 "Redirect to original URL"
// @Failure 404 "Shortened URL not found"
// @Router /{slug} [get]
func (h *ShortenedUrl) RedirectToLongUrl(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	shortenedUrl, err := h.usecase.FindBySlug(slug)
	if err != nil {
		pages.ErrorPage(http.StatusNotFound, "Not Found", "The requested URL could not be found.").Render(r.Context(), w)
		return
	}

	if shortenedUrl.PasswordHash != nil {
		components.PasswordPrompt(slug, "").Render(r.Context(), w)
		return
	}

	go h.usecase.RecordVisit(context.Background(), shortenedUrl.ID, r)

	http.Redirect(w, r, shortenedUrl.LongUrl, http.StatusFound)
}

func (h *ShortenedUrl) HandlePasswordSubmission(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	shortenedUrl, err := h.usecase.FindBySlug(slug)
	if err != nil {
		pages.ErrorPage(http.StatusNotFound, "Not Found", "The requested URL could not be found.").Render(r.Context(), w)
		return
	}

	if shortenedUrl.PasswordHash == nil {
		go h.usecase.RecordVisit(context.Background(), shortenedUrl.ID, r)
		http.Redirect(w, r, shortenedUrl.LongUrl, http.StatusFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		components.PasswordPrompt(slug, "Invalid form submission").Render(r.Context(), w)
		return
	}

	password := r.FormValue("password")
	err = bcrypt.CompareHashAndPassword([]byte(*shortenedUrl.PasswordHash), []byte(password))
	if err != nil {
		components.PasswordPrompt(slug, "Incorrect password").Render(r.Context(), w)
		return
	}

	go h.usecase.RecordVisit(context.Background(), shortenedUrl.ID, r)

	http.Redirect(w, r, shortenedUrl.LongUrl, http.StatusFound)
}

// DashboardPage handles GET /dashboard
// @Summary Render the dashboard page with the user's URLs
// @Success 200 "Dashboard HTML page"
// @Failure 401 "Unauthorized"
// @Failure 500 "Internal server error"
// @Router /dashboard [get]
func (h *ShortenedUrl) DashboardPage(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	urls, err := h.usecase.ListByUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response []dto.ShortenedUrlView
	for _, u := range urls {
		desc := ""
		if u.Description != nil {
			desc = *u.Description
		}
		response = append(response, *dto.NewShortenedUrlView(
			u.ShortSlug,
			u.LongUrl,
			u.ExpiresAt.Format("2006-01-02 15:04:05"),
			desc,
			u.PasswordHash != nil,
			u.ExpiresAt.Format("2006-01-02T15:04"),
		))
	}

	pages.Dashboard(response).Render(r.Context(), w)
}

// CreateUrl handles POST /dashboard/urls/shorten
// @Summary Create a new shortened URL
// @Accept application/x-www-form-urlencoded
// @Produce text/html
// @Param originalUrl formData string true "Original URL to shorten"
// @Param expiresAt formData string true "Expiration date (YYYY-MM-DDThh:mm)"
// @Success 200 "Redirect to dashboard via HX-Redirect"
// @Failure 400 "Validation error"
// @Failure 401 "Unauthorized"
// @Router /dashboard/urls/shorten [post]
func (h *ShortenedUrl) CreateUrl(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		pages.CreateUrlForm("Invalid form submission", "", time.Now().AddDate(0, 0, 7).Format("2006-01-02T15:04"), "", "").Render(r.Context(), w)
		return
	}

	var shortenedUrlInsert dto.ShortenedUrlInsert
	decoder := form.NewDecoder()
	if err := decoder.Decode(&shortenedUrlInsert, r.Form); err != nil {
		originalUrl := r.FormValue("originalUrl")
		expiresAt := r.FormValue("expiresAt")
		customSlug := r.FormValue("customSlug")
		description := r.FormValue("description")
		pages.CreateUrlForm("Form decoding error", originalUrl, expiresAt, customSlug, description).Render(r.Context(), w)
		return
	}

	err = shortenedUrlInsert.Validate()
	if err != nil {
		pages.CreateUrlForm(err.Error(), shortenedUrlInsert.OriginalUrl, shortenedUrlInsert.ExpiresAt, shortenedUrlInsert.CustomSlug, shortenedUrlInsert.Description).Render(r.Context(), w)
		return
	}

	shortSlug, err := h.usecase.Save(shortenedUrlInsert, userId)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to save: %s", err.Error())
		if errors.Is(err, entity.ErrCustomSlugAlreadyExists) {
			errMsg = "This custom slug is already taken. Please choose another one."
		}
		pages.CreateUrlForm(errMsg, shortenedUrlInsert.OriginalUrl, shortenedUrlInsert.ExpiresAt, shortenedUrlInsert.CustomSlug, shortenedUrlInsert.Description).Render(r.Context(), w)
		return
	}

	// Fetch the newly created URL to render the new row
	url, err := h.usecase.FindBySlug(shortSlug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	desc := ""
	if url.Description != nil {
		desc = *url.Description
	}

	view := dto.NewShortenedUrlView(
		url.ShortSlug,
		url.LongUrl,
		url.ExpiresAt.Format("2006-01-02 15:04:05"),
		desc,
		url.PasswordHash != nil,
		url.ExpiresAt.Format("2006-01-02T15:04"),
	)

	clearedForm := pages.CreateUrlForm("", "", time.Now().AddDate(0, 0, 7).Format("2006-01-02T15:04"), "", "")
	newRow := pages.NewUrlRow(*view)

	pages.CreateUrlSuccess(clearedForm, newRow).Render(r.Context(), w)
}

// DeleteUrl handles DELETE /dashboard/urls/{slug}
// @Summary Delete a shortened URL
// @Param slug path string true "Short URL slug"
// @Success 200 "URL deleted successfully"
// @Failure 401 "Unauthorized"
// @Failure 404 "URL not found"
// @Failure 500 "Internal server error"
// @Router /dashboard/urls/{slug} [delete]
func (h *ShortenedUrl) DeleteUrl(w http.ResponseWriter, r *http.Request) {
	shortSlug := chi.URLParam(r, "slug")

	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	url, err := h.usecase.FindBySlug(shortSlug)
	if err != nil || url.UserId != userId {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = h.usecase.Delete(url.ID, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// EditUrlForm handles GET /dashboard/urls/{slug}/edit
func (h *ShortenedUrl) EditUrlForm(w http.ResponseWriter, r *http.Request) {
	shortSlug := chi.URLParam(r, "slug")

	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	url, err := h.usecase.FindBySlug(shortSlug)
	if err != nil || url.UserId != userId {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	desc := ""
	if url.Description != nil {
		desc = *url.Description
	}

	view := dto.NewShortenedUrlView(
		url.ShortSlug,
		url.LongUrl,
		url.ExpiresAt.Format("2006-01-02 15:04:05"),
		desc,
		url.PasswordHash != nil,
		url.ExpiresAt.Format("2006-01-02T15:04"),
	)

	pages.EditUrlRow(*view, "").Render(r.Context(), w)
}

// UpdateUrl handles POST /dashboard/urls/{slug}/edit
func (h *ShortenedUrl) UpdateUrl(w http.ResponseWriter, r *http.Request) {
	shortSlug := chi.URLParam(r, "slug")

	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updateDto dto.ShortenedUrlUpdate
	decoder := form.NewDecoder()
	if err := decoder.Decode(&updateDto, r.Form); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = updateDto.Validate()
	if err != nil {
		// Just re-render the edit form with the error message
		view := dto.NewShortenedUrlView(
			shortSlug,
			updateDto.OriginalUrl,
			"",
			updateDto.Description,
			updateDto.Password != "",
			updateDto.ExpiresAt,
		)
		pages.EditUrlRow(*view, err.Error()).Render(r.Context(), w)
		return
	}

	err = h.usecase.Update(shortSlug, updateDto, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch updated to render UrlRow again
	url, err := h.usecase.FindBySlug(shortSlug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	desc := ""
	if url.Description != nil {
		desc = *url.Description
	}

	updatedView := dto.NewShortenedUrlView(
		url.ShortSlug,
		url.LongUrl,
		url.ExpiresAt.Format("2006-01-02 15:04:05"),
		desc,
		url.PasswordHash != nil,
		url.ExpiresAt.Format("2006-01-02T15:04"),
	)

	pages.UrlRow(*updatedView).Render(r.Context(), w)
}

// GetUrlRow handles GET /dashboard/urls/{slug}
func (h *ShortenedUrl) GetUrlRow(w http.ResponseWriter, r *http.Request) {
	shortSlug := chi.URLParam(r, "slug")

	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	url, err := h.usecase.FindBySlug(shortSlug)
	if err != nil || url.UserId != userId {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	desc := ""
	if url.Description != nil {
		desc = *url.Description
	}

	view := dto.NewShortenedUrlView(
		url.ShortSlug,
		url.LongUrl,
		url.ExpiresAt.Format("2006-01-02 15:04:05"),
		desc,
		url.PasswordHash != nil,
		url.ExpiresAt.Format("2006-01-02T15:04"),
	)

	pages.UrlRow(*view).Render(r.Context(), w)
}

// GetUrlStatsPage handles GET /dashboard/urls/{slug}/stats
func (h *ShortenedUrl) GetUrlStatsPage(w http.ResponseWriter, r *http.Request) {
	shortSlug := chi.URLParam(r, "slug")

	userIdStr, ok := appctx.GetAuthUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	stats, err := h.usecase.GetUrlStats(r.Context(), shortSlug, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pages.StatsPage(shortSlug, stats).Render(r.Context(), w)
}
