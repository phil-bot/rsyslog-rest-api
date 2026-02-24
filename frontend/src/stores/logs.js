import { ref, computed, watch } from 'vue'
import { api } from '@/api/client'

// Severity labels (RFC 5424)
export const SEVERITY_LABELS = {
  0: 'Emergency', 1: 'Alert', 2: 'Critical', 3: 'Error',
  4: 'Warning',   5: 'Notice', 6: 'Info',    7: 'Debug'
}

// Facility labels (RFC 5424)
export const FACILITY_LABELS = {
  0: 'kern', 1: 'user', 2: 'mail', 3: 'daemon',
  4: 'auth', 5: 'syslog', 6: 'lpr', 7: 'news',
  8: 'uucp', 9: 'cron', 10: 'authpriv', 11: 'ftp',
  16: 'local0', 17: 'local1', 18: 'local2', 19: 'local3',
  20: 'local4', 21: 'local5', 22: 'local6', 23: 'local7'
}

// --- Reactive state ---
const logs        = ref([])
const total       = ref(0)
const loading     = ref(false)
const error       = ref(null)

// Pagination
const page        = ref(1)
const pageSize    = ref(15)
const showAll     = ref(false)  // toggle: show all vs paginated

// Time filter
const timeMode    = ref('relative')   // 'relative' | 'absolute'
const relativeDur = ref('1h')         // '15m','1h','6h','24h','7d','30d'
const startDate   = ref('')
const endDate     = ref('')

// Filters
const severities  = ref([])           // selected severity values []int
const facilities  = ref([])
const hosts       = ref([])
const tags        = ref([])
const messageSearch = ref('')

// Selection
const selectedIds = ref(new Set())

// Detail panel
const detailEntry = ref(null)

// Available filter options (from /meta)
const availableHosts = ref([])
const availableTags        = ref([])
const availableSeverities  = ref([])   // [{val, label}] from /api/meta/Severity
const availableFacilities  = ref([])   // [{val, label}] from /api/meta/Facility

// Auto-refresh
const autoRefresh        = ref(true)
const newIds             = ref(new Set())  // IDs of freshly loaded rows → flash animation
const firstLoad          = ref(true)         // true until first successful fetch completes
const countdown          = ref(0)            // seconds until next auto-refresh
const autoRefreshInterval = ref(30)   // seconds
let   refreshTimer       = null
let   countdownTimer     = null

// --- Computed ---
const offset = computed(() => (page.value - 1) * pageSize.value)

const totalPages = computed(() =>
  showAll.value ? 1 : Math.max(1, Math.ceil(total.value / pageSize.value))
)

const selectedCount = computed(() => selectedIds.value.size)

const selectedLogs = computed(() =>
  logs.value.filter(l => selectedIds.value.has(l.ID))
)

// --- Helpers ---
function buildDateParams() {
  if (timeMode.value === 'absolute') {
    // datetime-local gives "YYYY-MM-DDTHH:MM" — append seconds + Z for RFC3339
    const toRFC = v => v ? (v.length === 16 ? v + ':00Z' : v.endsWith('Z') ? v : v + 'Z') : v
    return { start_date: toRFC(startDate.value), end_date: toRFC(endDate.value) }
  }
  const now  = new Date()
  const durations = {
    '15m': 15 * 60 * 1000,
    '1h':  60 * 60 * 1000,
    '6h':  6  * 60 * 60 * 1000,
    '24h': 24 * 60 * 60 * 1000,
    '7d':  7  * 24 * 60 * 60 * 1000,
    '30d': 30 * 24 * 60 * 60 * 1000,
  }
  const ms = durations[relativeDur.value] ?? durations['1h']
  const from = new Date(now - ms)
  // Backend requires RFC3339 format: 2025-02-15T10:00:00Z
  return {
    start_date: from.toISOString().slice(0, 19) + 'Z',
    end_date:   now.toISOString().slice(0, 19) + 'Z',
  }
}

function buildParams() {
  const params = {
    limit:  showAll.value ? 50000 : pageSize.value,
    offset: showAll.value ? 0     : offset.value,
    ...buildDateParams()
  }
  // Pass arrays — the API client serializes them as repeated params (?Severity=4&Severity=5)
  if (severities.value.length)  params['Severity']  = severities.value
  if (facilities.value.length)  params['Facility']  = facilities.value
  if (hosts.value.length)       params['FromHost']  = hosts.value
  if (tags.value.length)        params['SysLogTag'] = tags.value
  if (messageSearch.value.trim()) params['Message'] = messageSearch.value.trim()
  return params
}

