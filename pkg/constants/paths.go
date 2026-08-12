package constants

import "kazdel/pkg/infra/config"

// GetBasePath returns the base path, ensuring it doesn't end with a trailing slash
// unless it's just "/". (Though normally BASE_PATH is like "/url" or "")
func GetBasePath() string {
	return config.GetEnvConfig().BASE_PATH
}

// LoginPath returns the absolute path for the login page, accounting for BASE_PATH.
func LoginPath() string {
	return GetBasePath() + "/login"
}

// DashboardPath returns the absolute path for the dashboard page, accounting for BASE_PATH.
func DashboardPath() string {
	return GetBasePath() + "/dashboard"
}

// AdminUsersPath returns the absolute path for the admin users page, accounting for BASE_PATH.
func AdminUsersPath() string {
	return GetBasePath() + "/admin/users"
}
