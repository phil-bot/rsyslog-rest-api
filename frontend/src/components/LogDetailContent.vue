<template>
  <div class="detail-content">
    <!-- Message (most important - full text) -->
    <section class="detail-section">
      <div class="detail-label">Message</div>
      <pre class="detail-message">{{ entry.Message }}</pre>
    </section>

    <!-- Key fields -->
    <section class="detail-section">
      <div class="detail-label">Details</div>
      <table class="detail-table">
        <tbody>
          <tr v-for="field in keyFields" :key="field.label">
            <td class="field-label">{{ field.label }}</td>
            <td class="field-value mono">{{ field.value }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <!-- All fields (collapsible) -->
    <section class="detail-section">
      <button class="expand-btn" @click="expanded = !expanded">
        {{ expanded ? '▾' : '▸' }} All fields ({{ allFields.length }})
      </button>
      <div v-if="expanded">
        <table class="detail-table">
          <tbody>
            <tr v-for="field in allFields" :key="field.label">
              <td class="field-label">{{ field.label }}</td>
              <td class="field-value mono">{{ field.value }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Copy button -->
    <button class="copy-btn btn btn-ghost" @click="copyRaw">
      {{ copied ? '✓ Copied' : 'Copy raw JSON' }}
    </button>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { SEVERITY_LABELS, FACILITY_LABELS } from '@/stores/logs'

const props = defineProps({ entry: { type: Object, required: true } })

const expanded = ref(false)
const copied   = ref(false)

function fmt(val) {
  if (val == null) return '—'
  if (typeof val === 'object') return val.Label || (val.Val ?? JSON.stringify(val))
  return String(val)
}

function formatTs(ts) {
  if (!ts) return '—'
  const d = new Date(ts)
  return isNaN(d) ? ts : d.toLocaleString()
}

const keyFields = computed(() => {
  const e = props.entry
  const sevVal = typeof e.Severity === 'number' ? e.Severity
    : (typeof e.Priority === 'number' ? e.Priority % 8 : null)
  const sevLabel = e.Severity_Label || (sevVal != null ? (SEVERITY_LABELS[sevVal] ?? sevVal) : '—')
  const facVal   = typeof e.Facility === 'number' ? e.Facility : null
  const facLabel = e.Facility_Label || (facVal != null ? (FACILITY_LABELS[facVal] ?? facVal) : fmt(e.Facility))

  return [
    { label: 'Time (received)',  value: formatTs(e.ReceivedAt) },
    { label: 'Time (reported)',  value: formatTs(e.DeviceReportedTime) },
    { label: 'Host',             value: fmt(e.FromHost) },
    { label: 'Tag',              value: fmt(e.SysLogTag) },
    { label: 'Severity',         value: sevLabel },
    { label: 'Facility',         value: facLabel },
  ]
})

const allFields = computed(() => {
  const skip = new Set(['Message', 'ReceivedAt', 'DeviceReportedTime', 'FromHost',
    'SysLogTag', 'Severity', 'Facility', 'Priority'])
  return Object.entries(props.entry)
    .filter(([k, v]) => !skip.has(k) && v != null && v !== '')
    .map(([k, v]) => ({ label: k, value: fmt(v) }))
})

async function copyRaw() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(props.entry, null, 2))
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {}
}
</script>

<style scoped>
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.detail-section { display: flex; flex-direction: column; gap: .5rem; }

.detail-label {
  font-size: .7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .06em;
  color: var(--text-muted);
}

.detail-message {
  font-family: ui-monospace, 'Cascadia Code', 'SF Mono', Menlo, monospace;
  font-size: .78rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: .625rem;
  color: var(--text);
  max-height: 160px;
  overflow-y: auto;
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
  font-size: .8rem;
}
.detail-table tr { border-bottom: 1px solid var(--border); }
.detail-table tr:last-child { border-bottom: none; }

.field-label {
  padding: .3rem .5rem .3rem 0;
  color: var(--text-muted);
  width: 40%;
  vertical-align: top;
  font-size: .78rem;
}

.field-value {
  padding: .3rem 0;
  word-break: break-all;
  color: var(--text);
  font-size: .78rem;
}

.mono {
  font-family: ui-monospace, 'Cascadia Code', 'SF Mono', Menlo, monospace;
}

.expand-btn {
  background: none; border: none; cursor: pointer;
  color: var(--color-primary); font-size: .8rem;
  padding: .25rem 0; text-align: left;
}
.expand-btn:hover { text-decoration: underline; }

.copy-btn {
  font-size: .8rem;
  align-self: flex-start;
}
</style>
