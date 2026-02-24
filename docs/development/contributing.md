# Contributing Guidelines

Thank you for contributing to rsyslox!

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/rsyslox.git`
3. Create a feature branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- MySQL/MariaDB or Docker
- make

### Running Locally

**Backend:**
```bash
cp config.dev.toml.example config.dev.toml
# Edit config.dev.toml — fill in database credentials

RSYSLOX_CONFIG=./config.dev.toml go run .
# or: make dev
```

**Frontend (in a second terminal):**
```bash
cd frontend
npm install
npm run dev
# Vite dev server starts at http://localhost:5173
# Proxies API requests to localhost:8000
```

**Build:**
```bash
make all          # frontend + Redoc + Go binary
make build        # Go binary only (requires frontend/dist to exist)
make build-static # Static binary for production
```

### Docker (Optional)

A Docker test environment with MariaDB and live log generation is available — see [Docker Testing Environment](docker.md).

## Project Structure

```
rsyslox/
├── main.go                     # Entry point, CLI commands
├── embed.go                    # go:embed directives for frontend/dist and docs/api-ui
├── internal/
│   ├── auth/                   # Session token + API key verification
│   ├── cleanup/                # Disk-based log retention goroutine
│   ├── config/                 # TOML config load/save/validate, AES-GCM encryption
│   ├── database/               # MySQL connection, query helpers
│   └── server/                 # HTTP server, routing, handlers
├── frontend/
│   ├── src/
│   │   ├── api/                # API client (fetch wrappers)
│   │   ├── components/         # Vue components (AppHeader, LogTable, FilterPanel, …)
│   │   ├── composables/        # useLocale (i18n)
│   │   ├── i18n/               # en.json, de.json translation files
│   │   ├── router/             # Vue Router
│   │   ├── stores/             # Plain reactive stores: logs.js, auth.js, preferences.js
│   │   └── views/              # Page-level components: LogsView, AdminView
│   ├── package.json
│   └── vite.config.js
├── docs/                       # Docsify documentation (served via GitHub Pages)
├── docker/                     # Docker test environment
├── scripts/
│   └── install.sh              # Installer script
├── Makefile
└── rsyslox.service             # systemd unit file template
```

## Frontend Architecture

The frontend uses Vue 3 Composition API with plain reactive stores (no Pinia).

**Stores:**
- `stores/logs.js` — central log state, filter refs, watch-based reactivity, fetchLogs()
- `stores/auth.js` — session token in sessionStorage
- `stores/preferences.js` — language, timeFormat, fontSize, autoRefreshInterval in localStorage

**i18n:**
```javascript
import { useLocale } from '@/composables/useLocale'
const { t, fmtNumber } = useLocale()

t('filter.severity')              // → "Severity"
t('logs.showing', { n: 42 })      // → "Showing 42 entries"
fmtNumber(1234567)                 // → "1,234,567" (EN) or "1.234.567" (DE)
```

To add a translation key: add it to both `i18n/en.json` and `i18n/de.json`, then use `t('your.key')`.

## Coding Standards

- Go: run `go fmt ./...` and `go vet ./...` before committing
- Vue: keep components focused; use `<script setup>` syntax
- Commit messages: conventional commits format (see below)

## Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples:**
```
feat: add Facility filter to log viewer

Adds multi-select facility filter pills to FilterPanel,
wired to the existing facility filter in logs.js store.
```

```
fix: handle null SysLogTag in meta query

NULL values caused a panic in the distinct-value aggregation.
Filtered out with WHERE SysLogTag IS NOT NULL.

Fixes #42
```

## Branch Naming

```
feature/your-feature-name
fix/short-description
docs/what-you-are-documenting
refactor/what-you-are-refactoring
```

## Tests

```bash
# Run all Go tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/config/
```

## Pull Request Checklist

- [ ] `go fmt ./...` and `go vet ./...` pass
- [ ] All Go tests pass: `go test ./...`
- [ ] Frontend builds without errors: `cd frontend && npm run build`
- [ ] New translation keys added to both `en.json` and `de.json`
- [ ] Changelog entry added to `docs/development/changelog.md` under `[Unreleased]`
- [ ] Documentation updated if behaviour changed

## Release Process

For maintainers:

1. Move `[Unreleased]` entries to a new version section in `changelog.md`
2. Create and push a git tag: `git tag -a v0.X.Y -m "Release v0.X.Y" && git push origin v0.X.Y`
3. GitHub Actions builds amd64/arm64 binaries, creates an offline package, and publishes the release with SHA-256 checksums automatically

Pre-releases are detected automatically from the tag name (e.g. `v0.5.0-beta`).

## Reporting Bugs

Use [GitHub Issues](https://github.com/phil-bot/rsyslox/issues). Include:

- rsyslox version (`/health` endpoint → `version` field)
- OS and architecture
- Steps to reproduce
- Relevant log output: `sudo journalctl -u rsyslox -n 100`

## Feature Requests

Open a [GitHub Discussion](https://github.com/phil-bot/rsyslox/discussions) first to discuss the proposal before submitting a PR for larger features.

## Questions?

- [GitHub Discussions](https://github.com/phil-bot/rsyslox/discussions)
- [GitHub Issues](https://github.com/phil-bot/rsyslox/issues)
