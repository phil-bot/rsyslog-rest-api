# Docker Test Environment

Simple test environment with rsyslog + MariaDB + test data.

Binary is built **outside** Docker for simplicity.

## Quick Start

```bash
# 1. Build the API (on host)
make build-static

# 2. Start Docker environment
cd docker
docker-compose up -d

# 3. Wait for startup (watch logs)
docker-compose logs -f
# Wait for "Environment Ready!" then Ctrl+C

# 4. Test
curl http://localhost:8000/health
curl -H "X-API-Key: test123456789" "http://localhost:8000/logs?limit=5"

# 5. Run test suite
./test.sh
```

That's it! 🎉

## What happens?

1. ✅ You build the binary on your host with `make build-static`
2. ✅ Docker mounts `../build/` directory
3. ✅ Container copies binary to `/opt/rsyslog-rest-api/`
4. ✅ Container sets up MariaDB + rsyslog + test data
5. ✅ Container starts the API automatically


## API Testing

### From host
```bash
# Health Check
curl http://localhost:8000/health

# Get logs
curl -H "X-API-Key: test123456789" \
  "http://localhost:8000/logs?limit=5"

# Meta data
curl -H "X-API-Key: test123456789" \
  "http://localhost:8000/meta/FromHost"

# Only errors
curl -H "X-API-Key: test123456789" \
  "http://localhost:8000/logs?Priority=3"

# Run full test suite
cd docker
./test.sh
```

### Inside container
```bash
# Open shell
docker exec -it rsyslog-rest-api-test bash

# Check API process
ps aux | grep rsyslog-rest-api

# View logs
tail -f /var/log/rsyslog-rest-api.log

# Check installation
ls -la /opt/rsyslog-rest-api/
cat /opt/rsyslog-rest-api/.env
```

## Rebuild after code changes

```bash
# 1. Stop container
cd docker
docker-compose down

# 2. Rebuild binary
cd ..
make build-static

# 3. Restart container
cd docker
docker-compose up -d

# Binary is automatically copied on startup
```

## Konfiguration ändern

### API-Key ändern
```bash
# In docker-compose.yml
environment:
  - API_KEY=mein-neuer-key

# Neu starten
docker-compose down
docker-compose up -d
```

### Port ändern
```bash
# In docker-compose.yml
ports:
  - "8080:8000"  # Host:Container

environment:
  - SERVER_PORT=8000  # Container-intern bleibt 8000
```

## Datenbank

### Zugriff
```bash
# Im Container
docker exec -it rsyslog-rest-api-test bash
mysql -u rsyslog -ppassword Syslog

# Testdaten ansehen
SELECT ReceivedAt, FromHost, Priority, Message FROM SystemEvents LIMIT 5;
```

### Eigene Daten einfügen
```bash
docker exec -it rsyslog-rest-api-test mysql -u rsyslog -ppassword Syslog <<'EOF'
INSERT INTO SystemEvents (ReceivedAt, FromHost, Priority, Facility, Message, SysLogTag)
VALUES (NOW(), 'testhost', 6, 1, 'Test message', 'test');
EOF

# Prüfen
curl -H "X-API-Key: test123456789" \
  "http://localhost:8000/logs?FromHost=testhost"
```

## Was ist wo?

```
Container Layout:
├── /opt/rsyslog-rest-api/
│   ├── rsyslog-rest-api          # Binary
│   └── .env                      # Konfiguration
├── /etc/rsyslog.d/mysql.conf     # rsyslog Config
├── /var/log/rsyslog-rest-api.log # API Logs
└── /var/lib/mysql/Syslog/        # Datenbank
```

## Troubleshooting

### Container startet nicht
```bash
# Logs ansehen
docker-compose logs

# Neu bauen
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### API läuft nicht
```bash
# Logs prüfen
docker exec -it rsyslog-rest-api-test cat /var/log/rsyslog-rest-api.log

# Manuell starten
docker exec -it rsyslog-rest-api-test bash
cd /opt/rsyslog-rest-api
export $(cat .env | xargs)
./rsyslog-rest-api
```

### Datenbank leer
```bash
# Anzahl Einträge prüfen
docker exec -it rsyslog-rest-api-test \
  mysql -u rsyslog -ppassword Syslog -e "SELECT COUNT(*) FROM SystemEvents"

# Sollte 10 sein
```

### Port bereits belegt
```bash
# Anderen Port nutzen
# In docker-compose.yml ändern:
ports:
  - "8080:8000"

# Dann testen mit:
curl http://localhost:8080/health
```

## Neu starten

```bash
# Soft restart (Daten bleiben)
docker-compose restart

# Hard restart (neu bauen)
docker-compose down
docker-compose up -d --build

# Alles löschen (inkl. Daten!)
docker-compose down -v
```

## Test-Script

Automatisierte Tests:

```bash
# Im docker/ Verzeichnis
./test.sh

# Oder mit custom API key
API_KEY=mein-key ./test.sh
```

## Unterschied zu echtem Server

| Feature | Docker | Echter Server |
|---------|--------|---------------|
| API Start | Automatisch | systemd |
| Datenbank | Auto-Setup | Manuell |
| Testdaten | Inklusive | Keine |

Der Container simuliert die **komplette Installation** - perfekt zum Testen!

## Performance-Test

```bash
# Apache Bench (falls installiert)
ab -n 1000 -c 10 -H "X-API-Key: test123456789" \
  http://localhost:8000/logs?limit=10

# Oder einfach:
time curl -H "X-API-Key: test123456789" \
  "http://localhost:8000/logs?limit=100"
```

## Logs beobachten

```bash
# Container-Logs
docker-compose logs -f

# API-Logs
docker exec -it rsyslog-rest-api-test \
  tail -f /var/log/rsyslog-rest-api.log

# MariaDB-Logs
docker exec -it rsyslog-rest-api-test \
  tail -f /var/log/mysql/error.log
```

## Stop & Clean

```bash
# Stoppen (Daten bleiben)
docker-compose stop

# Stoppen und entfernen (Daten bleiben)
docker-compose down

# ALLES löschen (inkl. Daten!)
docker-compose down -v
```
