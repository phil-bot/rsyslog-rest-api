package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/phil-bot/rsyslox/internal/auth"
	"github.com/phil-bot/rsyslox/internal/config"
	"github.com/phil-bot/rsyslox/internal/models"
)

// KeysHandler handles /api/admin/keys endpoints.
type KeysHandler struct {
	cfg *config.Config
}

// NewKeysHandler creates a new KeysHandler.
func NewKeysHandler(cfg *config.Config) *KeysHandler {
	return &KeysHandler{cfg: cfg}
}

// KeyResponse is returned for each stored read-only key.
// The actual key hash is never exposed; only the name.
type KeyResponse struct {
	Name string `json:"name"`
}

// CreateKeyResponse includes the plaintext key, shown exactly once.
type CreateKeyResponse struct {
	Name      string `json:"name"`
	Key       string `json:"key"` // plaintext – shown once, never stored
	Message   string `json:"message"`
}

// CreateKeyRequest is the payload for POST /api/admin/keys.
type CreateKeyRequest struct {
	Name string `json:"name"`
}

// ServeHTTP routes based on method and path suffix.
func (h *KeysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// DELETE /api/admin/keys/{name}
	if r.Method == http.MethodDelete {
		name := strings.TrimPrefix(r.URL.Path, "/api/admin/keys/")
		name = strings.TrimSpace(name)
		if name == "" {
			respondError(w, http.StatusBadRequest,
				models.NewAPIError(models.ErrCodeInvalidParameter, "Key name required in path"))
			return
		}
		h.handleDelete(w, name)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleList(w)
	case http.MethodPost:
		h.handleCreate(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Allowed: GET, POST, DELETE"))
	}
}

func (h *KeysHandler) handleList(w http.ResponseWriter) {
	keys := make([]KeyResponse, len(h.cfg.Auth.ReadOnlyKeys))
	for i, k := range h.cfg.Auth.ReadOnlyKeys {
		keys[i] = KeyResponse{Name: k.Name}
	}
	respondJSON(w, http.StatusOK, keys)
}

func (h *KeysHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			models.NewAPIError(models.ErrCodeInvalidParameter, "Invalid JSON body"))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondError(w, http.StatusBadRequest,
			models.NewValidationError("name", "Key name is required"))
		return
	}

	// Check for duplicate names
	for _, k := range h.cfg.Auth.ReadOnlyKeys {
		if k.Name == req.Name {
			respondError(w, http.StatusConflict,
				models.NewAPIError("CONFLICT", "A key with this name already exists"))
			return
		}
	}

	plaintext, hash, err := auth.GenerateReadOnlyKey()
	if err != nil {
		log.Printf("Keys: failed to generate key: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to generate key"))
		return
	}

	h.cfg.Auth.ReadOnlyKeys = append(h.cfg.Auth.ReadOnlyKeys, config.ReadOnlyKey{
		Name:    req.Name,
		KeyHash: hash,
	})

	if err := config.Save(h.cfg); err != nil {
		log.Printf("Keys: failed to save config: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to save configuration"))
		return
	}

	log.Printf("Admin: created read-only key %q", req.Name)
	respondJSON(w, http.StatusCreated, CreateKeyResponse{
		Name:    req.Name,
		Key:     plaintext,
		Message: "Store this key securely — it will not be shown again.",
	})
}

func (h *KeysHandler) handleDelete(w http.ResponseWriter, name string) {
	found := false
	filtered := h.cfg.Auth.ReadOnlyKeys[:0]
	for _, k := range h.cfg.Auth.ReadOnlyKeys {
		if k.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, k)
	}

	if !found {
		respondError(w, http.StatusNotFound,
			models.NewAPIError(models.ErrCodeNotFound, "Key not found: "+name))
		return
	}

	h.cfg.Auth.ReadOnlyKeys = filtered

	if err := config.Save(h.cfg); err != nil {
		log.Printf("Keys: failed to save config: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to save configuration"))
		return
	}

	log.Printf("Admin: deleted read-only key %q", name)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Key deleted"})
}
