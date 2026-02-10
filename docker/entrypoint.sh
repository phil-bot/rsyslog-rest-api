#!/bin/bash
set -e

echo "================================================"
echo "rsyslog REST API - Test Environment"
echo "================================================"
echo ""

# Check if binary exists in mounted directory
if [ ! -f /host-build/rsyslog-rest-api ]; then
    echo "✗ ERROR: Binary not found!"
    echo ""
    echo "Please build first:"
    echo "  cd .. && make build-static"
    echo ""
    exit 1
fi

# Copy binary to installation directory
echo "[1/8] Installing API binary..."
cp /host-build/rsyslog-rest-api /opt/rsyslog-rest-api/
chmod +x /opt/rsyslog-rest-api/rsyslog-rest-api
echo "✓ Binary installed ($(ls -lh /opt/rsyslog-rest-api/rsyslog-rest-api | awk '{print $5}'))"

# Start MariaDB
echo "[2/8] Starting MariaDB..."
mysqld_safe --datadir=/var/lib/mysql --user=mysql &
sleep 5

# Wait for MariaDB
for i in {1..30}; do
    if mysqladmin ping --silent 2>/dev/null; then
        echo "✓ MariaDB ready"
        break
    fi
    [ $i -eq 30 ] && echo "✗ MariaDB timeout!" && exit 1
    sleep 1
done

# Create database and user
echo "[3/8] Creating database..."
mysql <<EOF
CREATE DATABASE IF NOT EXISTS Syslog;
CREATE USER IF NOT EXISTS 'rsyslog'@'localhost' IDENTIFIED BY 'password';
GRANT ALL ON Syslog.* TO 'rsyslog'@'localhost';
FLUSH PRIVILEGES;
EOF

# Create SystemEvents table
mysql Syslog <<'EOF'
CREATE TABLE IF NOT EXISTS SystemEvents (
    ID int unsigned not null auto_increment primary key,
    CustomerID bigint,
    ReceivedAt datetime NULL,
    DeviceReportedTime datetime NULL,
    Facility smallint NULL,
    Priority smallint NULL,
    FromHost varchar(60) NULL,
    Message text,
    NTSeverity int NULL,
    Importance int NULL,
    EventSource varchar(60),
    EventUser varchar(60) NULL,
    EventCategory int NULL,
    EventID int NULL,
    EventBinaryData text NULL,
    MaxAvailable int NULL,
    CurrUsage int NULL,
    MinUsage int NULL,
    MaxUsage int NULL,
    InfoUnitID int NULL,
    SysLogTag varchar(60),
    EventLogType varchar(60),
    GenericFileName VarChar(60),
    SystemID int NULL
);
EOF

# Insert test data
echo "[4/8] Inserting test data..."
mysql Syslog <<'EOF'
INSERT INTO SystemEvents (ReceivedAt, DeviceReportedTime, FromHost, Priority, Facility, Message, SysLogTag) VALUES
(NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR, 'webserver01', 6, 1, 'User login successful: admin', 'sshd'),
(NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 2 HOUR, 'webserver01', 3, 1, 'Failed login attempt from 192.168.1.100', 'sshd'),
(NOW() - INTERVAL 3 HOUR, NOW() - INTERVAL 3 HOUR, 'dbserver01', 4, 4, 'Database connection timeout', 'mysqld'),
(NOW() - INTERVAL 4 HOUR, NOW() - INTERVAL 4 HOUR, 'dbserver01', 6, 4, 'Query executed successfully', 'mysqld'),
(NOW() - INTERVAL 5 HOUR, NOW() - INTERVAL 5 HOUR, 'appserver01', 5, 16, 'Application started on port 3000', 'node'),
(NOW() - INTERVAL 6 HOUR, NOW() - INTERVAL 6 HOUR, 'appserver01', 3, 16, 'Critical error in module auth', 'node'),
(NOW() - INTERVAL 7 HOUR, NOW() - INTERVAL 7 HOUR, 'webserver02', 6, 1, 'HTTP request: GET /api/users', 'nginx'),
(NOW() - INTERVAL 8 HOUR, NOW() - INTERVAL 8 HOUR, 'webserver02', 4, 1, 'Slow response time detected: 2.5s', 'nginx'),
(NOW() - INTERVAL 9 HOUR, NOW() - INTERVAL 9 HOUR, 'mailserver01', 2, 3, 'Mail queue growing rapidly', 'postfix'),
(NOW() - INTERVAL 10 HOUR, NOW() - INTERVAL 10 HOUR, 'mailserver01', 6, 3, 'Email sent successfully', 'postfix');
EOF

LOGCOUNT=$(mysql -N Syslog -e 'SELECT COUNT(*) FROM SystemEvents')
echo "✓ Database ready ($LOGCOUNT log entries)"

# Create rsyslog config
echo "[5/8] Configuring rsyslog..."
cat > /etc/rsyslog.d/mysql.conf <<'EOF'
module(load="ommysql")
action(type="ommysql" server="localhost" db="Syslog" uid="rsyslog" pwd="password")
EOF
chmod 600 /etc/rsyslog.d/mysql.conf

# Start rsyslog
rsyslogd
echo "✓ rsyslog started"

# Configure API
echo "[6/8] Configuring API..."
cat > /opt/rsyslog-rest-api/.env <<EOF
API_KEY=${API_KEY:-test123456789}
SERVER_HOST=0.0.0.0
SERVER_PORT=${SERVER_PORT:-8000}
USE_SSL=false
ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-*}
RSYSLOG_CONFIG_PATH=/etc/rsyslog.d/mysql.conf
EOF

echo "✓ API configured"

# Start API
echo "[7/8] Starting API..."
cd /opt/rsyslog-rest-api
./rsyslog-rest-api > /var/log/rsyslog-rest-api.log 2>&1 &
API_PID=$!

# Wait for API to start
sleep 3
if kill -0 $API_PID 2>/dev/null; then
    echo "✓ API started (PID: $API_PID)"
else
    echo "✗ API failed to start!"
    cat /var/log/rsyslog-rest-api.log
    exit 1
fi

# Test API
echo "[8/8] Testing API..."
if curl -s http://localhost:8000/health > /dev/null; then
    echo "✓ API health check passed"
else
    echo "⚠ API health check failed (may still be starting)"
fi

echo ""
echo "================================================"
echo "✓ Environment Ready!"
echo "================================================"
echo ""
echo "API:      http://localhost:8000"
echo "API Key:  ${API_KEY:-test123456789}"
echo "Database: Syslog (rsyslog/password)"
echo "Logs:     $LOGCOUNT entries"
echo ""
echo "Test commands:"
echo "  curl http://localhost:8000/health"
echo "  curl -H 'X-API-Key: ${API_KEY:-test123456789}' http://localhost:8000/logs?limit=5"
echo "  curl -H 'X-API-Key: ${API_KEY:-test123456789}' http://localhost:8000/meta/FromHost"
echo ""
echo "Logs: tail -f /var/log/rsyslog-rest-api.log"
echo ""

exec "$@"
