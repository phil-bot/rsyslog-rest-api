<template>
  <aside class="filter-panel" :class="{ open: modelValue }">
    <div class="panel-header">
      <span class="panel-title">{{ t('filter.title') }}</span>
      <button class="reset-btn" @click="$emit('reset')">{{ t('filter.reset') }}</button>
    </div>

    <!-- ── Time Range ─────────────────────────── -->
    <section class="filter-section">
      <div class="section-label">{{ t('filter.time_range') }}</div>

      <!-- Duration segment control — single wide bar, no borders on items -->
      <div class="dur-seg">
        <button
          v-for="d in durations" :key="d.value"
          class="dur-btn"
          :class="{ active: activeDur === d.value }"
          @click="selectDuration(d.value)"
        >{{ d.label }}</button>
      </div>

      <!-- Date fields — always visible, filled by pills or typed directly -->
      <div class="date-row">
        <span class="date-lbl">{{ t('filter.from') }}</span>
        <input type="datetime-local" class="date-input" :value="startDate"
          @input="onDateInput('start', $event.target.value)" />
      </div>
      <div class="date-row">
        <span class="date-lbl">{{ t('filter.to') }}</span>
        <input type="datetime-local" class="date-input" :value="endDate"
          @input="onDateInput('end', $event.target.value)" />
      </div>

      <!-- Shift buttons to navigate the current window earlier/later -->
      <div class="shift-row">
        <button class="shift-btn" @click="$emit('shift', -1)">{{ t('filter.earlier') }}</button>
        <button class="shift-btn" @click="$emit('shift', 1)">{{ t('filter.later') }}</button>
      </div>
    </section>

    <!-- ── Severity ───────────────────────────── -->
    <section class="filter-section">
      <div class="section-label">{{ t('filter.severity') }}</div>
      <div v-if="availableSeverities.length" class="pill-row wrap">
        <button
          v-for="item in availableSeverities" :key="item.val"
          class="pill sev-pill"
          :class="['sev-' + item.val, { active: severities.includes(item.val) }]"
          @click="toggle('severities', item.val)"
        >
          <span class="sev-dot" :style="{ background: sevColor(item.val) }"></span>
          {{ item.label }}
        </button>
      </div>
      <p v-else class="empty-hint">{{ t('filter.loading') }}</p>
    </section>

    <!-- ── Facility ───────────────────────────── -->
    <section class="filter-section">
      <div class="section-label">{{ t('filter.facility') }}</div>
      <div v-if="availableFacilities.length" class="pill-row wrap">
        <button
          v-for="item in availableFacilities" :key="item.val"
          class="pill"
          :class="{ active: facilities.includes(item.val) }"
          @click="toggle('facilities', item.val)"
        >{{ item.label ?? item.val }}</button>
      </div>
      <p v-else class="empty-hint">{{ t('filter.loading') }}</p>
    </section>

    <!-- ── Host ───────────────────────────────── -->
    <section class="filter-section">
      <div class="section-label">{{ t('filter.host') }}</div>
      <div v-if="availableHosts.length" class="pill-row wrap scrollable">
        <button
          v-for="host in availableHosts" :key="host"
          class="pill mono-pill"
          :class="{ active: hosts.includes(host) }"
          @click="toggle('hosts', host)"
        >{{ host }}</button>
      </div>
      <p v-else class="empty-hint">{{ t('filter.no_hosts') }}</p>
    </section>

    <!-- ── Tag ────────────────────────────────── -->
    <section class="filter-section">
      <div class="section-label">{{ t('filter.tag') }}</div>
      <div v-if="availableTags.length" class="pill-row wrap scrollable">
        <button
          v-for="tag in availableTags" :key="tag"
          class="pill mono-pill"
          :class="{ active: tags.includes(tag) }"
          @click="toggle('tags', tag)"
        >{{ tag }}</button>
      </div>
      <p v-else class="empty-hint">{{ t('filter.no_tags') }}</p>
    </section>

    <!-- ── Message Search ─────────────────────── -->
    <section class="filter-section">
      <div class="section-label">{{ t('filter.message_search') }}</div>
      <div class="search-wrap">
        <input
          class="search-input"
          type="text"
          :placeholder="t('filter.search_placeholder')"
          :value="messageSearch"
          @input="$emit('update:messageSearch', $event.target.value)"
        />
        <button
          v-if="messageSearch"
          class="search-clear"
          @click="$emit('update:messageSearch', '')"
          title="Clear search"
          type="button"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    </section>
  </aside>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useLocale } from '@/composables/useLocale'

