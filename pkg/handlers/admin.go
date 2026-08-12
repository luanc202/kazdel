package handlers

import (
	"math"
	"net/http"
	"strconv"

	interfaces "kazdel/pkg/interface"
	customMiddleware "kazdel/pkg/middleware"
	"kazdel/pkg/infra/config"
	"kazdel/pkg/ui/pages"
	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	adminUseCase *usecase.AdminUseCase
	authUseCase  *usecase.AuthUseCase
	userRepo     interfaces.UserRepository
}

func init() {
	Register(&AdminHandler{})
}

func (h *AdminHandler) Init(deps *Dependencies) error {
	h.adminUseCase = deps.AdminUseCase
	h.authUseCase = deps.AuthUseCase
	h.userRepo = deps.UserRepo
	return nil
}

func (h *AdminHandler) Routes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(h.authUseCase))
		r.Use(customMiddleware.RequireAdminMiddleware(h.userRepo))

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			adminUsersPath := config.GetEnvConfig().BASE_PATH + "/admin/users"
			http.Redirect(w, req, adminUsersPath, http.StatusSeeOther)
		})

		r.Get("/users", h.handleGetUsers)
		r.Post("/users/{id}/status", h.handleUpdateUserStatus)

		r.Get("/urls", h.handleGetURLs)
		r.Delete("/urls/{slug}", h.handleDeleteURL)

		r.Get("/settings", h.handleGetSettings)
		r.Post("/settings", h.handleUpdateSetting)
	})
}

func (h *AdminHandler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	users, total, err := h.adminUseCase.GetUsers(search, page, limit)
	if err != nil {
		fail(w, err, "failed to get users")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	pages.AdminUsers(users, search, page, limit, totalPages).Render(r.Context(), w)
}

func (h *AdminHandler) handleUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		fail(w, err, "failed to parse form")
		return
	}

	isActive := r.FormValue("isActive") == "true"

	err := h.adminUseCase.UpdateUserStatus(id, isActive)
	if err != nil {
		fail(w, err, "failed to update user status")
		return
	}

	// Re-render the user row or just return success and let HX refresh?
	// The simplest way to return the updated user row is to fetch it.
	user, err := h.userRepo.FindById(id)
	if err != nil {
		fail(w, err, "failed to fetch updated user")
		return
	}

	pages.AdminUserRow(user).Render(r.Context(), w)
}

func (h *AdminHandler) handleGetURLs(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	urls, total, err := h.adminUseCase.GetURLs(search, page, limit)
	if err != nil {
		fail(w, err, "failed to get urls")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	pages.AdminURLs(urls, search, page, limit, totalPages).Render(r.Context(), w)
}

func (h *AdminHandler) handleDeleteURL(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	err := h.adminUseCase.DeleteURL(slug)
	if err != nil {
		fail(w, err, "failed to delete url")
		return
	}

	// HTMX will swap outerHTML with empty content if we just return 200 OK.
	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.adminUseCase.GetSettings()
	if err != nil {
		fail(w, err, "failed to get settings")
		return
	}

	pages.AdminSettings(settings).Render(r.Context(), w)
}

func (h *AdminHandler) handleUpdateSetting(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fail(w, err, "failed to parse form")
		return
	}

	key := r.FormValue("key")
	value := r.FormValue("value")

	err := h.adminUseCase.UpdateSetting(key, value)
	if err != nil {
		fail(w, err, "failed to update setting")
		return
	}

	// Return the updated setting row
	setting, err := h.adminUseCase.GetSettings() // Ideally get just this setting, but for now we iterate to find it.
	for _, s := range setting {
		if s.Key == key {
			pages.AdminSettingRow(s).Render(r.Context(), w)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
