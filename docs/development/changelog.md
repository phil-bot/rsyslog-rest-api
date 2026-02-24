# Changelog

All notable changes to rsyslox.

## [Unreleased]

---

## [v0.4.0] - 2026-02-23

This release introduces the complete web frontend and the browser-based preferences
system, replacing the previous API-only interface. All changes previously tracked
under the unreleased `v1.0.0` label are included here.

### Added

**Web Frontend (Vue 3 + Vite)**
- Full-featured log viewer embedded in the binary via `go:embed`
- Dark/light theme with system preference detection and manual toggle
- Responsive layout: sidebar panel on desktop, slide-over modal on mobile
- Skeleton loading states for first render

**Log Viewer**
- Filter panel: time range (relative quick-select or absolute date/time), severity,
  facility, host, tag, and free-text message search
- Time range selector redesigned as a compact segmented control — no individual pill
  borders, active segment filled with primary colour
- Date fields pre-filled on first render using the default duration (`1h`)
- Time shift buttons (`‹ Earlier` / `Later ›`) to step through log windows
- Log table with severity colour coding, monospace data columns, multi-row selection
- Detail panel: full message, all fields, expandable raw JSON, copy-to-clipboard
- Export selected or all visible rows as CSV or JSON (client-side, no server round-trip)
- Auto-refresh with countdown display; interval configurable from browser preferences
- Dynamic page size — number of rows computed from available viewport height,
  updates on window resize and font size change; rows stretch to fill the container
  exactly with no whitespace gap below the last entry

**Admin Panel**
- Server settings editor (CORS origins, SSL toggle)
- Database info view (read-only; password always masked)
- Log cleanup configuration (disk path, threshold %, batch size, check interval)
- Read-only API key management: create named keys, list, revoke
- One-time key reveal after creation with copy button (plaintext never stored or logged)
- Preferences tab (default landing page) — see below

**Internationalisation (i18n)**
- Translation files `src/i18n/en.json` and `src/i18n/de.json`; all UI strings externalised
- `useLocale` composable — reactive `t(key, vars?)` with variable interpolation;
  `fmtNumber()` for locale-aware thousands separators (`.` in German, `,` in English)
- All views and components fully translated (log viewer, admin panel, filter panel)
- Tab labels in Admin panel re-render immediately on language change

**User Preferences (browser-persisted)**
- `preferences.js` store — persists `language`, `timeFormat`, `fontSize`,
  `autoRefreshInterval` in `localStorage`; applied immediately, no restart needed
- Language: English / Deutsch
- Time format: 24-hour or 12-hour clock applied to the timestamp column
- Font size: Small (13 px) / Medium (14 px) / Large (15 px) — sets `font-size` on
  `<html>`, all `rem` values scale proportionally including row height
- Auto-refresh interval (seconds); replaces the former server-side setting

**Configuration (TOML)**
- `/etc/rsyslox/config.toml` replaces `.env`
- AES-GCM encrypted database password (key derived from machine-id)
- bcrypt admin password (cost 12)
- SHA-256 hashed read-only API keys
- First-run setup wizard (localhost-only until configured)

**Install & Operations**
- `scripts/install.sh` — guided installer: system user, binary, systemd unit, setup wizard
- `--uninstall` flag for clean removal
- Systemd hardening: `NoNewPrivileges`, `ProtectSystem=strict`, `PrivateTmp`
- Offline API documentation via Redoc (embedded in binary at `/docs`)

**CI/CD**
- GitHub Actions CI: frontend build → artifact → Go build + vet + test
- GitHub Actions Release: multi-arch binaries (amd64/arm64), offline package, SHA-256 checksums
- Pre-release auto-detection from tag name (e.g. `v0.5.0-beta`)

### Changed

- Single binary embeds `frontend/dist/` and `docs/api-ui/` via `go:embed`
- `database.Connect()` uses `cfg.DSN()` from TOML config instead of env vars
- `go.mod`: added `BurntSushi/toml`, `golang.org/x/crypto`
- `docker/entrypoint.sh` generates `config.toml` from env vars for test containers
- Systemd unit updated: new install paths, security hardening directives
- Settings gear icon in header replaces the "Admin" nav link; right-aligned with active state

### Fixed

- **Flash animation on page change** — `watch(page, fetchLogs)` passed the page
  number as the first positional argument to `fetchLogs`, making `fromRefresh` truthy;
  every row on the new page flashed as "new". Fixed with an arrow wrapper:
  `watch(page, () => fetchLogs())`
- **Impossible entry count display** (e.g. `3,660 von 2,919`) — a lazy `dbTotal`
  background fetch used no time filter and returned a different total than the active
  view; removed entirely, only the filtered count is shown
- **App.vue broken theme injection** — `import { provide }` appeared after `const`
  declarations; moved to the top import line, fixing the white screen on `/logs` and
  the broken dark/light toggle
- **FilterPanel late imports** — `import { computed }` and `import { useLocale }`
  placed after `const SEV_COLORS`; moved to top of `<script setup>`
- **AdminView stray `<`** — a lone `<` left in the template after removing a field
  caused the Vue template compiler to fail
- **LogsView missing `watch` import** — `watch` was used but absent from the Vue
  named imports

### Removed

- `.env` / `.env.example` — replaced by TOML config
- `API_KEY` env var — replaced by named, revocable read-only API keys
- `dbTotal` ref and its background API call from `logs.js`
- `autoRefresh` and `autoRefreshInterval` props from `AppHeader` (unused in template)
- `@toggle-refresh` event binding on `<AppHeader>` in `AdminView`
- `SEVERITY_LABELS` import in `LogTable` (imported but never referenced)
- `tableScrollRef` ref in `LogTable` (declared, never read)
- Duplicate `position: sticky` on `thead th` (superseded by `thead tr`)
- Redundant `.pill-row.wrap` CSS rule (identical to `.pill-row`)
- `api.getConfig()` call in `LogsView.onMounted` — race condition with
  `watch(prefAutoRefresh, { immediate: true })`; preferences store is now sole source

---

## [v0.3.0] - 2025-02-19

### Added
- **`Severity` field** in all `/logs` responses — RFC-5424 compliant (0–7)
- **`Severity_Label` field** in all `/logs` responses
- **`Priority` field** now contains true RFC PRI value (`Facility × 8 + Severity`)
- **Automatic rsyslog version detection** at startup
- **Per-row format detection** for mixed datasets
- **`?Severity=` filter parameter** — `?Priority=` kept as deprecated alias
- **`/meta/Severity` virtual column** — computed via `Priority MOD 8`
- **Cleanup service** — disk-based log retention

### Changed
- `Priority_Label` removed from responses
- `?Priority=` is a deprecated alias for `?Severity=`
- `/meta/{column}` no longer applies a default time filter

---

## [v0.2.3] - 2025-02-15

### Added
- Structured error responses (`code`, `message`, `details`, `field`)
- Enhanced multi-value filter performance
- Improved meta endpoint filtering

---

## [v0.2.2] - 2025-02-09

### Added
- Multi-value filters for all parameters
- All 25+ SystemEvents columns in responses
- Live log generator for Docker testing

---

## [v0.2.1] - 2025-01-15

### Fixed
- Database connection timeout
- Memory leak in queries
- CORS preflight handling

---

## [v0.2.0] - 2024-12-20

### Added
- RFC-5424 labels, meta endpoint, SSL/TLS support, CORS configuration

---

## [v0.1.0] - 2024-10-01

Initial release.
