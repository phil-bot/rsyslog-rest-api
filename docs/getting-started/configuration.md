# Configuration

All rsyslox settings are managed through the **Admin panel** at `/admin`. No manual config file editing is required.

## Admin Panel

Navigate to `http://<host>:8000/admin` and log in with your admin password.

### Server

| Setting | Description | Default |
|---|---|---|
| Bind host | Network interface to listen on (read-only after setup) | `0.0.0.0` |
| Port | TCP port (read-only after setup) | `8000` |
| CORS origins | Comma-separated allowed origins, `*` for all | `*` |
| SSL | Enable HTTPS — requires cert and key files | off |

### Database

Read-only view of the active database connection. Credentials are set once during setup and are not editable here.

### Cleanup

Automatic deletion of old log entries when disk usage exceeds a threshold. → See [Cleanup Guide](../guides/cleanup.md).

| Setting | Description | Default |
|---|---|---|
| Enabled | Toggle the cleanup service | off |
| Disk path | Partition to monitor (usually the MySQL data dir) | `/var/lib/mysql` |
| Threshold % | Delete entries when disk usage exceeds this | 85 % |
| Batch size | Rows deleted per cleanup run | 1 000 |
| Interval | Seconds between checks | 900 |

### API Keys

Named, revocable read-only API keys for external tools. Keys are shown in plaintext **once** at creation time — rsyslox stores only a SHA-256 hash. Pass a key via:

```
X-API-Key: <plaintext key>
```

Read-only keys can access `/api/logs` and `/api/meta` only. They cannot access admin endpoints.

### Preferences

Browser-persisted settings stored in `localStorage`. Apply instantly without restart and are independent per browser.

| Setting | Options | Default |
|---|---|---|
| Language | English, Deutsch | English |
| Time format | 24-hour, 12-hour | 24-hour |
| Font size | Small (13 px), Medium (14 px), Large (15 px) | Medium |
| Auto-refresh interval | 5–300 s | 30 s |

## Configuration File Reference

`/etc/rsyslox/config.toml` is written by the setup wizard and updated by the Admin panel. Manual editing is not required. The file is shown here for reference only.

```toml
[server]
host                  = "0.0.0.0"
port                  = 8000
use_ssl               = false
ssl_cert              = "/etc/rsyslox/certs/cert.pem"
ssl_key               = "/etc/rsyslox/certs/key.pem"
allowed_origins       = ["*"]
auto_refresh_interval = 30

[database]
host     = "localhost"
port     = 3306
name     = "Syslog"
user     = "rsyslox"
password = "enc:<base64>"   # AES-GCM encrypted by setup wizard

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

### Security Model

| Value | Storage |
|---|---|
| Database password | AES-GCM encrypted; key derived from `/etc/machine-id` — not portable between machines |
| Admin password | bcrypt hash (cost 12) |
| API key plaintext | Never stored; only SHA-256 hex hash written to disk |
| Config file | Mode `0640` — readable by `root` and group `rsyslox` only |

### SSL

Place a certificate and key at the paths defined in `ssl_cert` and `ssl_key` (default: `/etc/rsyslox/certs/`), then enable SSL in **Admin → Server**. rsyslox serves HTTPS on the same port.

**Self-signed certificate (development):**
```bash
sudo mkdir -p /etc/rsyslox/certs
sudo openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout /etc/rsyslox/certs/key.pem \
  -out /etc/rsyslox/certs/cert.pem \
  -days 365 -subj "/CN=localhost"
```

**Production:** Use Let's Encrypt certificates via Certbot — see [Deployment Guide](../guides/deployment.md).

## Next Steps

- [Quick Start Guide](quick-start.md)
- [Deployment Guide](../guides/deployment.md)
- [Security Guide](../guides/security.md)
