package middleware

import (
	"net/http"

	"kazdel/pkg/constants"
	appctx "kazdel/pkg/context"
	"kazdel/pkg/usecase"
)

// AuthMiddleware validates the session token from cookies or Authorization header,
// extracts the authenticated user ID, and stores it in the request context.
func AuthMiddleware(authUseCase *usecase.AuthUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Try to get token from cookie
			cookie, err := r.Cookie(constants.SessionCookieName)
			var tokenString string

			if err == nil {
				tokenString = cookie.Value
			}

			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := authUseCase.ValidateSession(tokenString)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			r = appctx.SetAuthUser(r, userID)
			next.ServeHTTP(w, r)
		})
	}
}
