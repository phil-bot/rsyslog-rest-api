# Docker Testing Environment

Complete test environment with live log generation.

## Features

- ✅ Ubuntu 24.04 + rsyslog + MariaDB
- ✅ 10 initial test entries
- ✅ **Live log generator** - 3 logs every 10 seconds
- ✅ All 25+ fields populated
- ✅ No API key required (optional)

## Quick Start

```bash
# 1. Build binary
make build-static

# 2. Start Docker
cd docker
docker-compose up -d

# 3. Test
curl "http://localhost:8000/health"
curl "http://localhost:8000/logs?limit=5"
```

## What's Included

```
docker/
├── Dockerfile
├── docker-compose.yml
├── entrypoint.sh
├── log-generator.sh    # Live log generation
└── test.sh            # Test suite
```

## Live Log Generator

Generates realistic logs every 10 seconds:

- **Hosts:** webserver01, webserver02, dbserver01, appserver01, mailserver01, firewall01
- **Tags:** sshd, nginx, mysqld, node, postfix, iptables
- **Priorities:** Weighted (more INFO, less CRITICAL)
- **All fields:** EventID, EventSource, EventUser, etc.

```bash
# Watch logs being generated
docker exec rsyslox-test tail -f /var/log/log-generator.log

# Check database growth
watch -n 5 'docker exec rsyslox-test mysql -N Syslog -e "SELECT COUNT(*) FROM SystemEvents"'
```

## Testing Features

### Multi-Value Filters

```bash
# Multiple hosts
curl "http://localhost:8000/logs?FromHost=webserver01&FromHost=webserver02&limit=5"

# Multiple priorities
curl "http://localhost:8000/logs?Priority=3&Priority=4&limit=5"

# Combined
curl "http://localhost:8000/logs?FromHost=webserver01&FromHost=dbserver01&Priority=3&Priority=6"
```

### Structured Errors (v0.2.3+)

```bash
# Test invalid priority
curl "http://localhost:8000/logs?Priority=99"

# Expected response:
{
  "code": "INVALID_PRIORITY",
  "message": "value 99 is out of range (must be 0-7)",
  "details": "See RFC-5424 for valid priority levels",
  "field": "Priority"
}
```

### Extended Columns

```bash
# See all 25+ columns
curl "http://localhost:8000/logs?limit=1" | jq .rows[0]

# Filter by SysLogTag
curl "http://localhost:8000/logs?SysLogTag=nginx&limit=5"
```

## Configuration

### Enable API Key

Edit `docker-compose.yml`:

```yaml
environment:
  - API_KEY=test123456789
```

Restart:
```bash
docker-compose down && docker-compose up -d
```

Then use:
```bash
curl -H "X-API-Key: test123456789" "http://localhost:8000/logs"
```

### Change Port

Edit `docker-compose.yml`:

```yaml
ports:
  - "9000:8000"  # Host:Container
```

### Adjust Log Generation

Edit `docker/log-generator.sh`:

```bash
INTERVAL=10        # Seconds between bursts
LOGS_PER_BURST=3   # Logs per burst
```

Rebuild:
```bash
docker-compose down
docker-compose up -d --build
```

## Monitoring

```bash
# Container logs
docker-compose logs -f

# API logs
docker exec rsyslox-test cat /var/log/rsyslox.log

# Generator logs
docker exec rsyslox-test cat /var/log/log-generator.log

# Database count
docker exec rsyslox-test mysql Syslog -e "SELECT COUNT(*) FROM SystemEvents"
```

## Database Access

```bash
# Connect to MySQL
docker exec -it rsyslox-test mysql -u rsyslog -ppassword Syslog

# Example queries
SELECT ReceivedAt, FromHost, Priority, Message 
FROM SystemEvents 
ORDER BY ReceivedAt DESC 
LIMIT 10;
```

## Cleanup

```bash
# Stop (data persists)
docker-compose stop

# Remove (data persists)
docker-compose down

# Delete everything including data
docker-compose down -v
```

## Troubleshooting

**Binary not found:**
```bash
cd .. && make build-static
cd docker && docker-compose up -d
```

**Live logs not generating:**
```bash
# Check generator status
docker exec rsyslox-test ps aux | grep log-generator

# Restart container
docker-compose restart
```

**Container won't start:**
```bash
# Check logs
docker-compose logs

# Rebuild
docker-compose down
docker-compose up -d --build
```

## Advanced Usage

### Custom Test Data

```bash
# Connect to container
docker exec -it rsyslox-test bash

# Insert custom log
mysql Syslog <<EOF
INSERT INTO SystemEvents (ReceivedAt, FromHost, Priority, Message, SysLogTag)
VALUES (NOW(), 'testhost', 3, 'Custom test message', 'test');
EOF

# Exit container
exit
```

### Performance Testing

```bash
# Install apache bench
sudo apt-get install apache2-utils

# Test health endpoint
ab -n 1000 -c 10 http://localhost:8000/health

# Test logs endpoint (requires API key if enabled)
ab -n 100 -c 5 -H "X-API-Key: test123456789" \
  "http://localhost:8000/logs?limit=10"
```

### Export Test Data

```bash
# Export logs to JSON
docker exec rsyslox-test mysql Syslog -e \
  "SELECT * FROM SystemEvents LIMIT 100" --batch --raw | \
  jq -Rs 'split("\n") | map(split("\t")) | .[0] as $headers | 
  .[1:] | map(select(length > 0) | [($headers | to_entries | map(.value)), .] | 
  transpose | map({key: .[0], value: .[1]}) | from_entries)'
```

## Docker Compose Reference

### Full docker-compose.yml

```yaml
services:
  rsyslog:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: rsyslox-test
    ports:
      - "8000:8000"
    volumes:
      - ../build:/host-build:ro
    environment:
      - SERVER_PORT=8000
      - ALLOWED_ORIGINS=*
      # Uncomment to enable API key authentication:
      # - API_KEY=test123456789
    stdin_open: true
    tty: true
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8000 | API server port |
| ALLOWED_ORIGINS | * | CORS allowed origins |
| API_KEY | (none) | API authentication key |

## Integration Testing

### Automated Tests

Create a test script:

```bash
#!/bin/bash
# test-api.sh

API_URL="http://localhost:8000"

echo "Testing health endpoint..."
curl -sf "$API_URL/health" || exit 1

echo "Testing logs endpoint..."
curl -sf "$API_URL/logs?limit=1" || exit 1

echo "Testing meta endpoint..."
curl -sf "$API_URL/meta/FromHost" || exit 1

echo "All tests passed!"
```

Run tests:
```bash
chmod +x test-api.sh
./test-api.sh
```

## Next Steps

- [API Examples](../api/examples.md) - More API usage examples
- [Contributing](contributing.md) - Contribute to the project
- [Deployment](../guides/deployment.md) - Deploy to production
