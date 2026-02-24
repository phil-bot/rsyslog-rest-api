# API Reference

Complete API documentation for rsyslox v0.4.0.

## Authentication

Two authentication methods are available depending on the use case:

| Method | Header | Access |
|---|---|---|
| Admin session token | `X-Session-Token: <token>` | Full access to all endpoints |
| Read-only API key | `X-API-Key: <key>` | `/api/logs` and `/api/meta` only |

**Obtain an admin session token:**
```bash
curl -X POST http://localhost:8000/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"password": "your-admin-password"}'
# Response: {"token": "<session-token>"}
```

**Create a read-only API key:** Admin panel → API Keys → Create. Keys are shown once in plaintext at creation time.

**Example request:**
```bash
curl -H "X-API-Key: your-key" "http://localhost:8000/api/logs?limit=5"
```

## Base URL

```
http://localhost:8000
```

Or with your configured host/port.

---

## Priority vs. Severity

rsyslog changed how it populates the `Priority` database column depending on version:

| rsyslog version | `Priority` column contains |
|---|---|
| < 8.2204.0 (legacy) | Severity only (0–7) |
| ≥ 8.2204.0 (modern) | RFC PRI = `Facility × 8 + Severity` |

The API detects the storage format automatically at startup by sampling the oldest and newest non-kernel entries. Mixed datasets (produced by a rsyslog upgrade) are handled correctly on a per-row basis.

**API response fields are always RFC-5424 compliant**, regardless of rsyslog version:

| Field | Description | Example |
|---|---|---|
| `Priority` | RFC PRI value (`Facility × 8 + Severity`) | `25` |
| `Severity` | Severity value 0–7 | `1` |
| `Severity_Label` | Human-readable severity | `"Alert"` |
| `Facility` | Facility value 0–23 | `3` |
| `Facility_Label` | Human-readable facility | `"daemon"` |

---

## Endpoints

### GET /health

Health check — no authentication required.

```bash
curl http://localhost:8000/health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.4.0",
  "timestamp": "2026-02-23T10:30:00Z"
}
```

**Response (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "database": "disconnected",
  "version": "v0.4.0",
  "timestamp": "2026-02-23T10:30:00Z"
}
```

---

### GET /api/logs

Retrieve log entries with filtering and pagination.

**Query Parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `offset` | Integer | 0 | Skip N entries |
| `limit` | Integer | 10 | Max results (max: 50 000 in show-all mode) |
| `start_date` | DateTime | −24 h | Start datetime (ISO 8601 / RFC 3339) |
| `end_date` | DateTime | now | End datetime (ISO 8601 / RFC 3339) |
| `FromHost` | String | — | Filter by hostname (repeatable) |
| `Severity` | Integer | — | Filter by severity 0–7 (repeatable) |
| `Priority` | Integer | — | Deprecated alias for `Severity` |
| `Facility` | Integer | — | Filter by facility 0–23 (repeatable) |
| `Message` | String | — | Text search in message field (repeatable = OR) |
| `SysLogTag` | String | — | Filter by syslog tag (repeatable) |

**Repeatable parameters** — repeat to filter by multiple values (OR logic):
```
?Severity=3&Severity=4
?FromHost=web01&FromHost=web02
```

**Severity Values (RFC-5424):**

| Value | Label | Description |
|---|---|---|
| 0 | Emergency | System unusable |
| 1 | Alert | Immediate action required |
| 2 | Critical | Critical conditions |
| 3 | Error | Error conditions |
| 4 | Warning | Warning conditions |
| 5 | Notice | Normal but significant |
| 6 | Informational | Informational messages |
| 7 | Debug | Debug messages |

**Example requests:**
```bash
# Latest 10 logs
curl -H "X-API-Key: $KEY" "http://localhost:8000/api/logs?limit=10"

# Errors from last hour
START=$(date -u -d "1 hour ago" "+%Y-%m-%dT%H:%M:%SZ")
curl -H "X-API-Key: $KEY" "http://localhost:8000/api/logs?Severity=3&start_date=$START"

# Errors and warnings from multiple hosts
curl -H "X-API-Key: $KEY" \
  "http://localhost:8000/api/logs?FromHost=web01&FromHost=web02&Severity=3&Severity=4"
