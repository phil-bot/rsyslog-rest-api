package handlers

import (
	"net/http"

	"github.com/phil-bot/rsyslox/internal/models"
)

// RootHandler handles GET / — returns basic API info as JSON.
type RootHandler struct {
	version string
}

// NewRootHandler creates a new RootHandler.
func NewRootHandler(version string) *RootHandler {
	return &RootHandler{version: version}
}

// ServeHTTP handles the root endpoint.
func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only GET method is allowed"))
		return
	}
	respondJSON(w, http.StatusOK, models.RootResponse{
		Name:    "rsyslox",
		Version: h.version,
		Endpoints: map[string]string{
			"health": "/health",
			"logs":   "/api/logs",
			"meta":   "/api/meta",
		},
	})
}
