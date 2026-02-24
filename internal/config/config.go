// Package config handles loading, saving and validating application configuration.
// Configuration is stored in TOML format at /etc/rsyslox/config.toml.
// During local development, an alternative path can be specified via the
// RSYSLOX_CONFIG environment variable.
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultConfigPath is the standard production config location.
	DefaultConfigPath = "/etc/rsyslox/config.toml"

	// EnvConfigPath overrides the config file path (for development).
	EnvConfigPath = "RSYSLOX_CONFIG"
)

// Load reads the configuration from disk.
// Returns (cfg, false, nil)  when the config file exists and is valid.
// Returns (defaults, true, nil) when the config file does not exist yet
// (first-run / setup-wizard mode).
// Returns (nil, false, err) on any other error.
func Load() (*Config, bool, error) {
	path := configPath()

	cfg := defaults()
	cfg.ConfigPath = path
	cfg.InstallPath = installPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// No config file → setup wizard needed.
		// Allow the port to be overridden via RSYSLOX_PORT so the installer
		// can pass a user-chosen port before any config.toml exists.
		if p := os.Getenv("RSYSLOX_PORT"); p != "" {
			var port int
			if _, err := fmt.Sscanf(p, "%d", &port); err == nil && port > 0 && port <= 65535 {
				cfg.Server.Port = port
			}
		}
		return cfg, true, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, false, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, false, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, false, nil
}

// Save writes the configuration to disk.
// The file is written with mode 0640 (root:rsyslox readable).
func Save(cfg *Config) error {
	path := cfg.ConfigPath
	if path == "" {
		path = configPath()
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// Validate checks that all required fields are set and values are in range.
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database.password is required")
	}
	if c.Auth.AdminPasswordHash == "" {
		return fmt.Errorf("auth.admin_password_hash is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if c.Cleanup.ThresholdPercent <= 0 || c.Cleanup.ThresholdPercent > 100 {
		return fmt.Errorf("cleanup.threshold_percent must be between 1 and 100")
	}
	return nil
}

// DSN builds a MySQL DSN string from the database configuration.
// The password is decrypted if it has the "enc:" prefix.
func (c *Config) DSN() (string, error) {
	pass, err := DecryptPassword(c.Database.Password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt database password: %w", err)
	}
	port := c.Database.Port
	if port == 0 {
		port = 3306
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		c.Database.User, pass, c.Database.Host, port, c.Database.Name), nil
}

// configPath returns the active configuration file path.
func configPath() string {
	if p := os.Getenv(EnvConfigPath); p != "" {
		return p
	}
	return DefaultConfigPath
}

// installPath returns the directory of the running binary.
func installPath() string {
	exe, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "/opt/rsyslox"
	}
	abs, err := filepath.Abs(exe)
	if err != nil {
		return "/opt/rsyslox"
	}
	return filepath.Dir(abs)
}

// ActiveConfigPath returns the config file path that is currently in use.
// Exported so other packages (e.g. the health handler) can check whether
// the file exists without importing the full config loading logic.
func ActiveConfigPath() string {
	return configPath()
}
