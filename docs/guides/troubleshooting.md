# Troubleshooting Guide

## Quick Diagnosis

```bash
# Service status
sudo systemctl status rsyslox

# Recent logs
sudo journalctl -u rsyslox -n 100

# Health check
curl http://localhost:8000/health

# Database connectivity test
mysql -u rsyslox -p Syslog -e "SELECT COUNT(*) FROM SystemEvents"
```

---

## Common Issues

### Service Won't Start

**Symptoms:** `systemctl status rsyslox` shows "failed"

```bash
# View detailed error
sudo journalctl -u rsyslox -n 50

# Common causes:

# 1. Configuration missing or invalid
ls -la /etc/rsyslox/config.toml
# If missing: binary starts in setup wizard mode (normal for first install)

# 2. Database unreachable
mysql -u rsyslox -p Syslog   # Test manually

# 3. Port already in use
sudo lsof -i :8000

# 4. Binary not executable
ls -la /opt/rsyslox/rsyslox
sudo chmod +x /opt/rsyslox/rsyslox
```

### Setup Wizard Not Accessible

The wizard runs on port `8000` and is reachable from **any machine on the network** while no config exists. If you cannot reach it, check the firewall:

```bash
sudo ufw allow 8000/tcp
```

### Database Connection Failed

**Log message:** `Failed to connect to database`

```bash
# 1. Check database is running
sudo systemctl status mysql

# 2. Test credentials manually
mysql -h localhost -u rsyslox -p Syslog

# 3. Verify the database and table exist
mysql -u root -p -e "SHOW DATABASES; USE Syslog; SHOW TABLES;"

# 4. Re-run setup if credentials changed
# Delete config and restart → triggers setup wizard again
sudo systemctl stop rsyslox
sudo rm /etc/rsyslox/config.toml
sudo systemctl start rsyslox
# Then open http://localhost:8000 to redo setup
```

### Authentication Issues

**Admin panel won't accept password:**
- Verify Caps Lock
- If you've forgotten the password, reset it:
```bash
/opt/rsyslox/rsyslox hash-password "yournewpassword"
# Copy the output, edit /etc/rsyslox/config.toml:
# admin_password_hash = "<paste>"
sudo systemctl restart rsyslox
```

**API key rejected (`401 Unauthorized`):**
```bash
# Check header format — must be exactly:
curl -H "X-API-Key: your-plaintext-key" ...

# Not:
curl -H "API-Key: ..." ...
curl -H "Authorization: Bearer ..." ...
```

If the key was revoked in Admin → API Keys, create a new one.

**Session expired (redirect loop to /login):**
- Clear `sessionStorage` in your browser's DevTools and reload, or open a private window

### No Logs in the UI

**Symptoms:** Table is empty, total shows 0

```bash
# 1. Verify the database has data
mysql -u rsyslox -p Syslog -e "SELECT COUNT(*) FROM SystemEvents"

# 2. Check the time range — default is last 1h
# Extend it in the filter panel (try 24h or 7d)

# 3. Remove all active filters
# Click "Reset" in the filter panel

# 4. Verify rsyslog is writing to the database
tail -f /var/log/syslog   # Should show new entries
mysql -u rsyslox -p Syslog -e "SELECT ReceivedAt FROM SystemEvents ORDER BY ReceivedAt DESC LIMIT 5"
```

### Filters Not Working

```bash
# ❌ Wrong: comma-separated multi-value
?FromHost=web01,web02

# ✅ Correct: repeat the parameter
?FromHost=web01&FromHost=web02

# ❌ Wrong: lowercase parameter
?fromhost=web01

# ✅ Correct: CamelCase
?FromHost=web01

# ❌ Wrong: Priority (deprecated)
?Priority=3

# ✅ Correct: Severity
?Severity=3
```

### SSL Certificate Errors

```bash
# Verify certificate file exists and is readable
ls -la /etc/rsyslox/certs/
sudo openssl x509 -in /etc/rsyslox/certs/cert.pem -text -noout | head -20

# Verify key permissions
sudo chmod 600 /etc/rsyslox/certs/key.pem
sudo chown rsyslox:rsyslox /etc/rsyslox/certs/key.pem

# Restart after fixing
sudo systemctl restart rsyslox
```

### Performance Issues

**Slow queries:**

```bash
# Check database indexes
mysql -u rsyslox -p Syslog -e "SHOW INDEX FROM SystemEvents"

# Narrow the time window in your query — start with 1h, not 30d
# Reduce limit parameter: ?limit=100 instead of ?limit=10000
```

See [Performance Guide → Database Indexes](performance.md#database-indexes) for the recommended index definitions.

### Config File Issues

If the config file was corrupted or manually edited incorrectly, rsyslox will fail to start with a parse error in the journal. Re-run setup:

```bash
sudo systemctl stop rsyslox
sudo cp /etc/rsyslox/config.toml /etc/rsyslox/config.toml.bak
sudo rm /etc/rsyslox/config.toml
sudo systemctl start rsyslox
# Complete setup wizard at http://localhost:8000
```

---

## FAQ

### How do I access the API docs?

Navigate to `http://<host>:8000/docs` — interactive Redoc documentation is served directly from the binary.

### Can rsyslox run without internet access?

Yes — all assets (frontend, API docs, Redoc) are embedded in the binary. No CDN or external resources are required at runtime.

### How do I change the admin password?

See [Authentication Issues → Admin panel won't accept password](#authentication-issues) above.

### Can I run multiple instances?

Yes — each instance needs its own config file path. Use the `RSYSLOX_CONFIG` environment variable to override the default path `/etc/rsyslox/config.toml`.

### Does rsyslox support PostgreSQL?

Currently only MySQL/MariaDB is supported.

---

## Getting Help

- **Logs:** `sudo journalctl -u rsyslox -n 200`
- **Issues:** [GitHub Issues](https://github.com/phil-bot/rsyslox/issues)
- **Discussions:** [GitHub Discussions](https://github.com/phil-bot/rsyslox/discussions)
