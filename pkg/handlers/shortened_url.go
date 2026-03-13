package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	authMiddleware "kazdel/pkg/api/middleware"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/uniqueEntityId"
	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
)

type ShortenedUrlHandler struct {
	Usecase *usecase.ShortenedUrlUsecase
}

func NewShortenedUrlHandler(usecase *usecase.ShortenedUrlUsecase) *ShortenedUrlHandler {
	return &ShortenedUrlHandler{
		Usecase: usecase,
	}
}

// ShowDashboard renders the dashboard page with user's URLs
func (h *ShortenedUrlHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	urls, err := h.Usecase.ListByUser(userId)
	if err != nil {
		http.Error(w, "Failed to list URLs", http.StatusInternalServerError)
		return
	}

	// TODO: Render templ dashboard page passing urls
	w.Write([]byte(fmt.Sprintf("Dashboard Page - You have %d URLs", len(urls))))
}

// HandleCreateUrl processes create URL form via HTMX
func (h *ShortenedUrlHandler) HandleCreateUrl(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	originalUrl := r.FormValue("original_url")
	expiresAtStr := r.FormValue("expires_at") // Expected format: 2006-01-02T15:04

	var expiresAt time.Time
	if expiresAtStr != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04", expiresAtStr)
		if err == nil {
			expiresAt = parsedTime
		}
	} else {
		// Provide a default expiration of 7 days if not set
		expiresAt = time.Now().AddDate(0, 0, 7)
	}

	insertDto := dto.ShortenedUrlInsert{
		OriginalUrl: originalUrl,
		ExpiresAt:   expiresAt,
	}

	if err := insertDto.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.Usecase.Save(insertDto, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Tell HTMX to refresh the page to show the new URL in the list
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusCreated)
}

// HandleDeleteUrl processes URL deletion
func (h *ShortenedUrlHandler) HandleDeleteUrl(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userIdStr := r.Context().Value(authMiddleware.UserIDKey).(string)
	userId, err := uniqueEntityId.ParseID(userIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.Usecase.Delete(id, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return empty 200 OK for HTMX to remove the row from the table
	w.WriteHeader(http.StatusOK)
}
