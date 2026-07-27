package handlers

import (
	"net/http"
	"time"

	"kazdel/pkg/constants"
	appctx "kazdel/pkg/context"
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
	// Public UI routes
	r.Get("/signup", h.SignupPage)
	r.Get("/login", h.LoginPage)
	r.Get("/forgot-password", h.ForgotPasswordPage)
	r.Get("/reset-password", h.ResetPasswordPage)
	r.Get("/verify-email", h.VerifyEmailPage)

	// Public auth API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Group(func(auth chi.Router) {
				auth.Use(chiMiddleware.AllowContentType("application/x-www-form-urlencoded"))
				auth.Post("/auth/signup", h.SignupSubmit)
				auth.Post("/auth/login", h.LoginSubmit)
				auth.Get("/auth/verify-email", h.VerifyEmail)
				auth.Post("/auth/resend-verification", h.ResendVerificationSubmit)
				auth.Post("/auth/forgot-password", h.ForgotPasswordSubmit)
				auth.Post("/auth/reset-password", h.ResetPasswordSubmit)
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
		pages.LoginForm("Failed to parse form data", "", "").Render(r.Context(), w)
		return
	}

	if err := decoder.Decode(&req, r.Form); err != nil {
		pages.LoginForm("Invalid request form", req.Username, "").Render(r.Context(), w)
		return
	}

	if err := req.Validate(); err != nil {
		pages.LoginForm(err.Error(), req.Username, "").Render(r.Context(), w)
		return
	}

	token, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		if err == usecase.ErrEmailNotVerified {
			email := req.Username
			if u, err := h.authUseCase.GetUserByUsername(req.Username); err == nil && u != nil {
				email = u.Email
			}
			pages.LoginForm("Your email is not verified.", req.Username, email).Render(r.Context(), w)
			return
		}
		pages.LoginForm("Invalid credentials", req.Username, "").Render(r.Context(), w)
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

// VerifyEmail handles GET /api/v1/auth/verify-email
func (h *Auth) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing token"))
		return
	}

	err := h.authUseCase.VerifyEmail(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid or expired token"))
		return
	}

	// Assuming a simple success message or redirect
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email successfully verified! You can now login."))
}

// VerifyEmailPage handles GET /verify-email
func (h *Auth) VerifyEmailPage(w http.ResponseWriter, r *http.Request) {
	_, ok := appctx.GetAuthUser(r)
	if ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	email := r.URL.Query().Get("email")
	pages.VerifyEmailPrompt(email, "", "").Render(r.Context(), w)
}

// ResendVerificationSubmit handles POST /api/v1/auth/resend-verification
func (h *Auth) ResendVerificationSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		pages.VerifyEmailPromptForm(email, "Email is required", "").Render(r.Context(), w)
		return
	}

	_ = h.authUseCase.ResendVerificationEmail(email)

	pages.VerifyEmailPromptForm(email, "", EMAIL_VERIFICATION_RESEND).Render(r.Context(), w)
}

// ForgotPasswordSubmit handles POST /api/v1/auth/forgot-password
func (h *Auth) ForgotPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		pages.ForgotPasswordForm("Email is required").Render(r.Context(), w)
		return
	}

	// We don't wait or handle errors here to prevent email enumeration attacks
	go func() {
		_ = h.authUseCase.RequestPasswordReset(email)
	}()

	pages.ForgotPasswordSuccess().Render(r.Context(), w)
}

// ResetPasswordSubmit handles POST /api/v1/auth/reset-password
func (h *Auth) ResetPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if token == "" || password == "" {
		pages.ResetPasswordForm(token, "Password is required").Render(r.Context(), w)
		return
	}

	if password != confirmPassword {
		pages.ResetPasswordForm(token, "Passwords do not match").Render(r.Context(), w)
		return
	}

	err := h.authUseCase.ResetPassword(token, password)
	if err != nil {
		pages.ResetPasswordForm(token, err.Error()).Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}

// SignupPage handles GET /signup
func (h *Auth) SignupPage(w http.ResponseWriter, r *http.Request) {
	_, ok := appctx.GetAuthUser(r)
	if ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	pages.SignUp().Render(r.Context(), w)
}

// LoginPage handles GET /login
func (h *Auth) LoginPage(w http.ResponseWriter, r *http.Request) {
	_, ok := appctx.GetAuthUser(r)
	if ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	pages.Login().Render(r.Context(), w)
}

// ForgotPasswordPage handles GET /forgot-password
func (h *Auth) ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	_, ok := appctx.GetAuthUser(r)
	if ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	pages.ForgotPassword().Render(r.Context(), w)
}

// ResetPasswordPage handles GET /reset-password
func (h *Auth) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	_, ok := appctx.GetAuthUser(r)
	if ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pages.ResetPassword(token).Render(r.Context(), w)
}
