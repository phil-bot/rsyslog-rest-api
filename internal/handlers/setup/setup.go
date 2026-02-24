// Package setup provides the first-run setup wizard API handler.
// The setup endpoint is only reachable from localhost and only when no
// configuration file exists yet.
package setup

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/phil-bot/rsyslox/internal/auth"
	"github.com/phil-bot/rsyslox/internal/config"
	"github.com/phil-bot/rsyslox/internal/models"
)

// SetupRequest is the payload sent by the setup wizard on first run.
type SetupRequest struct {
	// Database connection
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`

	// Admin credentials
	AdminPassword string `json:"admin_password"`

	// Server settings
	ServerHost string `json:"server_host"`
	ServerPort int    `json:"server_port"`
	UseSSL     bool   `json:"use_ssl"`
}

// SetupResponse is returned after a successful setup.
type SetupResponse struct {
	Message string `json:"message"`
}

// Handler handles the POST /api/setup endpoint.
type Handler struct {
	cfg          *config.Config
	sessionStore *auth.SessionStore
}

// New creates a new setup Handler.
func New(cfg *config.Config, store *auth.SessionStore) *Handler {
	return &Handler{cfg: cfg, sessionStore: store}
}

// PrefillResponse contains database defaults for the setup wizard.
// Values are read from RSYSLOX_PREFILL_* environment variables,
// which the Docker entrypoint sets so the operator does not have to
// type the credentials manually.
type PrefillResponse struct {
	DBHost       string `json:"db_host"`
	DBPort       int    `json:"db_port"`
	DBName       string `json:"db_name"`
	DBUser       string `json:"db_user"`
	ServerHost   string `json:"server_host"`
	ServerPort   int    `json:"server_port"`
}

// ServeHTTP processes GET (prefill) and POST (submit) requests for the setup wizard.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handlePrefill(w, r)
	case http.MethodPost:
		h.handleSubmit(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only GET and POST are allowed"))
	}
}

// handlePrefill returns database defaults from environment variables.
func (h *Handler) handlePrefill(w http.ResponseWriter, r *http.Request) {
	port := 3306
	if p := os.Getenv("RSYSLOX_PREFILL_DB_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}
	srvPort := h.cfg.Server.Port
	if srvPort == 0 {
		srvPort = 8000
	}

	respondJSON(w, http.StatusOK, PrefillResponse{
		DBHost:     getEnv("RSYSLOX_PREFILL_DB_HOST", "localhost"),
		DBPort:     port,
		DBName:     getEnv("RSYSLOX_PREFILL_DB_NAME", "Syslog"),
		DBUser:     getEnv("RSYSLOX_PREFILL_DB_USER", ""),
		ServerHost: h.cfg.Server.Host,
		ServerPort: srvPort,
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// handleSubmit processes the setup form submission.
func (h *Handler) handleSubmit(w http.ResponseWriter, r *http.Request) {

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			models.NewAPIError(models.ErrCodeInvalidParameter, "Invalid JSON body"))
		return
	}

	if err := validateSetupRequest(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	// Hash admin password
	hash, err := auth.HashAdminPassword(req.AdminPassword)
	if err != nil {
		log.Printf("Setup: failed to hash password: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to hash password"))
		return
	}

	// Encrypt database password
	encPass, err := config.EncryptPassword(req.DBPassword)
	if err != nil {
		log.Printf("Setup: failed to encrypt DB password: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to encrypt database password"))
		return
	}

	// Apply setup values to config
	h.cfg.Database.Host = req.DBHost
	h.cfg.Database.Port = req.DBPort
	h.cfg.Database.Name = req.DBName
	h.cfg.Database.User = req.DBUser
	h.cfg.Database.Password = encPass
	h.cfg.Auth.AdminPasswordHash = hash

	if req.ServerHost != "" {
		h.cfg.Server.Host = req.ServerHost
	}
	if req.ServerPort > 0 {
		h.cfg.Server.Port = req.ServerPort
	}
	h.cfg.Server.UseSSL = req.UseSSL

	// Persist configuration
	if err := config.Save(h.cfg); err != nil {
		log.Printf("Setup: failed to save config: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to save configuration: "+err.Error()))
		return
	}

	log.Println("✓ Setup completed, configuration saved to", h.cfg.ConfigPath)

	respondJSON(w, http.StatusOK, SetupResponse{
		Message: "Setup complete. Restarting…",
	})

	// Restart the process so it picks up the new config and registers all routes.
	// syscall.Exec replaces the current process image — the Docker container /
	// systemd unit keeps the same PID and stays alive.
	go func() {
		time.Sleep(500 * time.Millisecond)
		exe, err := os.Executable()
		if err != nil {
			log.Printf("Setup: failed to determine executable path: %v — please restart manually", err)
			return
		}
		log.Println("Restarting rsyslox with new configuration…")
		if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
			log.Printf("Setup: exec restart failed: %v — please restart manually", err)
		}
	}()
}

func validateSetupRequest(req *SetupRequest) *models.APIError {
	if req.DBHost == "" {
		return models.NewValidationError("db_host", "Database host is required")
	}
	if req.DBName == "" {
		return models.NewValidationError("db_name", "Database name is required")
	}
	if req.DBUser == "" {
		return models.NewValidationError("db_user", "Database user is required")
	}
	if req.DBPassword == "" {
		return models.NewValidationError("db_password", "Database password is required")
	}
	if len(req.AdminPassword) < 12 {
		return models.NewValidationError("admin_password", "Admin password must be at least 12 characters")
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, err *models.APIError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(err)
}