```

**Response (200 OK):**
```json
{
  "total": 1234,
  "offset": 0,
  "limit": 10,
  "rows": [
    {
      "ID": 12345,
      "ReceivedAt": "2026-02-23T10:30:15Z",
      "DeviceReportedTime": "2026-02-23T10:30:13Z",
      "Facility": 3,
      "Facility_Label": "daemon",
      "Priority": 25,
      "Severity": 1,
      "Severity_Label": "Alert",
      "FromHost": "webserver01",
      "Message": "Connection timeout to database",
      "SysLogTag": "nginx",
      "CustomerID": 42,
      "EventSource": "web-service",
      "EventUser": "www-data",
      "EventID": 504
    }
  ]
}
```

Core fields always present: `ID`, `ReceivedAt`, `FromHost`, `Priority`, `Severity`, `Severity_Label`, `Facility`, `Facility_Label`, `Message`.
Extended fields (25+ total) populated when available: `CustomerID`, `DeviceReportedTime`, `SysLogTag`, `EventSource`, `EventUser`, `EventID`, `EventCategory`, `NTSeverity`, `Importance`, `SystemID`, `InfoUnitID`.

---

### GET /api/meta

List all available column names.

```bash
curl -H "X-API-Key: $KEY" "http://localhost:8000/api/meta"
```

**Response:**
```json
{
  "available_columns": [
    "ID", "CustomerID", "ReceivedAt", "DeviceReportedTime",
    "Facility", "Priority", "FromHost", "Message", "NTSeverity",
    "Importance", "EventSource", "EventUser", "EventCategory",
    "EventID", "SysLogTag", "InfoUnitID", "SystemID", "Severity"
  ],
  "usage": "GET /api/meta/{column} to get distinct values for a column"
}
```

?> **Note:** `Severity` is a virtual column computed from `Priority MOD 8` — it is not a physical database column.

---

### GET /api/meta/{column}

Get distinct values for a column. No default time filter is applied — without parameters, returns values from the **entire dataset**.

**Query Parameters:** Same as `/api/logs` (all optional).

**Examples:**
```bash
# All distinct hosts (entire dataset)
curl -H "X-API-Key: $KEY" "http://localhost:8000/api/meta/FromHost"

# Severity levels with labels
curl -H "X-API-Key: $KEY" "http://localhost:8000/api/meta/Severity"

# Hosts that had errors in a specific window
curl -H "X-API-Key: $KEY" \
  "http://localhost:8000/api/meta/FromHost?Severity=3&start_date=2026-02-01T00:00:00Z"
```

**Response for `Severity` or `Facility` (with labels):**
```json
[
  {"val": 3, "label": "Error"},
  {"val": 4, "label": "Warning"},
  {"val": 6, "label": "Informational"}
]
```

**Response for string columns:**
```json
["webserver01", "webserver02", "dbserver01"]
```

**Response for integer columns:**
```json
[1, 2, 5, 10]
```

---

### POST /api/admin/login

Obtain an admin session token.

```bash
curl -X POST http://localhost:8000/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"password": "your-admin-password"}'
```

**Response (200 OK):**
```json
{"token": "<session-token>"}
```

Use the token in subsequent requests:
```bash
curl -H "X-Session-Token: <token>" "http://localhost:8000/api/admin/config"
```

---

## HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | OK |
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — missing or invalid credentials |
| 500 | Internal Server Error |
| 503 | Service Unavailable — database unreachable |

## Rate Limiting

No built-in rate limiting. Use a reverse proxy (nginx/Apache) for rate limiting in production — see [Deployment Guide](../guides/deployment.md).

## What's New in v0.4.0

- ✅ Web UI — full-featured log viewer embedded in the binary
- ✅ Admin session token authentication (`X-Session-Token`)
- ✅ Named, revocable read-only API keys (replaces single `API_KEY`)
- ✅ API paths prefixed with `/api/` (e.g. `/api/logs`, `/api/meta`)
- ✅ Interactive API documentation at `/docs`
- ✅ TOML configuration — no more `.env` file

[View Full Changelog](../development/changelog.md)
