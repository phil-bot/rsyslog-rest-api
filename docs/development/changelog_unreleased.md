## [Unreleased]

### Added

- **`db_total` field in `/api/logs` response** — total entry count in `SystemEvents`
  (no filter applied), returned alongside the existing filtered `total` on every request
- **`db_total` and `oldest_entry` in `GET /api/meta` response** — database-level stats
  (total row count, timestamp of oldest entry) without a separate API call
- **`internal/database/cache.go`** — TTL cache (60 s) for `QueryDistinctValues` results;
  reduces redundant `SELECT DISTINCT` queries on poll-heavy setups
- **Toolbar DB total display** — log viewer toolbar shows
  `{filtered} entries · {db_total} total in DB` when a filter is active and the counts differ

### Changed

- **`/api/logs` runs three DB queries in parallel** (`CountLogs`, `QueryLogs`,
  `TotalCount`) via goroutines + `sync.WaitGroup`, reducing per-request latency
- **`QueryDistinctValues` results are cached** for 60 s per unique
  column + filter combination; subsequent identical meta requests are served from memory
- **Date range limit removed** — the 90-day hard cap on `start_date`/`end_date`
  has been dropped; arbitrarily large time windows are now accepted

### Fixed

- **Args slice mutation** — `QueryLogs` previously appended `LIMIT`/`OFFSET` directly
  to the caller's `args` slice; replaced with an internal copy to prevent data races
  when running queries concurrently in `QueryLogsWithTotal`
- **German strings in LogTable toolbar** — `ctrl-btn` title attributes were in German;
  replaced with English strings (consistent with UI language policy)
