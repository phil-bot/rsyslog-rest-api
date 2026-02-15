# Deployment Guide

Production deployment guide for rsyslox.

## Production Checklist

Before deploying:

- ✅ API key configured
- ✅ SSL/TLS certificates ready
- ✅ Database credentials secured
- ✅ Firewall rules configured
- ✅ Reverse proxy setup
- ✅ Monitoring configured
- ✅ Backup strategy in place

## systemd Service

### Installation

```bash
# Copy service file
sudo cp rsyslox.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable and start
sudo systemctl enable --now rsyslox

# Check status
sudo systemctl status rsyslox
```

### Service File

`/etc/systemd/system/rsyslox.service`:

```ini
[Unit]
Description=rsyslox - rsyslog REST API
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/rsyslox
EnvironmentFile=/opt/rsyslox/.env
ExecStart=/opt/rsyslox/rsyslox
Restart=on-failure
RestartSec=5s

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/rsyslox

[Install]
WantedBy=multi-user.target
```

## Reverse Proxy

### nginx (Recommended)

```nginx
# /etc/nginx/sites-available/rsyslox

upstream rsyslox_api {
    server 127.0.0.1:8000;
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

server {
    listen 443 ssl http2;
    server_name api.example.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Logging
    access_log /var/log/nginx/rsyslox-access.log;
    error_log /var/log/nginx/rsyslox-error.log;

    location / {
        # Rate limiting
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://rsyslox_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name api.example.com;
    return 301 https://$server_name$request_uri;
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/rsyslox /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Apache

```apache
# /etc/apache2/sites-available/rsyslox.conf

<VirtualHost *:443>
    ServerName api.example.com

    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/api.example.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/api.example.com/privkey.pem

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:8000/
    ProxyPassReverse / http://127.0.0.1:8000/

    # Security Headers
    Header always set Strict-Transport-Security "max-age=31536000"
    Header always set X-Frame-Options "DENY"
    Header always set X-Content-Type-Options "nosniff"

    ErrorLog ${APACHE_LOG_DIR}/rsyslox-error.log
    CustomLog ${APACHE_LOG_DIR}/rsyslox-access.log combined
</VirtualHost>

<VirtualHost *:80>
    ServerName api.example.com
    Redirect permanent / https://api.example.com/
</VirtualHost>
```

Enable modules and site:
```bash
sudo a2enmod proxy proxy_http ssl headers
sudo a2ensite rsyslox
sudo apache2ctl configtest
sudo systemctl reload apache2
```

## Firewall Configuration

### UFW (Ubuntu/Debian)

```bash
# Allow SSH
sudo ufw allow ssh

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable
sudo ufw status
```

### firewalld (CentOS/RHEL)

```bash
# Allow HTTP/HTTPS
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https

# Reload
sudo firewall-cmd --reload
sudo firewall-cmd --list-all
```

## SSL/TLS Setup

### Let's Encrypt (Certbot)

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx  # Ubuntu/Debian
sudo yum install certbot python3-certbot-nginx      # CentOS/RHEL

# Obtain certificate
sudo certbot --nginx -d api.example.com

# Test auto-renewal
sudo certbot renew --dry-run
```

## Monitoring

### systemd Journal

```bash
# View logs
sudo journalctl -u rsyslox -n 50

# Follow logs
sudo journalctl -u rsyslox -f

# Logs since boot
sudo journalctl -u rsyslox -b
```

### Health Check Script

```bash
#!/bin/bash
# /usr/local/bin/rsyslox-healthcheck.sh

API_URL="http://localhost:8000"
ALERT_EMAIL="admin@example.com"

if ! curl -sf "$API_URL/health" > /dev/null; then
    echo "rsyslox health check failed!" | mail -s "API Alert" $ALERT_EMAIL
    systemctl restart rsyslox
fi
```

Add to crontab:
```bash
# Check every 5 minutes
*/5 * * * * /usr/local/bin/rsyslox-healthcheck.sh
```

## Backup Strategy

### Configuration Backup

```bash
#!/bin/bash
# /usr/local/bin/backup-rsyslox-config.sh

BACKUP_DIR="/backup/rsyslox"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup configuration
cp /opt/rsyslox/.env "$BACKUP_DIR/env_$DATE"
cp /etc/systemd/system/rsyslox.service "$BACKUP_DIR/service_$DATE"

# Keep only last 30 days
find "$BACKUP_DIR" -name "env_*" -mtime +30 -delete
find "$BACKUP_DIR" -name "service_*" -mtime +30 -delete
```

## Security Hardening

### Dedicated User

```bash
# Create dedicated user
sudo useradd -r -s /bin/false rsyslox-api

# Update service file to run as rsyslox-api
sudo nano /etc/systemd/system/rsyslox.service
# Change: User=rsyslox-api, Group=rsyslox-api

# Fix permissions
sudo chown rsyslox-api:rsyslox-api /opt/rsyslox/rsyslox
sudo chmod 500 /opt/rsyslox/rsyslox
sudo chown rsyslox-api:rsyslox-api /opt/rsyslox/.env
sudo chmod 400 /opt/rsyslox/.env

# Restart service
sudo systemctl daemon-reload
sudo systemctl restart rsyslox
```

### Database Security

```sql
-- Create READ-ONLY user
CREATE USER 'rsyslox_readonly'@'localhost'
  IDENTIFIED BY 'secure-password';

-- Grant SELECT only on SystemEvents
GRANT SELECT ON Syslog.SystemEvents TO 'rsyslox_readonly'@'localhost';

FLUSH PRIVILEGES;
```

Update `.env`:
```bash
DB_USER=rsyslox_readonly
DB_PASS=secure-password
```

## Updates

### Update Procedure

```bash
# 1. Download new version
wget https://github.com/phil-bot/rsyslox/releases/download/v0.X.X/rsyslox-linux-amd64

# 2. Stop service
sudo systemctl stop rsyslox

# 3. Backup current binary
sudo cp /opt/rsyslox/rsyslox /opt/rsyslox/rsyslox.backup

# 4. Replace binary
sudo mv rsyslox-linux-amd64 /opt/rsyslox/rsyslox
sudo chmod +x /opt/rsyslox/rsyslox

# 5. Start service
sudo systemctl start rsyslox

# 6. Verify
sudo systemctl status rsyslox
curl http://localhost:8000/health
```

## Performance Tuning

### Database Optimization

```sql
-- Add indexes for common queries
CREATE INDEX idx_receivedat ON SystemEvents(ReceivedAt);
CREATE INDEX idx_fromhost ON SystemEvents(FromHost);
CREATE INDEX idx_priority ON SystemEvents(Priority);
CREATE INDEX idx_facility ON SystemEvents(Facility);
```

### nginx Optimization

```nginx
# Increase worker connections
events {
    worker_connections 2048;
}

# Enable gzip compression
gzip on;
gzip_types application/json;
gzip_min_length 1000;
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u rsyslox -n 50

# Check configuration
sudo /opt/rsyslox/rsyslox --help

# Test manually
sudo -u rsyslox-api /opt/rsyslox/rsyslox
```

### High Memory Usage

```bash
# Check memory
free -h

# Restart service
sudo systemctl restart rsyslox
```

### Database Connection Issues

```bash
# Test database connection
mysql -u rsyslox_readonly -p Syslog -e "SELECT COUNT(*) FROM SystemEvents"

# Check credentials in .env
sudo cat /opt/rsyslox/.env | grep DB_
```

## More Resources

- [Security Guide](security.md) - Security best practices
- [Performance Guide](performance.md) - Performance tuning
- [Troubleshooting](troubleshooting.md) - Common issues
