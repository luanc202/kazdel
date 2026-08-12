package middleware

import (
	"net/http"

	appctx "kazdel/pkg/context"
	"kazdel/pkg/entity"
	"kazdel/pkg/infra/config"
	interfaces "kazdel/pkg/interface"
)

// RequireAdminMiddleware checks if the authenticated user has RoleAdmin.
// It assumes AuthMiddleware has already run and placed the userID in the context.
func RequireAdminMiddleware(userRepo interfaces.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			loginPath := config.GetEnvConfig().BASE_PATH + "/login"
			dashboardPath := config.GetEnvConfig().BASE_PATH + "/dashboard"

			userID, ok := appctx.GetAuthUser(r)
			if !ok || userID == "" {
				// Not authenticated
				http.Redirect(w, r, loginPath, http.StatusSeeOther)
				return
			}

			user, err := userRepo.FindById(userID)
			if err != nil {
				// User not found or db error
				http.Redirect(w, r, loginPath, http.StatusSeeOther)
				return
			}

			if user.Role != entity.RoleAdmin {
				// User is not an admin
				// Redirect to dashboard
				http.Redirect(w, r, dashboardPath, http.StatusSeeOther)
				return
			}

			// User is an admin, proceed
			next.ServeHTTP(w, r)
		})
	}
}
