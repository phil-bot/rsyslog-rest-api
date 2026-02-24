package admin

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/phil-bot/rsyslox/internal/auth"
	"github.com/phil-bot/rsyslox/internal/models"
)

// LoginRequest is the payload for POST /api/admin/login.
type LoginRequest struct {
	Password string `json:"password"`
}

// LoginResponse is returned on successful authentication.
type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler handles admin login requests.
type LoginHandler struct {
	mgr   *auth.Manager
	store *auth.SessionStore
}

// NewLoginHandler creates a new LoginHandler.
func NewLoginHandler(mgr *auth.Manager, store *auth.SessionStore) *LoginHandler {
	return &LoginHandler{mgr: mgr, store: store}
}

// ServeHTTP handles POST /api/admin/login.
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only POST is allowed"))
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			models.NewAPIError(models.ErrCodeInvalidParameter, "Invalid JSON body"))
		return
	}

	if !h.mgr.VerifyAdminPassword(req.Password) {
		// Use constant-time-safe generic message to avoid user enumeration
		respondError(w, http.StatusUnauthorized,
			models.NewAPIError(models.ErrCodeUnauthorized, "Invalid credentials"))
		return
	}

	token, err := h.store.Create()
	if err != nil {
		log.Printf("Login: failed to create session: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to create session"))
		return
	}

	log.Println("Admin login successful")
	respondJSON(w, http.StatusOK, LoginResponse{Token: token})
}

// LogoutHandler handles POST /api/admin/logout.
type LogoutHandler struct {
	store *auth.SessionStore
}

// NewLogoutHandler creates a new LogoutHandler.
func NewLogoutHandler(store *auth.SessionStore) *LogoutHandler {
	return &LogoutHandler{store: store}
}

// ServeHTTP handles POST /api/admin/logout.
func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only POST is allowed"))
		return
	}

	token := extractToken(r)
	if token != "" {
		h.store.Revoke(token)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
}
