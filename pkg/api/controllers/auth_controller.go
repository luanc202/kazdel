package controllers

import (
	"encoding/json"
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

	if err := c.AuthUseCase.Signup(dto.Name, dto.Username, dto.Email, dto.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := c.AuthUseCase.Login(req.Email, req.Password)
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

	w.WriteHeader(http.StatusOK)

}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie
	cookie, err := r.Cookie(constants.SessionCookieName)
	if err == nil {
		c.AuthUseCase.Logout(cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constants.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Past time to delete
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}
