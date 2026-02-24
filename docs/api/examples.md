# API Examples

Practical examples for using the rsyslox API.

## Setup

All examples assume you have created a read-only API key in **Admin → API Keys**.

```bash
API_KEY="your-key-here"
BASE_URL="http://localhost:8000"
```

---

## Health Check

```bash
curl "$BASE_URL/health"
```

Response:
```json
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.4.0",
  "timestamp": "2026-02-23T10:30:00Z"
}
```

---

## Log Queries

### Recent Logs

```bash
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?limit=10"
```

### Filter by Severity

```bash
# Errors only (severity 3)
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?Severity=3&limit=50"

# Errors and Critical
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?Severity=2&Severity=3"
```

### Filter by Host

```bash
# Single host
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?FromHost=web01"

# Multiple hosts
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/logs?FromHost=web01&FromHost=web02&FromHost=db01"
```

### Filter by Date Range

```bash
# Last hour
START=$(date -u -d "1 hour ago" "+%Y-%m-%dT%H:%M:%SZ")
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?start_date=$START"

# Specific window
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/logs?start_date=2026-02-23T08:00:00Z&end_date=2026-02-23T12:00:00Z"
```

### Combined Filters

```bash
# Errors from specific hosts in last 24h
START=$(date -u -d "24 hours ago" "+%Y-%m-%dT%H:%M:%SZ")
curl -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/logs?FromHost=web01&FromHost=web02&Severity=3&Severity=4&start_date=$START"
```

### Message Search

```bash
# Search for "failed"
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?Message=failed"

# Multiple search terms (OR)
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?Message=error&Message=failed"
```

### Pagination

```bash
# First page
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?limit=10&offset=0"

# Second page
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?limit=10&offset=10"
```

---

## Metadata Queries

### Get All Hosts

```bash
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/meta/FromHost"
```

Response:
```json
["web01", "web02", "db01", "app01"]
```

### Get Severity Levels

```bash
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/meta/Severity"
```

Response:
```json
[
  {"val": 3, "label": "Error"},
  {"val": 4, "label": "Warning"},
  {"val": 6, "label": "Informational"}
]
```

### Filtered Metadata

```bash
# Hosts that had errors
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/meta/FromHost?Severity=3&Severity=4"

# Tags used on a specific host
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/meta/SysLogTag?FromHost=web01"
```

---

## Script Examples

### Bash

```bash
#!/bin/bash
API_KEY="your-key-here"
BASE_URL="http://localhost:8000"

# Get recent errors
curl -s -H "X-API-Key: $API_KEY" \
  "$BASE_URL/api/logs?Severity=3&limit=100" | jq '.rows[] | {time: .ReceivedAt, host: .FromHost, msg: .Message}'
```

### Python

```python
import requests

API_KEY = "your-key-here"
BASE_URL = "http://localhost:8000"
headers = {"X-API-Key": API_KEY}

# Get logs
response = requests.get(
    f"{BASE_URL}/api/logs",
    headers=headers,
    params={"Severity": 3, "limit": 100}
)

data = response.json()
print(f"Total: {data['total']}")
for log in data["rows"]:
    print(f"{log['ReceivedAt']} [{log['Severity_Label']}] {log['FromHost']}: {log['Message']}")
```

### Node.js

```javascript
const axios = require('axios');

const API_KEY = 'your-key-here';
const BASE_URL = 'http://localhost:8000';

async function getLogs() {
  const response = await axios.get(`${BASE_URL}/api/logs`, {
    headers: { 'X-API-Key': API_KEY },
    params: { Severity: 3, limit: 100 }
  });

  const { total, rows } = response.data;
  console.log('Total:', total);
  rows.forEach(log => {
    console.log(`${log.ReceivedAt} [${log.Severity_Label}] ${log.FromHost}: ${log.Message}`);
  });
}

getLogs().catch(console.error);
```

---

## Error Responses

```bash
# Invalid severity value
curl -H "X-API-Key: $API_KEY" "$BASE_URL/api/logs?Severity=99"
```

Response:
```json
{
  "code": "INVALID_SEVERITY",
  "message": "value 99 is out of range (must be 0-7)",
  "details": "See RFC-5424 for valid severity levels",
  "field": "Severity"
}
```

## Tips & Best Practices

- Always set a `limit` on `/api/logs` to avoid large result sets
- Use narrow time windows for faster queries — combine `start_date` + `end_date`
- Cache `/api/meta` responses — they cover the full dataset and change slowly
- Use `/health` to verify the service is reachable before querying
- Filter on indexed fields (`Severity`, `FromHost`, `start_date`) for best performance

## More Resources

- [API Reference](reference.md) — Complete endpoint documentation
- [Troubleshooting](../guides/troubleshooting.md) — Common issues and filter mistakes
- [Security Guide](../guides/security.md) — Best practices
