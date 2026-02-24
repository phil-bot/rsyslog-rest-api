package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/phil-bot/rsyslox/internal/auth"
	"github.com/phil-bot/rsyslox/internal/models"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const roleKey contextKey = "auth_role"

// AuthReadOnly returns a middleware that accepts both admin session tokens
// and read-only API keys. It rejects unauthenticated requests.
func AuthReadOnly(mgr *auth.Manager, store *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := resolveRole(r, mgr, store)
			if role == auth.RoleNone {
				respondError(w, http.StatusUnauthorized, models.NewAPIError(
					models.ErrCodeUnauthorized,
					"Authentication required").
					WithDetails("Provide X-API-Key header or X-Session-Token header"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthAdmin returns a middleware that only accepts admin session tokens.
// Read-only keys are rejected.
func AuthAdmin(store *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" || !store.Validate(token) {
				respondError(w, http.StatusUnauthorized, models.NewAPIError(
					models.ErrCodeUnauthorized,
					"Admin authentication required").
					WithDetails("Provide a valid X-Session-Token header"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// LocalhostOnly returns a middleware that restricts access to localhost.
// Used to protect /api/setup in normal (post-config) mode.
// In setup mode (no config yet) the server registers /api/setup without
// this middleware so headless servers can be configured remotely.
func LocalhostOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isLocalhost(r) {
				respondError(w, http.StatusForbidden, models.NewAPIError(
					"FORBIDDEN",
					"This endpoint is only accessible from localhost"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// resolveRole determines the auth.Role for a request.
func resolveRole(r *http.Request, mgr *auth.Manager, store *auth.SessionStore) auth.Role {
	if token := extractToken(r); token != "" && store.Validate(token) {
		return auth.RoleAdmin
	}
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" && mgr.VerifyReadOnlyKey(apiKey) != "" {
		return auth.RoleReadOnly
	}
	return auth.RoleNone
}

// extractToken extracts the session token from the request headers.
func extractToken(r *http.Request) string {
	if t := r.Header.Get("X-Session-Token"); t != "" {
		return t
	}
	if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
		return strings.TrimPrefix(a, "Bearer ")
	}
	return ""
}

// isLocalhost reports whether the request originates from localhost.
func isLocalhost(r *http.Request) bool {
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	host = strings.Trim(host, "[]")
	return host == "127.0.0.1" || host == "::1" || host == "localhost"
}

// respondError writes a JSON error response. Defined locally to avoid
// import cycles with the handlers package.
func respondError(w http.ResponseWriter, status int, err *models.APIError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(err); encErr != nil {
		log.Printf("middleware: failed to encode error response: %v", encErr)
	}
}