// --- Actions ---
async function fetchLogs(fromRefresh = false) {
  loading.value = true
  error.value   = null
  try {
    const res  = await api.getLogs(buildParams())
    const rows = res.rows ?? []

    // Flash only on auto-refresh: highlights genuinely new incoming entries.
    // NOT on filter changes or manual reloads (those change the result set entirely).
    if (fromRefresh && logs.value.length > 0 && rows.length > 0) {
      const prevIds = new Set(logs.value.map(r => r.ID))
      const fresh   = new Set(rows.filter(r => !prevIds.has(r.ID)).map(r => r.ID))
      if (fresh.size > 0) {
        newIds.value = fresh
        setTimeout(() => { newIds.value = new Set() }, 1500)
      }
    }

    logs.value  = rows
    total.value = res.total ?? 0
    firstLoad.value = false
  } catch (e) {
    error.value = e.message || 'Failed to load logs'
  } finally {
    loading.value = false
  }
}

async function fetchFilterOptions() {
  try {
    const [hostsRes, tagsRes, sevRes, facRes] = await Promise.all([
      api.getMetaColumn('FromHost'),
      api.getMetaColumn('SysLogTag'),
      api.getMetaColumn('Severity'),
      api.getMetaColumn('Facility'),
    ])
    availableHosts.value       = hostsRes ?? []
    availableTags.value        = tagsRes  ?? []
    availableSeverities.value  = Array.isArray(sevRes) ? sevRes : []
    availableFacilities.value  = Array.isArray(facRes) ? facRes : []
  } catch { /* filter options are optional */ }
}

function setPage(n) {
  page.value = Math.min(Math.max(1, n), totalPages.value)
  newIds.value = new Set()  // never flash rows on page navigation
}

function resetPage() {
  page.value = 1
}

function toggleSelection(id) {
  const s = new Set(selectedIds.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  selectedIds.value = s
}

function toggleSelectAll() {
  if (selectedIds.value.size === logs.value.length) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(logs.value.map(l => l.ID))
  }
}

function clearSelection() {
  selectedIds.value = new Set()
}

function openDetail(entry) {
  detailEntry.value = entry
}

function closeDetail() {
  detailEntry.value = null
}

function startAutoRefresh(intervalSec) {
  stopAutoRefresh()
  if (!intervalSec) return
  autoRefresh.value         = true
  autoRefreshInterval.value = intervalSec
  countdown.value           = intervalSec

  refreshTimer = setInterval(() => {
    fetchLogs(true)   // fromRefresh=true → enables flash for new entries
    countdown.value = intervalSec
  }, intervalSec * 1000)

  // Tick every second
  countdownTimer = setInterval(() => {
    if (countdown.value > 0) countdown.value--
  }, 1000)
}

function stopAutoRefresh() {
  if (refreshTimer)    { clearInterval(refreshTimer);    refreshTimer    = null }
  if (countdownTimer)  { clearInterval(countdownTimer);  countdownTimer  = null }
  autoRefresh.value = false
  countdown.value   = 0
}

function setPageSize(n) {
  pageSize.value = n
  // Reset to page 1 when size changes to avoid being past last page
  page.value = 1
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) startAutoRefresh(autoRefreshInterval.value)
  else stopAutoRefresh()
}

// Export helpers
function exportCSV(rows) {
  if (!rows.length) return
  const cols = Object.keys(rows[0])
  const header = cols.join(',')
  const body = rows.map(r =>
    cols.map(c => JSON.stringify(r[c] ?? '')).join(',')
  ).join('\n')
  download(header + '\n' + body, 'rsyslox-logs.csv', 'text/csv')
}

function exportJSON(rows) {
  download(JSON.stringify(rows, null, 2), 'rsyslox-logs.json', 'application/json')
}

function download(content, filename, mime) {
  const blob = new Blob([content], { type: mime })
  const url  = URL.createObjectURL(blob)
  const a    = Object.assign(document.createElement('a'), { href: url, download: filename })
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// Reset filters and re-fetch
function resetFilters() {
  severities.value    = []
  facilities.value    = []
  hosts.value         = []
  tags.value          = []
  messageSearch.value = ''
  relativeDur.value   = '1h'
  timeMode.value      = 'relative'
  startDate.value     = ''
  endDate.value       = ''
  resetPage()
}

// Watch filter changes → reset page and re-fetch
watch([severities, facilities, hosts, tags, messageSearch, relativeDur, startDate, endDate], () => {
  resetPage()
  fetchLogs()
}, { deep: true })

watch(page, () => fetchLogs())  // arrow wrapper: prevents page value being passed as fromRefresh arg

export function useLogsStore() {
  return {
    // state
    logs, total, loading, error,
    page, pageSize, offset, totalPages, showAll,
    timeMode, relativeDur, startDate, endDate,
    severities, facilities, hosts, tags, messageSearch,
    selectedIds, selectedCount, selectedLogs,
    detailEntry,
    availableHosts, availableTags, availableSeverities, availableFacilities,
    autoRefresh, autoRefreshInterval, newIds, countdown, firstLoad,
    // actions
    fetchLogs, fetchFilterOptions, setPageSize,
    setPage, resetPage, resetFilters,
    toggleSelection, toggleSelectAll, clearSelection,
    openDetail, closeDetail,
    toggleAutoRefresh, startAutoRefresh, stopAutoRefresh,
    exportCSV, exportJSON
  }
}
