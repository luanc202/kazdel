package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"url-shortener/m/api/errors"
	authMiddleware "url-shortener/m/api/middleware"
	"url-shortener/m/entity/dto"
	"url-shortener/m/internal/uniqueEntityId"
	"url-shortener/m/usecase"

	"github.com/go-chi/chi/v5"
)

type ShortenedUrlController struct {
	Usecase *usecase.ShortenedUrlUsecase
}

func NewShortenedUrlController(usecase *usecase.ShortenedUrlUsecase) *ShortenedUrlController {
	return &ShortenedUrlController{
		Usecase: usecase,
	}
}

func (cntrl *ShortenedUrlController) FindShortenedUrl(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	shortenedUrl, err := cntrl.Usecase.FindBySlug(slug)
	if err != nil {
		http.Error(w, "Shortened URL not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(dto.NewShortenedUrlView(shortenedUrl.ShortSlug, shortenedUrl.LongUrl, shortenedUrl.ExpiresAt.Format("2006-01-02 15:04:05"))); err != nil {
		http.Error(w, "Failed to encode shortened url", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cntrl *ShortenedUrlController) RedirectToLongUrl(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	shortenedUrl, err := cntrl.Usecase.FindBySlug(slug)
	if err != nil {
		http.Error(w, "Shortened URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, shortenedUrl.LongUrl, http.StatusFound)

}

func (cntrl *ShortenedUrlController) CreateShortenedUrl(w http.ResponseWriter, r *http.Request) {
	var shortenedUrlInsert dto.ShortenedUrlInsert
	// user ID should be extracted from the token
	// no authentication for now
	err := json.NewDecoder(r.Body).Decode(&shortenedUrlInsert)
	defer r.Body.Close()

	if err != nil {
		fmt.Printf("Invalid request: could not decode shortened url data from request body %s", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errors.ErrInvalidBody{
			Description: "The body is invalid",
		})
		return
	}

	err = shortenedUrlInsert.Validate()
	if err != nil {
		fmt.Printf("Invalid request: could not validate shortened url data from request body %s", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errors.ErrInvalidBody{
			Description: err.Error(),
		})
		return
	}

	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		fmt.Printf("Invalid user ID: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = cntrl.Usecase.Save(shortenedUrlInsert, userId)

	if err != nil {
		fmt.Printf("Error in usecase: %s", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errors.ErrInvalidBody{
			Description: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (cntrl *ShortenedUrlController) ListUserUrls(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)

	if err != nil {
		fmt.Printf("Invalid user ID: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := cntrl.Usecase.ListByUser(userId)
	if err != nil {
		http.Error(w, "Failed to list URLs", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		http.Error(w, "No URLs found", http.StatusOK)
		return
	}

	var response []dto.ShortenedUrlView
	for _, u := range urls {
		response = append(response, *dto.NewShortenedUrlView(u.ShortSlug, u.LongUrl, u.ExpiresAt.Format("2006-01-02 15:04:05")))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
