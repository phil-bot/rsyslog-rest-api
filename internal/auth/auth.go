// Package auth provides authentication helpers for the admin and read-only API.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/phil-bot/rsyslox/internal/config"
)

const bcryptCost = 12

// Role represents the access level of an authenticated caller.
type Role int

const (
	RoleNone     Role = iota // not authenticated
	RoleReadOnly             // valid read-only API key
	RoleAdmin                // valid admin password (session token)
)

// Manager handles password hashing and API key validation.
type Manager struct {
	cfg *config.Config
}

// New creates a new Manager.
func New(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

// HashAdminPassword hashes a plaintext password using bcrypt.
// The result should be stored in config.Auth.AdminPasswordHash.
func HashAdminPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyAdminPassword checks a plaintext password against the stored bcrypt hash.
func (m *Manager) VerifyAdminPassword(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(m.cfg.Auth.AdminPasswordHash),
		[]byte(plaintext),
	)
	return err == nil
}

// GenerateReadOnlyKey generates a new random API key and its SHA-256 hash.
// The caller receives the plaintext key (shown once) and the hash to store.
func GenerateReadOnlyKey() (plaintext, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("failed to generate key: %w", err)
	}
	plaintext = hex.EncodeToString(raw)
	hash = hashKey(plaintext)
	return plaintext, hash, nil
}

// VerifyReadOnlyKey checks whether the given API key matches any stored read-only key.
// Returns the name of the matching key, or empty string if none matches.
func (m *Manager) VerifyReadOnlyKey(key string) string {
	h := hashKey(key)
	for _, k := range m.cfg.Auth.ReadOnlyKeys {
		if k.KeyHash == h {
			return k.Name
		}
	}
	return ""
}

// hashKey returns the hex-encoded SHA-256 hash of a key.
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
