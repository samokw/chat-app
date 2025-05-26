package user

import (
	"encoding/json"
	"log"
	"net/http"
	"server/internal/utils"
)

type Handler struct {
	Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service: s,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid payload", err)
		return
	}
	user, err := h.Service.CreateUser(ctx, &req)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "could not create user", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	log.Println("user created")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid payload", err)
		return
	}

	// Login and get user response with tokens
	user, err := h.Service.Login(ctx, &request)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, "invalid credentials", err)
		return
	}

	// Set HTTP-only cookies
	h.setTokenCookies(w, user.AccessToken, user.RefreshToken)

	response := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"message":  "login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.Println("user signed in")
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get refresh token from cookie
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, "refresh token not found", err)
		return
	}

	req := &RefreshTokenRequest{
		RefreshToken: refreshCookie.Value,
	}

	response, err := h.Service.RefreshToken(ctx, req)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, "failed to refresh token", err)
		return
	}

	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    response.AccessToken,
		Path:     "/",
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, accessCookie)

	responseBody := map[string]string{
		"message": response.Message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
	log.Println("token refreshed")
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "refresh token not found", err)
		return
	}

	req := &LogoutRequest{
		RefreshToken: refreshCookie.Value,
	}

	err = h.Service.Logout(ctx, req)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to logout", err)
		return
	}

	h.clearTokenCookies(w)

	w.WriteHeader(http.StatusOK)
	responseBody := map[string]string{
		"message": "logged out successfully",
	}
	json.NewEncoder(w).Encode(responseBody)
	log.Println("user logged out")
}

func (h *Handler) setTokenCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
}

func (h *Handler) clearTokenCookies(w http.ResponseWriter) {
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
}
