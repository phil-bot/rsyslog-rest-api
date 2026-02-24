# Operations Guide

## Installation

### Prerequisites

- Linux x86-64 or ARM64
- MySQL or MariaDB database populated by rsyslog (or compatible)
- systemd
- Root access for the installer

### Install

Download the release binary and run the installer:

```bash
chmod +x rsyslox
sudo ./install.sh
```

The installer:

1. Creates a system user and group `rsyslox`
2. Copies the binary to `/opt/rsyslox/rsyslox`
3. Installs and starts a systemd service
4. Prints the URL for the setup wizard

### Setup Wizard

On first start rsyslox has no configuration and serves a setup wizard on `http://localhost:8000` (accessible from localhost only). Open it in a browser and fill in:

- **Database** — host, port, name, user, password
- **Admin password** — minimum 12 characters
- **Server** — bind host, port, optional CORS origins

rsyslox writes `/etc/rsyslox/config.toml` and immediately begins serving the log viewer. No restart is required.

### Uninstall

```bash
sudo ./install.sh --uninstall
```

Stops and removes the service and binary. Configuration at `/etc/rsyslox/` is kept and must be removed manually if no longer needed.

---

## Admin Panel

All server-side settings are available at `/admin`. Changes take effect immediately unless noted.

### Server

| Field | Description |
|---|---|
| Bind host | Interface to listen on (read-only after setup) |
| Port | TCP port (read-only after setup) |
| CORS origins | Comma-separated allowed origins, `*` for all |
| SSL | Enable HTTPS — requires cert and key at `/etc/rsyslox/certs/` |

### Database

Read-only view of the active database connection. Credentials are set during setup and cannot be changed here.

### Cleanup

Automatic deletion of old log entries when disk usage exceeds a threshold.

| Field | Description | Default |
|---|---|---|
| Enabled | Toggle cleanup | off |
| Disk path | Path to monitor (usually the MySQL data directory) | `/var/lib/mysql` |
| Threshold % | Delete entries when disk usage exceeds this | 85 % |
| Batch size | Rows deleted per cleanup run | 1 000 |
| Interval | Seconds between cleanup checks | 900 |

### API Keys

Read-only API keys allow external tools to query `/api/logs` and `/api/meta` without admin credentials. Keys are shown in plaintext once at creation; rsyslox stores only a SHA-256 hash.

Pass a key via the `X-API-Key` header:

```
X-API-Key: <plaintext key>
```

### Preferences

Browser-side settings stored in `localStorage`. Apply immediately without restart.

| Setting | Options | Default |
|---|---|---|
| Language | English, Deutsch | English |
| Time format | 24-hour, 12-hour | 24-hour |
| Font size | Small (13 px), Medium (14 px), Large (15 px) | Medium |
| Auto-refresh interval | 5–300 s | 30 s |

---

## Configuration File

`/etc/rsyslox/config.toml` is written by the setup wizard and updated by the Admin panel. Manual editing is not required. The file is shown here for reference.

```toml
[server]
host                 = "0.0.0.0"
port                 = 8000
use_ssl              = false
ssl_cert             = "/etc/rsyslox/certs/cert.pem"
ssl_key              = "/etc/rsyslox/certs/key.pem"
allowed_origins      = ["*"]
auto_refresh_interval = 30   # seconds; UI overrides this per-browser

[database]
host     = "localhost"
port     = 3306
name     = "Syslog"
user     = "rsyslox"
password = "enc:<base64>"   # AES-GCM encrypted; written by setup wizard

[auth]
admin_password_hash = "$2a$12$..."   # bcrypt hash

[[auth.read_only_keys]]
name     = "monitoring"
key_hash = "<sha256 hex>"

[cleanup]
enabled           = false
disk_path         = "/var/lib/mysql"
threshold_percent = 85.0
batch_size        = 1000
interval          = "15m"
```

### Security

- The database password is encrypted with AES-GCM. The key is derived from the machine's `/etc/machine-id` — the encrypted value is not portable between machines.
- The admin password is hashed with bcrypt (cost 12).
- API key plaintext is never stored; only the SHA-256 hex hash is written to disk.
- The config file is written with mode `0640` (owner `root`, group `rsyslox`).

### SSL

Place a certificate and key at the paths defined in `ssl_cert` and `ssl_key`, then enable SSL in Admin → Server. rsyslox will start serving HTTPS on the same port.

---

## Service Management

```bash
# Status
systemctl status rsyslox

# Restart
sudo systemctl restart rsyslox

# Logs
journalctl -u rsyslox -f
```

---

## Updating

1. Stop the service: `sudo systemctl stop rsyslox`
2. Replace the binary at `/opt/rsyslox/rsyslox`
3. Start the service: `sudo systemctl start rsyslox`

Configuration is preserved across updates.

---

## Troubleshooting

**White screen / redirect loop after update**

Clear `sessionStorage` in the browser (or open a private window) — the session token format may have changed.

**Service fails to start**

Check `journalctl -u rsyslox -e` for the error. Common causes: database unreachable, config file missing or invalid.

**Config file path**

The `RSYSLOX_CONFIG` environment variable overrides the default path `/etc/rsyslox/config.toml`. Useful when running multiple instances or during development.
