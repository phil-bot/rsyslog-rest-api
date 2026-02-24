# Development Guide

## Prerequisites

- Go 1.21+
- Node.js 18+
- MySQL / MariaDB (for a real database) or the Docker test container

## Project Structure

```
rsyslox/
├── main.go                 # Entry point; embeds frontend and docs
├── embed.go                # go:embed directives
├── internal/
│   ├── auth/               # Session tokens, bcrypt, API key verification
│   ├── cleanup/            # Disk-based log retention service
│   ├── config/             # TOML config: load, save, validate, encrypt
│   ├── database/           # MySQL connection and query layer
│   └── server/             # HTTP server, routes, handlers, setup wizard
├── frontend/
│   ├── src/
│   │   ├── api/            # API client (fetch wrapper, auth headers)
│   │   ├── assets/         # Global CSS, CSS variables
│   │   ├── components/     # AppHeader, FilterPanel, LogTable, LogDetail
│   │   ├── composables/    # useLocale (i18n), useClickOutside
│   │   ├── i18n/           # en.json, de.json — all UI strings
│   │   ├── router/         # Vue Router — auth guard, setup detection
│   │   ├── stores/         # logs.js, auth.js, preferences.js
│   │   └── views/          # LogsView, AdminView, LoginView, SetupView
│   └── dist/               # Built output (embedded into binary)
├── docs/
│   ├── api-ui/             # openapi.yaml + Redoc HTML (embedded into binary)
│   ├── development/        # changelog.md, this file
│   └── operations/         # operations.md
└── scripts/
    └── install.sh          # Installer / uninstaller
```

## Running Locally

### Backend

Copy the example config and fill in your local database credentials:

```bash
cp config.dev.toml.example config.dev.toml
# Edit config.dev.toml — set database credentials and generate an admin hash:
go run . hash-password yourpassword
```

Then run the backend:

```bash
export RSYSLOX_CONFIG=./config.dev.toml
go run .
# or: make dev
```

The backend serves on the port defined in the config (default 8000). If the config file does not exist, it starts in setup wizard mode.

### Frontend

The frontend dev server proxies API requests to the Go backend:

```bash
cd frontend
npm install
npm run dev   # starts on http://localhost:5173
```

Vite is configured to forward `/api/*` and `/health` to `http://localhost:8000`.

### Building

```bash
# 1. Build the frontend
cd frontend && npm run build && cd ..

# 2. Build the binary (frontend/dist/ is embedded at compile time)
go build -ldflags "-X main.Version=0.4.0" -o rsyslox .
```

### Docker Test Container

A Docker Compose setup is available for testing against a real syslog database:

```bash
cd docker && docker compose up
```

The entrypoint generates a `config.toml` from environment variables and starts rsyslox. See `docker/docker-compose.yml` for variables.

---

## Frontend Architecture

The frontend is a Vue 3 + Vite single-page application using the Composition API throughout. State is managed in plain reactive modules (`stores/`), not Pinia.

### State Stores

**`stores/logs.js`** — central log state (entries, filters, pagination, selection, auto-refresh). All filter changes trigger a `resetPage()` + `fetchLogs()` via a single `watch`. Page changes trigger `fetchLogs()` via an arrow-wrapped watcher to prevent the page number being passed as the `fromRefresh` argument.

**`stores/auth.js`** — session token and role in `sessionStorage`. The API client reads `rsyslox_token` directly from `sessionStorage` to avoid circular imports.

**`stores/preferences.js`** — language, time format, font size, auto-refresh interval in `localStorage`. Exports reactive refs directly; consumers import and use them without calling a setup function. Font size is applied immediately to `document.documentElement.style.fontSize` on load and on every change.

### i18n

Translation keys live in `src/i18n/en.json` and `src/i18n/de.json`. The `useLocale` composable provides:

- `t(key, vars?)` — returns the translated string; falls back to English; interpolates `{name}`-style variables
- `fmtNumber(n)` — formats numbers with locale-appropriate thousands separators

The active language is the `language` ref from `stores/preferences.js`.

### Dynamic Table Sizing

`LogsView.computePageSize()` is called by a `ResizeObserver` on the `.logs-main` container and on font size preference changes. It:

1. Measures `toolbar`, `thead`, and `pagination` heights from the live DOM
2. Derives natural row height from the current `font-size` of `<html>` (set by font size preference)
3. Calculates `n = floor(available / naturalRowH) - 1` (minus one to prevent the last row being clipped by pagination)
4. Sets `exactRowH = floor(available / n)` and pushes it as `--row-h` to the `.table-scroll` element
5. Calls `setPageSize(n)` + `fetchLogs()` if `n` has changed

Rows use `height: var(--row-h)` so they stretch to fill the container exactly with no whitespace gap.

### Flash Animation

`fetchLogs(fromRefresh = false)` compares incoming row IDs against the previous set. Only IDs absent from the previous set receive the `row-new` class, which triggers a `row-flash` keyframe animation. The `fromRefresh` flag is `true` only when called from the auto-refresh timer — never from page changes, filter changes, or manual reloads.

---

## Backend Architecture

### Config

`internal/config` handles loading (`config.Load()`), saving (`config.Save()`), and validation. If the config file does not exist, `Load()` returns `setupMode = true` — the server then mounts only the setup routes. The setup handler calls `config.Save()` after writing the wizard result, and the frontend reloads into normal mode.

The database password is encrypted with AES-GCM (`internal/config/crypto.go`). The encryption key is derived from `/etc/machine-id` using SHA-256. Passwords with an `enc:` prefix are decrypted when building the DSN; plain passwords (during initial setup) are encrypted before being saved.

### Auth

Admin sessions use a random 32-byte token stored in an in-memory map with expiry. Tokens are transmitted via `X-Session-Token`. Read-only API keys are verified by computing SHA-256 of the submitted value and comparing against stored hashes in the config.

### Cleanup

`internal/cleanup` runs as a goroutine. When enabled, it periodically checks disk usage at the configured path using `statvfs`. If usage exceeds the threshold, it deletes the oldest `batch_size` rows from the syslog table and repeats until usage drops below the threshold or no more rows remain.

---

## Adding a Translation Key

1. Add the key and English value to `src/i18n/en.json`
2. Add the German translation to `src/i18n/de.json`
3. Use `t('your.key')` in the component

Keys use dot notation with a section prefix (`nav.`, `filter.`, `table.`, `admin.`, `prefs.`).

## Release

Releases are built by the GitHub Actions workflow on tag push (`v*`). The workflow:

1. Builds the frontend
2. Builds `linux/amd64` and `linux/arm64` binaries with `-ldflags "-X main.Version=<tag>"`
3. Creates an offline package (binary + install script + docs)
4. Publishes a GitHub Release with SHA-256 checksums

Pre-release tags (e.g. `v1.1.0-beta`) are automatically marked as pre-release.
