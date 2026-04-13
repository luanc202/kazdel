package controllers

import (
	"kazdel/pkg/constants"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/ui/pages"
	"kazdel/pkg/usecase"
	"net/http"
	"time"

	"github.com/go-playground/form"
)

type AuthController struct {
	AuthUseCase *usecase.AuthUseCase
}

func NewAuthController(authUseCase *usecase.AuthUseCase) *AuthController {
	return &AuthController{
		AuthUseCase: authUseCase,
	}
}

var decoder *form.Decoder

func (c *AuthController) Signup(w http.ResponseWriter, r *http.Request) {
	var dto dto.SignUpRequest
	decoder = form.NewDecoder()

	if err := r.ParseForm(); err != nil {
		pages.SignUpForm("Failed to parse form data", "", "", "").Render(r.Context(), w)
		return
	}

	if err := decoder.Decode(&dto, r.Form); err != nil {
		pages.SignUpForm("Invalid request form", dto.Name, dto.Username, dto.Email).Render(r.Context(), w)
		return
	}

	if err := dto.Validate(); err != nil {
		pages.SignUpForm(err.Error(), dto.Name, dto.Username, dto.Email).Render(r.Context(), w)
		return
	}

	if token, err := c.AuthUseCase.Signup(dto.Name, dto.Username, dto.Email, dto.Password); err != nil {
		pages.SignUpForm(err.Error(), dto.Name, dto.Username, dto.Email).Render(r.Context(), w)
		return
	} else {
		setSessionCookie(w, token, time.Now().Add(24*time.Hour))
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	decoder = form.NewDecoder()

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

	token, err := c.AuthUseCase.Login(req.Username, req.Password)
	if err != nil {
		pages.LoginForm("Invalid credentials", req.Username).Render(r.Context(), w)
		return
	}

	// Set HttpOnly cookie
	setSessionCookie(w, token, time.Now().Add(24*time.Hour))

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie
	cookie, err := r.Cookie(constants.SessionCookieName)
	if err == nil {
		c.AuthUseCase.Logout(cookie.Value)
	}

	// Clear the cookie
	setSessionCookie(w, "", time.Now().Add(-1*time.Hour))

	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}

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