const SEV_COLORS = ['#7c3aed','#9333ea','#dc2626','#ea580c','#d97706','#2563eb','#475569','#94a3b8']

const props = defineProps({
  modelValue:          { type: Boolean, default: true },
  timeMode:            { type: String,  default: 'relative' },
  relativeDur:         { type: String,  default: '1h' },
  startDate:           { type: String,  default: '' },
  endDate:             { type: String,  default: '' },
  severities:          { type: Array,   default: () => [] },
  facilities:          { type: Array,   default: () => [] },
  hosts:               { type: Array,   default: () => [] },
  tags:                { type: Array,   default: () => [] },
  messageSearch:       { type: String,  default: '' },
  availableHosts:      { type: Array,   default: () => [] },
  availableTags:       { type: Array,   default: () => [] },
  availableSeverities: { type: Array,   default: () => [] },
  availableFacilities: { type: Array,   default: () => [] },
})

const { t } = useLocale()
const emit = defineEmits([
  'update:timeMode','update:relativeDur','update:startDate','update:endDate',
  'update:severities','update:facilities','update:hosts','update:tags',
  'update:messageSearch','shift','reset',
])

const durations = [
  { value: '15m', label: '15m' }, { value: '1h',  label: '1h'  },
  { value: '6h',  label: '6h'  }, { value: '24h', label: '24h' },
  { value: '7d',  label: '7d'  }, { value: '30d', label: '30d' },
]

// Duration offsets in ms
const DURATION_MS = {
  '15m': 15*60*1000, '1h': 60*60*1000, '6h': 6*60*60*1000,
  '24h': 24*60*60*1000, '7d': 7*24*60*60*1000, '30d': 30*24*60*60*1000,
}

// Which pill is currently active (matches current date range, or null if custom)
const activeDur = computed(() => {
  // Check if current date range matches a known duration
  if (!props.startDate || !props.endDate) return props.relativeDur
  const start = new Date(props.startDate)
  const end   = new Date(props.endDate)
  const diff  = end - start
  const match = Object.entries(DURATION_MS).find(([, ms]) => Math.abs(diff - ms) < 60000)
  return match ? match[0] : null
})

// Clicking a pill fills both date fields with the corresponding range ending now
// Fill date fields immediately on mount using the default relativeDur
onMounted(() => selectDuration(props.relativeDur))

function selectDuration(val) {
  const now   = new Date()
  const from  = new Date(now - DURATION_MS[val])
  const fmt   = d => d.toISOString().slice(0, 16)  // "YYYY-MM-DDTHH:MM"
  emit('update:startDate', fmt(from))
  emit('update:endDate',   fmt(now))
  emit('update:timeMode',  'absolute')
  emit('update:relativeDur', val)
}

// Typing in a date field deactivates the pill highlight
function onDateInput(which, value) {
  if (which === 'start') emit('update:startDate', value)
  else                   emit('update:endDate',   value)
  emit('update:timeMode', 'absolute')
}

function sevColor(val) { return SEV_COLORS[val] ?? '#94a3b8' }

function toggle(field, val) {
  const copy = [...props[field]]
  const idx  = copy.indexOf(val)
  idx === -1 ? copy.push(val) : copy.splice(idx, 1)
  emit('update:' + field, copy)
}
</script>

<style scoped>
.filter-panel {
  width: var(--sidebar-width); flex-shrink: 0;
  background: var(--bg-surface); border-right: 1px solid var(--border);
  overflow-y: auto; display: flex; flex-direction: column; transition: width .2s;
}
@media (max-width: 768px) {
  .filter-panel {
    position: fixed; top: var(--header-height); left: 0; bottom: 0;
    z-index: 50; transform: translateX(-100%);
    transition: transform .25s; box-shadow: 4px 0 16px rgba(0,0,0,.15);
  }
  .filter-panel.open { transform: translateX(0); }
}
@media (min-width: 769px) {
  .filter-panel:not(.open) { width: 0; overflow: hidden; border: none; }
}

.panel-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: .75rem 1rem; border-bottom: 1px solid var(--border);
  position: sticky; top: 0; background: var(--bg-surface); z-index: 1;
}
.panel-title { font-weight: 600; font-size: .875rem; color: var(--text); }
.reset-btn {
  background: none; border: none; cursor: pointer;
  color: var(--color-primary); font-size: .78rem;
  padding: .2rem .45rem; border-radius: var(--radius);
}
.reset-btn:hover { background: var(--bg-hover); }

