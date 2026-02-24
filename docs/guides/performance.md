# Performance Guide

## Performance Overview

Typical response times with 1 M+ log entries:

| Operation | Response Time |
|---|---|
| Health check | < 5 ms |
| Simple query (limit 10, indexed fields) | 10–50 ms |
| Complex multi-filter query | 50–200 ms |
| Metadata query (`/api/meta/{column}`) | 100–500 ms |

## Database Indexes

rsyslox queries `SystemEvents` on `ReceivedAt`, `FromHost`, `Priority`, and `Facility`. Ensure these indexes exist:

```sql
CREATE INDEX IF NOT EXISTS idx_receivedat ON SystemEvents(ReceivedAt);
CREATE INDEX IF NOT EXISTS idx_fromhost   ON SystemEvents(FromHost);
CREATE INDEX IF NOT EXISTS idx_priority   ON SystemEvents(Priority);
CREATE INDEX IF NOT EXISTS idx_facility   ON SystemEvents(Facility);
CREATE INDEX IF NOT EXISTS idx_syslogtag  ON SystemEvents(SysLogTag);
```

Check existing indexes:
```sql
SHOW INDEX FROM SystemEvents;
```

## Query Optimization

**Fast — uses indexes:**
```bash
?Severity=3&start_date=2026-02-23T00:00:00Z&limit=100
?FromHost=web01&Severity=3
```

**Slower — full-text scan:**
```bash
?Message=keyword   # avoids this for large datasets without full-text index
```

**Best practice:** Combine indexed fields and use narrow time windows.

## MySQL Configuration

For large log volumes, tune `my.cnf`:

```ini
[mysqld]
innodb_buffer_pool_size = 2G    # Set to ~70% of RAM if dedicated to MySQL
innodb_log_file_size    = 512M
max_connections         = 200
```

## API Usage Tips

**Pagination — use reasonable page sizes:**
```bash
# Good: paginated
?limit=100&offset=0

# Avoid: huge single requests
?limit=50000   # Only use with show-all mode, sparingly
```

**Time windows — narrow reduces scan range significantly:**
```bash
# Fast: 1 hour
?start_date=2026-02-23T09:00:00Z&end_date=2026-02-23T10:00:00Z

# Slow: 30 days (scans entire table)
?start_date=2026-01-23T00:00:00Z
```

**Metadata — safe to cache:**
```bash
# Hosts and tags change slowly; cache /api/meta responses for 5–60 minutes
curl ".../api/meta/FromHost"
```

## Benchmarking

```bash
# Install wrk
sudo apt-get install wrk

# Benchmark health endpoint
wrk -t4 -c100 -d30s http://localhost:8000/health

# Benchmark log endpoint
wrk -t4 -c50 -d30s \
    -H "X-API-Key: $API_KEY" \
    "http://localhost:8000/api/logs?limit=10"
```

## Scaling

**Vertical:** More RAM for MySQL buffer pool, SSD for fast I/O — most effective for read-heavy syslog workloads.

**Horizontal:** rsyslox is stateless — run multiple instances behind a load balancer pointing to the same MySQL read replica.

## Troubleshooting

**Slow meta queries:** Run `ANALYZE TABLE SystemEvents` to refresh statistics.

**High memory usage:** Check `innodb_buffer_pool_size` and reduce if MySQL is sharing the machine.

**Timeouts under load:** Reduce connection pool pressure by limiting concurrent API consumers.
