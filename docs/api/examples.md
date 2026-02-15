# API Examples

Practical examples for using rsyslox API.

## Basic Usage

### Health Check

```bash
curl http://localhost:8000/health
```

Response:
```json
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.2.3",
  "timestamp": "2025-02-15T10:30:00Z"
}
```

### Retrieve Recent Logs

```bash
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?limit=10"
```

## Filtering Examples

### Filter by Host

```bash
# Single host
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?FromHost=web01&limit=20"

# Multiple hosts (Multi-value!)
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?FromHost=web01&FromHost=web02&FromHost=db01"
```

### Filter by Priority

```bash
# Errors only (Priority 3)
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?Priority=3&limit=50"

# Errors and Critical (Multiple priorities)
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?Priority=2&Priority=3"
```

### Filter by Date Range

```bash
# Last hour
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?start_date=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)"

# Specific time range
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?start_date=2025-02-15T08:00:00Z&end_date=2025-02-15T12:00:00Z"
```

### Combined Filters

```bash
# Errors from specific hosts in last 24h
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?FromHost=web01&FromHost=web02&Priority=3&Priority=4&start_date=$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%SZ)"
```

### Message Search

```bash
# Search for "failed" in messages
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?Message=failed&limit=20"

# Multiple search terms
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?Message=error&Message=failed"
```

## Metadata Queries

### Get All Hosts

```bash
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/meta/FromHost"
```

Response:
```json
{
  "column": "FromHost",
  "values": ["web01", "web02", "db01", "app01"]
}
```

### Get Priorities with Labels

```bash
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/meta/Priority"
```

Response:
```json
{
  "column": "Priority",
  "values": [
    {"val": 0, "label": "Emergency"},
    {"val": 1, "label": "Alert"},
    {"val": 2, "label": "Critical"},
    {"val": 3, "label": "Error"},
    {"val": 4, "label": "Warning"},
    {"val": 5, "label": "Notice"},
    {"val": 6, "label": "Informational"},
    {"val": 7, "label": "Debug"}
  ]
}
```

### Filtered Metadata

```bash
# Get hosts that had errors
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/meta/FromHost?Priority=3&Priority=4"
```

## Pagination

```bash
# First page (10 items)
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?limit=10&offset=0"

# Second page
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?limit=10&offset=10"

# Third page
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?limit=10&offset=20"
```

## Advanced Examples

### Monitor Specific Service

```bash
# nginx logs only
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?SysLogTag=nginx&limit=50"
```

### Security Audit

```bash
# Failed login attempts
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?SysLogTag=sshd&Message=Failed&Priority=4"
```

### Performance Monitoring

```bash
# Slow queries from database
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?FromHost=db01&Priority=4&Message=slow"
```

## Using with Scripts

### Bash Script

```bash
#!/bin/bash
API_KEY="your-api-key-here"
API_URL="http://localhost:8000"

# Get recent errors
curl -s -H "X-API-Key: $API_KEY" \
  "$API_URL/logs?Priority=3&limit=100" | jq '.rows[]'
```

### Python Script

```python
import requests
import json

API_KEY = "your-api-key-here"
API_URL = "http://localhost:8000"

headers = {"X-API-Key": API_KEY}

# Get logs
response = requests.get(
    f"{API_URL}/logs",
    headers=headers,
    params={"Priority": 3, "limit": 100}
)

logs = response.json()
for log in logs['rows']:
    print(f"{log['ReceivedAt']} - {log['FromHost']}: {log['Message']}")
```

### Node.js Script

```javascript
const axios = require('axios');

const API_KEY = 'your-api-key-here';
const API_URL = 'http://localhost:8000';

async function getLogs() {
  try {
    const response = await axios.get(`${API_URL}/logs`, {
      headers: { 'X-API-Key': API_KEY },
      params: { Priority: 3, limit: 100 }
    });
    
    console.log('Total logs:', response.data.total);
    response.data.rows.forEach(log => {
      console.log(`${log.ReceivedAt} - ${log.FromHost}: ${log.Message}`);
    });
  } catch (error) {
    console.error('Error:', error.message);
  }
}

getLogs();
```

## Error Handling

### Structured Errors (v0.2.3+)

```bash
# Invalid priority
curl -H "X-API-Key: YOUR_KEY" \
  "http://localhost:8000/logs?Priority=99"
```

Response:
```json
{
  "code": "INVALID_PRIORITY",
  "message": "value 99 is out of range (must be 0-7)",
  "details": "See RFC-5424 for valid priority levels",
  "field": "Priority"
}
```

## Tips & Best Practices

1. **Use pagination** for large result sets
2. **Combine filters** to narrow down results
3. **Use date ranges** to limit time scope
4. **Store API key securely** - never commit to git
5. **Use HTTPS** in production
6. **Implement rate limiting** in client code
7. **Cache metadata** queries (hosts, priorities)

## More Resources

- [API Reference](reference.md) - Complete API documentation
- [Troubleshooting](../guides/troubleshooting.md) - Common issues
- [Security Guide](../guides/security.md) - Best practices
