package handlers

import (
	"net/http"
	"time"

	"kazdel/pkg/constants"
	"kazdel/pkg/usecase"
)

type AuthHandler struct {
	AuthUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		AuthUseCase: authUseCase,
	}
}

// ShowLoginPage renders the login page
func (h *AuthHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	// TODO: Render templ login page
	w.Write([]byte("Login Page"))
}

// ShowSignupPage renders the signup page
func (h *AuthHandler) ShowSignupPage(w http.ResponseWriter, r *http.Request) {
	// TODO: Render templ signup page
	w.Write([]byte("Signup Page"))
}

// HandleLogin processes login form
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	token, err := h.AuthUseCase.Login(email, password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constants.SessionCookieName,
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// HTMX redirect to dashboard or home
	w.Header().Set("HX-Redirect", "/")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleSignup processes signup form
func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	err = h.AuthUseCase.Signup(name, email, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect to login page on success
	w.Header().Set("HX-Redirect", "/login")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleLogout processes logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(constants.SessionCookieName)
	if err == nil {
		h.AuthUseCase.Logout(cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constants.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.Header().Set("HX-Redirect", "/")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
