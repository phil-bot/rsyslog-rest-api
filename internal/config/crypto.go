package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

const encPrefix = "enc:"

// EncryptPassword encrypts a plaintext password using AES-256-GCM.
// The encryption key is derived from the machine ID so the ciphertext
// is only decryptable on the same host.
// Returns a string with the "enc:" prefix.
func EncryptPassword(plaintext string) (string, error) {
	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encPrefix + encoded, nil
}

// DecryptPassword decrypts a password encrypted by EncryptPassword.
// If the value does not have the "enc:" prefix it is returned as-is
// (plaintext fallback for migration / first-run).
func DecryptPassword(value string) (string, error) {
	if !strings.HasPrefix(value, encPrefix) {
		return value, nil
	}

	encoded := strings.TrimPrefix(value, encPrefix)
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted password: %w", err)
	}

	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted reports whether a password value is already encrypted.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, encPrefix)
}

// deriveKey derives a 32-byte AES key from the machine ID.
func deriveKey() ([]byte, error) {
	machineID, err := readMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to read machine ID: %w", err)
	}
	const salt = "rsyslox-v1-config-key"
	h := sha256.Sum256([]byte(machineID + salt))
	return h[:], nil
}

// readMachineID returns a stable host identifier.
//
// Priority:
//  1. /etc/machine-id          (systemd, most Linux distros)
//  2. /var/lib/dbus/machine-id (older systems)
//  3. /etc/rsyslox/machine-id  (generated once; covers Docker containers and
//     any other environment without a system machine-id)
func readMachineID() (string, error) {
	// Standard system locations
	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		data, err := os.ReadFile(path)
		if err == nil && len(strings.TrimSpace(string(data))) > 0 {
			return strings.TrimSpace(string(data)), nil
		}
	}

	// Fallback: persistent generated ID in the config directory
	const fallbackPath = "/etc/rsyslox/machine-id"
	data, err := os.ReadFile(fallbackPath)
	if err == nil && len(strings.TrimSpace(string(data))) > 0 {
		return strings.TrimSpace(string(data)), nil
	}

	// Generate and persist a new ID
	id, err := generateMachineID()
	if err != nil {
		return "", fmt.Errorf("could not determine machine ID: %w", err)
	}

	// Best-effort write — if it fails (e.g. read-only FS) the ID is
	// ephemeral for this run, but setup will still succeed.
	_ = os.MkdirAll("/etc/rsyslox", 0755)
	_ = os.WriteFile(fallbackPath, []byte(id+"\n"), 0644)

	return id, nil
}

// generateMachineID returns a random 32-hex-character string.
func generateMachineID() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
