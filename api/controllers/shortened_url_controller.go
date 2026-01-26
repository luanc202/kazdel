package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"url-shortener/m/api/errors"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(dto.NewShortenedUrlView(shortenedUrl.ShortSlug, shortenedUrl.OriginalUrl)); err != nil {
		http.Error(w, "Failed to encode shortened url", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cntrl *ShortenedUrlController) CreateShortenedUrl(w http.ResponseWriter, r *http.Request) {
	var shortenedUrlInsert dto.ShortenedUrlInsert
	// user ID should be extracted from the token
	// no authentication for now
	var userID uniqueEntityId.ID
	userID, _ = uniqueEntityId.ParseID("1")

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
		json.NewEncoder(w).Encode(err)
		return
	}

	err = cntrl.Usecase.Save(shortenedUrlInsert, userID)

	if err != nil {
		fmt.Printf("Error in usecase: %s", err.Error())

		err := err.Error()

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
