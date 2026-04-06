package controllers

import (
	"kazdel/pkg/constants"
	"kazdel/pkg/entity/dto"
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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var decoder *form.Decoder

func (c *AuthController) Signup(w http.ResponseWriter, r *http.Request) {
	var dto dto.SignUpRequest
	decoder = form.NewDecoder()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	if err := decoder.Decode(&dto, r.Form); err != nil {
		http.Error(w, "Invalid request form", http.StatusBadRequest)
		return
	}

	if err := dto.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if token, err := c.AuthUseCase.Signup(dto.Name, dto.Username, dto.Email, dto.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		setSessionCookie(w, token, time.Now().Add(24*time.Hour))
	}

	w.WriteHeader(http.StatusCreated)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	decoder = form.NewDecoder()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	if err := decoder.Decode(&req, r.Form); err != nil {
		http.Error(w, "Invalid request form", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := c.AuthUseCase.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set HttpOnly cookie
	setSessionCookie(w, token, time.Now().Add(24*time.Hour))

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie
	cookie, err := r.Cookie(constants.SessionCookieName)
	if err == nil {
		c.AuthUseCase.Logout(cookie.Value)
	}

	// Clear the cookie
	setSessionCookie(w, "", time.Now().Add(-1*time.Hour))

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
