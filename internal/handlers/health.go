package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/phil-bot/rsyslox/internal/config"
	"github.com/phil-bot/rsyslox/internal/database"
)

// HealthResponse is the JSON body returned by GET /health.
// SetupMode=true signals that no config.toml exists yet — the frontend
// should redirect to the setup wizard.
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database,omitempty"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	SetupMode bool   `json:"setup_mode,omitempty"`
}

// HealthHandler handles GET /health.
type HealthHandler struct {
	db      *database.DB // nil while in setup mode
	version string
}

// NewHealthHandler creates a HealthHandler.
// db may be nil when the server starts without a config (setup mode).
func NewHealthHandler(db *database.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Check config file existence on every request.
	// This allows the frontend to detect when setup has completed without
	// a server restart: once config.toml is written the next /health call
	// returns setup_mode=false and the router navigates to /login.
	cfgPath := config.ActiveConfigPath()
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusOK)
		if encErr := json.NewEncoder(w).Encode(HealthResponse{
			Status:    "setup",
			Version:   h.version,
			Timestamp: time.Now().Format(time.RFC3339),
			SetupMode: true,
		}); encErr != nil {
			log.Printf("health: encode error: %v", encErr)
		}
		return
	}

	// Config exists — normal operation.
	// db may still be nil if the server hasn't restarted yet after setup;
	// report as "pending" so the frontend can show a useful message.
	if h.db == nil {
		w.WriteHeader(http.StatusOK)
		if encErr := json.NewEncoder(w).Encode(HealthResponse{
			Status:    "pending_restart",
			Version:   h.version,
			Timestamp: time.Now().Format(time.RFC3339),
		}); encErr != nil {
			log.Printf("health: encode error: %v", encErr)
		}
		return
	}

	if err := h.db.Health(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		if encErr := json.NewEncoder(w).Encode(HealthResponse{
			Status:    "unhealthy",
			Database:  "disconnected",
			Version:   h.version,
			Timestamp: time.Now().Format(time.RFC3339),
		}); encErr != nil {
			log.Printf("health: encode error: %v", encErr)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(HealthResponse{
		Status:    "healthy",
		Database:  "connected",
		Version:   h.version,
		Timestamp: time.Now().Format(time.RFC3339),
	}); encErr != nil {
		log.Printf("health: encode error: %v", encErr)
	}
}
