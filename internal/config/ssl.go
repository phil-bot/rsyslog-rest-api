package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// EnsureSSLCerts checks whether the configured cert and key files exist.
// If either is missing it generates a self-signed ECDSA P-256 certificate
// valid for 10 years and writes both files to disk.
// Returns an error only if generation or writing fails.
func EnsureSSLCerts(cfg *ServerConfig) error {
	certPath := cfg.SSLCertFile
	keyPath  := cfg.SSLKeyFile

	if certPath == "" {
		certPath = "/etc/rsyslox/certs/cert.pem"
	}
	if keyPath == "" {
		keyPath = "/etc/rsyslox/certs/key.pem"
	}

	// Both files already exist — nothing to do.
	certOK := fileExists(certPath)
	keyOK  := fileExists(keyPath)
	if certOK && keyOK {
		return nil
	}

	log.Printf("SSL: certificate or key not found — generating self-signed cert (%s)", certPath)

	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return fmt.Errorf("ssl: failed to create cert directory: %w", err)
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("ssl: failed to generate private key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("ssl: failed to generate serial number: %w", err)
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
		return fmt.Errorf("ssl: failed to create certificate: %w", err)
	}

	// Write certificate
	certFile, err := os.OpenFile(certPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("ssl: failed to open cert file for writing: %w", err)
	}
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certFile.Close()
		return fmt.Errorf("ssl: failed to encode certificate: %w", err)
	}
	certFile.Close()

	// Write private key
	keyDER, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("ssl: failed to marshal private key: %w", err)
	}
	keyFile, err := os.OpenFile(keyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("ssl: failed to open key file for writing: %w", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		keyFile.Close()
		return fmt.Errorf("ssl: failed to encode private key: %w", err)
	}
	keyFile.Close()

	log.Printf("✓ SSL: self-signed certificate generated (valid until %s)",
		now.Add(10*365*24*time.Hour).Format("2006-01-02"))
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