.filter-section {
  padding: .8rem 1rem; border-bottom: 1px solid var(--border);
  display: flex; flex-direction: column; gap: .5rem;
}
.section-label {
  font-size: .68rem; font-weight: 700;
  text-transform: uppercase; letter-spacing: .07em; color: var(--text-muted);
}


/* ── Duration segment control ───────────────── */
.dur-seg {
  display: flex;
  width: 100%;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}
.dur-btn {
  flex: 1;
  padding: .3rem .25rem;
  background: transparent;
  border: none;
  border-right: 1px solid var(--border);
  cursor: pointer;
  font-size: .78rem;
  color: var(--text-muted);
  transition: background .12s, color .12s;
  white-space: nowrap;
}
.dur-btn:last-child { border-right: none; }
.dur-btn:hover { background: var(--bg-hover); color: var(--text); }
.dur-btn.active { background: var(--color-primary); color: #fff; }

/* ── Pills ──────────────────────────────────── */
.pill-row { display: flex; gap: .3rem; justify-content: center; flex-wrap: wrap; }
.pill-row.scrollable { max-height: 130px; overflow-y: auto; align-content: flex-start; justify-content: flex-start; }

.pill {
  display: inline-flex; align-items: center; gap: .3rem;
  padding: .22rem .6rem;
  border: 1.5px solid var(--border);
  border-radius: 20px;
  background: transparent;
  color: var(--text-muted);
  font-size: .78rem;
  cursor: pointer;
  transition: border-color .15s, color .15s, background .15s;
  white-space: nowrap; line-height: 1.4;
}
.pill:hover { border-color: var(--color-primary); color: var(--text); }

/* Generic active: filled */
.pill.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
}

/* ── Severity pills: active = colored border only ── */
.sev-pill { color: var(--text-muted); }
.sev-pill:hover { color: var(--text); }

/* Each severity: its own border color when active */
.sev-pill.sev-0.active { border-color: #7c3aed; color: #7c3aed; background: transparent; }
.sev-pill.sev-1.active { border-color: #9333ea; color: #9333ea; background: transparent; }
.sev-pill.sev-2.active { border-color: #dc2626; color: #dc2626; background: transparent; }
.sev-pill.sev-3.active { border-color: #ea580c; color: #ea580c; background: transparent; }
.sev-pill.sev-4.active { border-color: #d97706; color: #d97706; background: transparent; }
.sev-pill.sev-5.active { border-color: #2563eb; color: #2563eb; background: transparent; }
.sev-pill.sev-6.active { border-color: #475569; color: #475569; background: transparent; }
.sev-pill.sev-7.active { border-color: #94a3b8; color: #94a3b8; background: transparent; }

.sev-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }

.mono-pill { font-family: ui-monospace,'SF Mono',Menlo,monospace; font-size: .74rem; }

/* Shift row (absolute mode only) */
.shift-row { display: flex; gap: .375rem; }
.shift-btn {
  flex: 1; background: var(--bg); border: 1px solid var(--border);
  border-radius: var(--radius); cursor: pointer;
  padding: .28rem .5rem; font-size: .78rem; color: var(--text-muted);
}
.shift-btn:hover { background: var(--bg-hover); color: var(--text); }

.date-row {
  display: flex; align-items: center; gap: .5rem;
}
.date-lbl {
  font-size: .8rem; color: var(--text-muted); flex-shrink: 0; width: 2rem; text-align: right;
}
.date-input {
  flex: 1; padding: .32rem .45rem; border: 1px solid var(--border);
  border-radius: var(--radius); background: var(--bg); color: var(--text); font-size: .78rem;
  min-width: 0;
}

.search-wrap {
  position: relative;
}
.search-input {
  width: 100%; padding: .38rem 1.75rem .38rem .6rem;
  border: 1px solid var(--border); border-radius: var(--radius);
  background: var(--bg); color: var(--text); font-size: .875rem;
}
.search-input:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.search-clear {
  position: absolute; right: .35rem; top: 50%;
  transform: translateY(-50%);
  background: none; border: none; cursor: pointer;
  color: var(--text-muted); padding: .2rem; border-radius: 3px;
  display: flex; align-items: center;
  transition: color .15s, background .15s;
}
.search-clear:hover { color: var(--text); background: var(--bg-hover); }

.empty-hint { font-size: .78rem; color: var(--text-muted); }
</style>
