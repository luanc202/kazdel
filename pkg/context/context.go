package context

import (
	"context"
	"errors"
	"net/http"
)

// contextKey is a private type for context keys to prevent collisions.
type contextKey string

const (
	// authUserKey is the context key for the authenticated user's ID.
	authUserKey contextKey = "auth_user_id"
)

// SetAuthUser stores the authenticated user ID in the request context
// and returns the modified request.
func SetAuthUser(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), authUserKey, userID)
	return r.WithContext(ctx)
}

// GetAuthUser retrieves the authenticated user ID from the request context.
// Returns the user ID and a boolean indicating if it was found.
func GetAuthUser(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(authUserKey).(string)
	return v, ok
}

// IsCanceledError determines if an error is due to a context cancellation.
func IsCanceledError(err error) bool {
	return errors.Is(err, context.Canceled)
}
