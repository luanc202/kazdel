package handlers

import (
	"net/http"
	"time"

	"kazdel/pkg/constants"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/middleware"
	"kazdel/pkg/ui/pages"
	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/form"
)

// Auth handles authentication-related HTTP routes.
type Auth struct {
	authUseCase *usecase.AuthUseCase
}

func init() {
	Register(new(Auth))
}

func (h *Auth) Init(deps *Dependencies) error {
	h.authUseCase = deps.AuthUseCase
	return nil
}

func (h *Auth) Routes(r chi.Router) {
	// Public auth API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Group(func(auth chi.Router) {
				auth.Use(chiMiddleware.AllowContentType("application/x-www-form-urlencoded"))
				auth.Post("/auth/signup", h.SignupSubmit)
				auth.Post("/auth/login", h.LoginSubmit)
			})

			// Protected auth API routes
			r.Group(func(private chi.Router) {
				private.Use(chiMiddleware.AllowContentType("application/x-www-form-urlencoded"))
				private.Use(middleware.AuthMiddleware(h.authUseCase))
				// We allow standard GET for logout matching the anchor tag on dashboard
				private.Get("/auth/logout", h.Logout)
				private.Post("/auth/logout", h.Logout)
			})
		})
	})
}

// SignupSubmit handles POST /api/v1/auth/signup
// @Summary Register a new user
// @Accept application/x-www-form-urlencoded
// @Produce text/html
// @Param name formData string true "User's display name"
// @Param username formData string true "Username"
// @Param email formData string true "Email address"
// @Param password formData string true "Password"
// @Success 200 "Redirect to dashboard via HX-Redirect"
// @Failure 400 "Validation error or duplicate user"
// @Router /api/v1/auth/signup [post]
func (h *Auth) SignupSubmit(w http.ResponseWriter, r *http.Request) {
	var signupReq dto.SignUpRequest
	decoder := form.NewDecoder()

	if err := r.ParseForm(); err != nil {
		pages.SignUpForm("Failed to parse form data", "", "", "").Render(r.Context(), w)
		return
	}

	if err := decoder.Decode(&signupReq, r.Form); err != nil {
		pages.SignUpForm("Invalid request form", signupReq.Name, signupReq.Username, signupReq.Email).Render(r.Context(), w)
		return
	}

	if err := signupReq.Validate(); err != nil {
		pages.SignUpForm(err.Error(), signupReq.Name, signupReq.Username, signupReq.Email).Render(r.Context(), w)
		return
	}

	token, err := h.authUseCase.Signup(signupReq.Name, signupReq.Username, signupReq.Email, signupReq.Password)
	if err != nil {
		pages.SignUpForm(err.Error(), signupReq.Name, signupReq.Username, signupReq.Email).Render(r.Context(), w)
		return
	}

	setSessionCookie(w, token, time.Now().Add(24*time.Hour))

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// LoginSubmit handles POST /api/v1/auth/login
// @Summary Authenticate a user
// @Accept application/x-www-form-urlencoded
// @Produce text/html
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 "Redirect to dashboard via HX-Redirect"
// @Failure 400 "Validation error"
// @Failure 401 "Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *Auth) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	decoder := form.NewDecoder()

	if err := r.ParseForm(); err != nil {
		pages.LoginForm("Failed to parse form data", "").Render(r.Context(), w)
		return
	}

	if err := decoder.Decode(&req, r.Form); err != nil {
		pages.LoginForm("Invalid request form", req.Username).Render(r.Context(), w)
		return
	}

	if err := req.Validate(); err != nil {
		pages.LoginForm(err.Error(), req.Username).Render(r.Context(), w)
		return
	}

	token, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		pages.LoginForm("Invalid credentials", req.Username).Render(r.Context(), w)
		return
	}

	// Set HttpOnly cookie
	setSessionCookie(w, token, time.Now().Add(24*time.Hour))

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// Logout handles GET/POST /api/v1/auth/logout
// @Summary Log out the current user
// @Success 200 "Redirect to login via HX-Redirect"
// @Router /api/v1/auth/logout [get]
func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie
	cookie, err := r.Cookie(constants.SessionCookieName)
	if err == nil {
		h.authUseCase.Logout(cookie.Value)
	}

	// Clear the cookie
	setSessionCookie(w, "", time.Now().Add(-1*time.Hour))

	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}

// setSessionCookie sets or clears the session cookie.
func setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     constants.SessionCookieName,
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}
