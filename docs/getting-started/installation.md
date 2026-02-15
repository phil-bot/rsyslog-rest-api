# Installation

Complete guide for installing rsyslox.

## Prerequisites

### System Requirements

- Linux (Ubuntu, Debian, CentOS, RHEL)
- Architecture: x86_64 or ARM64
- rsyslog with MySQL/MariaDB support
- MySQL 5.7+ or MariaDB 10.3+
- Minimum 256 MB RAM

### rsyslog MySQL Setup

Before installation, configure rsyslog:

```bash
# Install rsyslog-mysql
sudo apt-get install rsyslog-mysql  # Ubuntu/Debian
sudo yum install rsyslog-mysql       # CentOS/RHEL

# Create rsyslog configuration
sudo nano /etc/rsyslog.d/mysql.conf
```

Content of `/etc/rsyslog.d/mysql.conf`:
```
module(load="ommysql")
action(type="ommysql" server="localhost" db="Syslog" uid="rsyslog" pwd="password")
```

```bash
# Restart rsyslog
sudo systemctl restart rsyslog
```

## Installation Methods

### Method 1: Binary (Recommended)

Fastest and easiest method:

```bash
# Download latest release
wget https://github.com/phil-bot/rsyslox/releases/latest/download/rsyslox-linux-amd64

# Make executable
chmod +x rsyslox-linux-amd64

# Move to system path
sudo mv rsyslox-linux-amd64 /usr/local/bin/rsyslox

# Verify
rsyslox --version
```

### Method 2: From Source

For developers:

```bash
# Prerequisites
# - Go 1.21+
# - git
# - make

# Clone repository
git clone https://github.com/phil-bot/rsyslox.git
cd rsyslox

# Build and install
make build-static
sudo make install
```

## Configuration

Create configuration file:

```bash
# Create directory
sudo mkdir -p /opt/rsyslox

# Create .env
sudo nano /opt/rsyslox/.env
```

Minimal configuration:

```bash
# API Key (required for production!)
API_KEY=your-secret-key-here

# Database
DB_HOST=localhost
DB_NAME=Syslog
DB_USER=rsyslog
DB_PASS=your-database-password

# Server
SERVER_PORT=8000
```

**Generate secure API key:**
```bash
openssl rand -hex 32
```

## Verification

### Test Service

```bash
# Start in foreground (testing)
cd /opt/rsyslox
rsyslox
```

### Health Check

```bash
curl http://localhost:8000/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2025-02-15T10:30:00Z"
}
```

### API Test

```bash
# Get API key
API_KEY=$(sudo grep "^API_KEY=" /opt/rsyslox/.env | cut -d'=' -f2)

# Test API
curl -H "X-API-Key: $API_KEY" "http://localhost:8000/logs?limit=5"
```

## Production Setup

For production deployment with systemd service:

→ See [Deployment Guide](../guides/deployment.md)

## Troubleshooting

Common installation issues:

**Binary not found:**
- Check path: `which rsyslox`
- Verify permissions: `ls -la /usr/local/bin/rsyslox`

**Database connection failed:**
- Check credentials in `.env`
- Test MySQL connection: `mysql -u rsyslog -p Syslog`

**Permission denied:**
- Fix `.env` permissions: `sudo chmod 600 /opt/rsyslox/.env`

For more issues see [Troubleshooting Guide](../guides/troubleshooting.md).

## Next Steps

- [Configure the API](configuration.md)
- [Quick Start Tutorial](quick-start.md)
- [Deploy to Production](../guides/deployment.md)
