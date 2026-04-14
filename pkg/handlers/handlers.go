package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
)

var registeredHandlers []Handler

// Handler handles one or more HTTP routes.
type Handler interface {
	// Routes allows for self-registration of HTTP routes on the router.
	Routes(r chi.Router)

	// Init provides dependencies to initialize the handler.
	Init(deps *Dependencies) error
}

// Dependencies holds all shared dependencies for handlers.
type Dependencies struct {
	ShortenedUrlUseCase *usecase.ShortenedUrlUsecase
	AuthUseCase         *usecase.AuthUseCase
}

// Register registers a handler for automatic initialization and routing.
func Register(h Handler) {
	registeredHandlers = append(registeredHandlers, h)
}

// GetHandlers returns all registered handlers.
func GetHandlers() []Handler {
	return registeredHandlers
}

// fail is a helper to fail a request by returning a 500 error and logging the error.
func fail(w http.ResponseWriter, err error, msg string) {
	slog.Error(fmt.Sprintf("%s: %v", msg, err))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
