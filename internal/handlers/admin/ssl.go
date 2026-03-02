package admin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/phil-bot/rsyslox/internal/config"
	"github.com/phil-bot/rsyslox/internal/models"
)

// SSLHandler handles SSL certificate management.
// POST /api/admin/ssl/generate  — generate a self-signed certificate
// POST /api/admin/ssl/upload    — upload a custom cert + key (multipart/form-data)
type SSLHandler struct {
	cfg *config.Config
}

func NewSSLHandler(cfg *config.Config) *SSLHandler {
	return &SSLHandler{cfg: cfg}
}

func (h *SSLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/admin/ssl/generate":
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed,
				models.NewAPIError("METHOD_NOT_ALLOWED", "Only POST is allowed"))
			return
		}
		h.handleGenerate(w, r)
	case "/api/admin/ssl/upload":
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed,
				models.NewAPIError("METHOD_NOT_ALLOWED", "Only POST is allowed"))
			return
		}
		h.handleUpload(w, r)
	default:
		respondError(w, http.StatusNotFound,
			models.NewAPIError("NOT_FOUND", "Unknown SSL endpoint"))
	}
}

// handleGenerate creates a self-signed ECDSA P-256 certificate valid for 10 years
// and writes it to the configured cert/key paths.
func (h *SSLHandler) handleGenerate(w http.ResponseWriter, _ *http.Request) {
	certPath := h.cfg.Server.SSLCertFile
	keyPath  := h.cfg.Server.SSLKeyFile

	if certPath == "" {
		certPath = "/etc/rsyslox/certs/cert.pem"
	}
	if keyPath == "" {
		keyPath = "/etc/rsyslox/certs/key.pem"
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to create certificate directory: "+err.Error()))
		return
	}

	// Generate ECDSA P-256 key (smaller and faster than RSA-2048, broadly supported)
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to generate private key: "+err.Error()))
		return
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to generate serial number: "+err.Error()))
		return
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"rsyslox"},
			CommonName:   "rsyslox self-signed",
		},
		NotBefore:             now,
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &privKey.PublicKey, privKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to create certificate: "+err.Error()))
		return
	}

	// Write certificate
	certFile, err := os.OpenFile(certPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to write certificate file: "+err.Error()))
		return
	}
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certFile.Close()
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to encode certificate: "+err.Error()))
		return
	}
	certFile.Close()

	// Write private key
	keyDER, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to marshal private key: "+err.Error()))
		return
	}
	keyFile, err := os.OpenFile(keyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to write key file: "+err.Error()))
		return
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		keyFile.Close()
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to encode private key: "+err.Error()))
		return
	}
	keyFile.Close()

	validUntil := now.Add(10 * 365 * 24 * time.Hour).Format(time.RFC3339)
	log.Printf("Admin: generated self-signed certificate at %s (valid until %s)", certPath, validUntil)

	type sslResponse struct {
		CertPath   string `json:"cert_path"`
		KeyPath    string `json:"key_path"`
		ValidUntil string `json:"valid_until"`
	}
	body, _ := json.Marshal(sslResponse{CertPath: certPath, KeyPath: keyPath, ValidUntil: validUntil})
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// handleUpload accepts a multipart form with fields "cert" and "key"
// and saves them to the configured paths.
func (h *SSLHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	const maxSize = 1 << 20 // 1 MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		respondError(w, http.StatusBadRequest,
			models.NewAPIError("INVALID_PARAMETER", "Failed to parse multipart form"))
		return
	}

	certPath := h.cfg.Server.SSLCertFile
	keyPath  := h.cfg.Server.SSLKeyFile

	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("INTERNAL_ERROR", "Failed to create certificate directory"))
		return
	}

	if err := saveFormFile(r, "cert", certPath, 0644); err != nil {
		respondError(w, http.StatusBadRequest,
			models.NewAPIError("INVALID_PARAMETER", "Missing or unreadable 'cert' field: "+err.Error()))
		return
	}
	if err := saveFormFile(r, "key", keyPath, 0600); err != nil {
		// Roll back cert if key fails
		_ = os.Remove(certPath)
		respondError(w, http.StatusBadRequest,
			models.NewAPIError("INVALID_PARAMETER", "Missing or unreadable 'key' field: "+err.Error()))
		return
	}

	log.Printf("Admin: uploaded custom SSL certificate to %s", certPath)
	respondJSON(w, http.StatusOK, map[string]string{
		"cert_path": certPath,
		"key_path":  keyPath,
	})
}

func saveFormFile(r *http.Request, field, dest string, mode os.FileMode) error {
	f, _, err := r.FormFile(field)
	if err != nil {
		return err
	}
	defer f.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, f)
	return err
}
