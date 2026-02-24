import { ref, watch } from 'vue'

const STORAGE_KEY = 'rsyslox_prefs'

function load() {
  try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}') } catch { return {} }
}
function save(prefs) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs))
}

const stored = load()

export const language            = ref(stored.language            ?? 'en')
export const timeFormat          = ref(stored.timeFormat          ?? '24h')
export const fontSize            = ref(stored.fontSize            ?? 'medium')  // 'small'|'medium'|'large'
export const autoRefreshInterval = ref(stored.autoRefreshInterval ?? 30)        // seconds

// Apply font-size immediately on load
applyFontSize(fontSize.value)

export function applyFontSize(size) {
  const map = { small: '13px', medium: '14px', large: '15px' }
  document.documentElement.style.setProperty('font-size', map[size] ?? '14px')
}

// Persist + apply on every change
watch([language, timeFormat, fontSize, autoRefreshInterval], () => {
  save({
    language:            language.value,
    timeFormat:          timeFormat.value,
    fontSize:            fontSize.value,
    autoRefreshInterval: autoRefreshInterval.value,
  })
  applyFontSize(fontSize.value)
})

export function usePreferences() {
  return { language, timeFormat, fontSize, autoRefreshInterval }
}
