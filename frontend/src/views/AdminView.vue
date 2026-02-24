<template>
  <div class="admin-layout">
    <AppHeader @toggle-sidebar="() => {}" />

    <div class="admin-body">
      <nav class="admin-nav">
        <button v-for="tab in tabs" :key="tab.id" class="nav-tab"
          :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id">
          <span class="tab-icon" v-html="tab.svg"></span>
          {{ tab.label }}
        </button>
      </nav>

      <main class="admin-main">
        <!-- Server -->
        <section v-if="activeTab === 'server'" class="admin-section">
          <div class="section-header">
            <h2>{{ t('admin.server_title') }}</h2>
            <p class="section-desc">{{ t('admin.server_desc') }}</p>
          </div>
          <div v-if="configLoading" class="loading-msg">{{ t('admin.loading') }}</div>
          <form v-else class="config-form" @submit.prevent="saveServer">
            <div class="field-row">
              <label class="field-label">{{ t('admin.field_host') }}
                <input v-model="serverForm.host" class="field-input" disabled />
                <span class="field-hint">{{ t('admin.hint_restart') }}</span>
              </label>
              <label class="field-label">{{ t('admin.field_port') }}
                <input v-model.number="serverForm.port" class="field-input" disabled />
              </label>
            </div>
            <label class="field-label">{{ t('admin.field_origins') }}
              <input v-model="serverForm.originsStr" class="field-input" placeholder="*" />
              <span class="field-hint">{{ t('admin.field_origins_hint') }}</span>
            </label>
            <label class="toggle-label big">
              <input type="checkbox" v-model="serverForm.useSSL" />
              {{ t('admin.field_ssl') }}
            </label>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="saving">
                {{ saving ? t('admin.saving') : t('admin.save') }}
              </button>
              <span v-if="saveMsg.server" class="save-msg" :class="saveMsg.serverOk ? 'ok' : 'err'">
                {{ saveMsg.server }}
              </span>
            </div>
          </form>
        </section>

        <!-- Database -->
        <section v-if="activeTab === 'database'" class="admin-section">
          <div class="section-header">
            <h2>{{ t('admin.db_title') }}</h2>
            <p class="section-desc">{{ t('admin.db_desc') }}</p>
          </div>
          <div v-if="configLoading" class="loading-msg">{{ t('admin.loading') }}</div>
          <div v-else class="info-grid">
            <div class="info-row"><span class="info-key">Host</span><span class="info-val mono">{{ cfg.database?.host }}:{{ cfg.database?.port }}</span></div>
            <div class="info-row"><span class="info-key">Database</span><span class="info-val mono">{{ cfg.database?.name }}</span></div>
            <div class="info-row"><span class="info-key">{{ t('admin.db_user') }}</span><span class="info-val mono">{{ cfg.database?.user }}</span></div>
            <div class="info-row"><span class="info-key">{{ t('admin.db_password') }}</span><span class="info-val mono text-muted">{{ t('admin.db_password_val') }}</span></div>
          </div>
        </section>

        <!-- Cleanup -->
        <section v-if="activeTab === 'cleanup'" class="admin-section">
          <div class="section-header">
            <h2>{{ t('admin.cleanup_title') }}</h2>
            <p class="section-desc">{{ t('admin.cleanup_desc') }}</p>
          </div>
          <div v-if="configLoading" class="loading-msg">{{ t('admin.loading') }}</div>
          <form v-else class="config-form" @submit.prevent="saveCleanup">
            <label class="toggle-label big">
              <input type="checkbox" v-model="cleanupForm.enabled" />
              {{ t('admin.cleanup_enable') }}
            </label>
            <div class="field-row" :class="{ disabled: !cleanupForm.enabled }">
              <label class="field-label">{{ t('admin.cleanup_disk_path') }}
                <input v-model="cleanupForm.diskPath" class="field-input"
                  :disabled="!cleanupForm.enabled" placeholder="/var/lib/mysql" />
              </label>
              <label class="field-label">{{ t('admin.cleanup_threshold') }}
                <input v-model.number="cleanupForm.thresholdPercent" type="number" min="1" max="100"
                  class="field-input" style="max-width:100px" :disabled="!cleanupForm.enabled" />
              </label>
            </div>
            <div class="field-row" :class="{ disabled: !cleanupForm.enabled }">
              <label class="field-label">{{ t('admin.cleanup_batch') }}
                <input v-model.number="cleanupForm.batchSize" type="number" min="1"
                  class="field-input" style="max-width:120px" :disabled="!cleanupForm.enabled" />
              </label>
              <label class="field-label">{{ t('admin.cleanup_interval') }}
                <input v-model.number="cleanupForm.intervalSeconds" type="number" min="60"
                  class="field-input" style="max-width:120px" :disabled="!cleanupForm.enabled" />
              </label>
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="saving">
                {{ saving ? t('admin.saving') : t('admin.save') }}
              </button>
              <span v-if="saveMsg.cleanup" class="save-msg" :class="saveMsg.cleanupOk ? 'ok' : 'err'">
                {{ saveMsg.cleanup }}
              </span>
            </div>
          </form>
        </section>

        <!-- API Keys -->
        <section v-if="activeTab === 'keys'" class="admin-section">
          <div class="section-header">
            <h2>{{ t('admin.keys_title') }}</h2>
            <p class="section-desc">{{ t('admin.keys_desc') }}</p>
          </div>
          <div class="new-key-form">
            <input v-model="newKeyName" class="field-input"
              :placeholder="t('admin.keys_placeholder')" @keydown.enter.prevent="createKey"
              style="max-width:280px" />
            <button class="btn btn-primary" :disabled="!newKeyName.trim() || keyCreating" @click="createKey">
              {{ keyCreating ? t('admin.keys_creating') : t('admin.keys_create') }}
            </button>
          </div>
          <div v-if="newKeyPlaintext" class="key-reveal">
            <div class="key-reveal-header">
              <span>🔑</span>
              <strong>{{ t('admin.keys_created_for') }} "{{ newKeyRevealName }}"</strong>
              <span class="key-reveal-warn">{{ t('admin.keys_copy_note') }}</span>
            </div>
            <div class="key-reveal-value">
              <code class="mono">{{ newKeyPlaintext }}</code>
              <button class="copy-btn" @click="copyKey">{{ keyCopied ? t('admin.keys_copied') : t('admin.keys_copy') }}</button>
            </div>
            <button class="btn btn-ghost btn-sm" @click="newKeyPlaintext = null">{{ t('admin.keys_dismiss') }}</button>
          </div>
          <div v-if="keysLoading" class="loading-msg">{{ t('admin.loading') }}</div>
          <div v-else-if="!keys.length" class="empty-keys">{{ t('admin.keys_none') }}</div>
          <ul v-else class="keys-list">
            <li v-for="key in keys" :key="key.name" class="key-item">
              <div class="key-info">
                <span class="key-name mono">{{ key.name }}</span>
                <span class="key-badge">{{ t('admin.keys_readonly') }}</span>
              </div>
              <button class="btn btn-danger btn-sm" @click="confirmDelete(key.name)">{{ t('admin.keys_revoke') }}</button>
            </li>
          </ul>
        </section>

        <!-- ── Preferences ──────────────────────────── -->
        <section v-if="activeTab === 'prefs'" class="admin-section">
          <div class="section-header">
            <h2>{{ t('prefs.title') }}</h2>
            <p class="section-desc">{{ t('prefs.desc') }}</p>
          </div>
          <div class="config-form">
            <!-- Language -->
            <label class="field-label">
              {{ t('prefs.language') }}
              <select v-model="language" class="field-input">
                <option value="en">English</option>
                <option value="de">Deutsch</option>
              </select>
            </label>
            <!-- Time format -->
            <label class="field-label">
              {{ t('prefs.time_format') }}
              <select v-model="timeFormat" class="field-input">
                <option value="24h">{{ t('prefs.time_format_24h') }}</option>
                <option value="12h">{{ t('prefs.time_format_12h') }}</option>
              </select>
            </label>
            <!-- Font size -->
            <label class="field-label">
              {{ t('prefs.font_size') }}
              <div class="radio-group">
                <label class="radio-opt"><input type="radio" v-model="fontSize" value="small" /> {{ t('prefs.font_small') }}</label>
                <label class="radio-opt"><input type="radio" v-model="fontSize" value="medium" /> {{ t('prefs.font_medium') }}</label>
                <label class="radio-opt"><input type="radio" v-model="fontSize" value="large" /> {{ t('prefs.font_large') }}</label>
              </div>
            </label>
            <!-- Auto-refresh interval -->
            <label class="field-label">
              {{ t('prefs.auto_refresh') }}
              <div class="inline-field">
                <input v-model.number="autoRefreshIntervalPref" type="number" min="5" max="300"
                  class="field-input" style="max-width:100px" />
                <span class="field-hint">{{ t('prefs.auto_refresh_unit') }}</span>
              </div>
            </label>
          </div>
        </section>
      </main>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="deleteTarget" class="modal-backdrop" @click.self="deleteTarget = null">
          <div class="confirm-dialog">
            <h3>{{ t('admin.keys_revoke_title', { name: deleteTarget }) }}</h3>
            <p>{{ t('admin.keys_revoke_desc') }}</p>
            <div class="confirm-actions">
              <button class="btn btn-ghost" @click="deleteTarget = null">{{ t('admin.cancel') }}</button>
              <button class="btn btn-danger" :disabled="deleting" @click="deleteKey">
                {{ deleting ? t('admin.keys_revoking') : t('admin.keys_revoke') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { language, timeFormat, fontSize, autoRefreshInterval as autoRefreshIntervalPref } from '@/stores/preferences'
import { useLocale } from '@/composables/useLocale'
import AppHeader from '@/components/AppHeader.vue'
import { api } from '@/api/client'

const { t } = useLocale()

const PREFS_SVG  = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>'

const tabs = computed(() => [
  { id: 'prefs',    label: t('admin.tab_preferences'), svg: PREFS_SVG },
  { id: 'server',   label: t('admin.tab_server'),   svg: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="5" rx="1"/><rect x="2" y="11" width="20" height="5" rx="1"/><rect x="2" y="19" width="20" height="5" rx="1"/></svg>' },
  { id: 'database', label: t('admin.tab_database'), svg: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v4c0 1.66 4.03 3 9 3s9-1.34 9-3V5"/><path d="M3 9v4c0 1.66 4.03 3 9 3s9-1.34 9-3V9"/><path d="M3 13v4c0 1.66 4.03 3 9 3s9-1.34 9-3v-4"/></svg>' },
  { id: 'cleanup',  label: t('admin.tab_cleanup'),  svg: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M9 6V4h6v2"/></svg>' },
  { id: 'keys',     label: t('admin.tab_keys'),     svg: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="7.5" cy="15.5" r="5.5"/><path d="M21 2l-9.6 9.6"/><path d="M15.5 7.5l3 3L22 7l-3-3"/></svg>' },
])
const activeTab = ref('prefs')

const cfg           = ref({})
const configLoading = ref(false)
const saving        = ref(false)
const saveMsg       = reactive({ server:'', serverOk:true, cleanup:'', cleanupOk:true })

const serverForm  = reactive({ host:'', port:8000, originsStr:'*', autoRefreshInterval:30, useSSL:false })
const cleanupForm = reactive({ enabled:false, diskPath:'/var/lib/mysql', thresholdPercent:85, batchSize:1000, intervalSeconds:900 })

async function loadConfig() {
  configLoading.value = true
  try {
    cfg.value = await api.getConfig()
    const s = cfg.value.server ?? {}
    serverForm.host                = s.host ?? ''
    serverForm.port                = s.port ?? 8000
    serverForm.originsStr          = (s.allowed_origins ?? ['*']).join(', ')
    serverForm.autoRefreshInterval = s.auto_refresh_interval ?? 30
    serverForm.useSSL              = s.use_ssl ?? false
    const c = cfg.value.cleanup ?? {}
    cleanupForm.enabled          = c.enabled ?? false
    cleanupForm.diskPath         = c.disk_path ?? '/var/lib/mysql'
    cleanupForm.thresholdPercent = c.threshold_percent ?? 85
    cleanupForm.batchSize        = c.batch_size ?? 1000
    cleanupForm.intervalSeconds  = c.interval_seconds ?? 900
  } catch (e) {
    saveMsg.server = 'Failed to load: ' + (e.message || e); saveMsg.serverOk = false
  } finally { configLoading.value = false }
}

async function saveServer() {
  saving.value = true; saveMsg.server = ''
  try {
    await api.updateConfig({ server: {
      allowed_origins: serverForm.originsStr.split(',').map(s=>s.trim()).filter(Boolean),
      auto_refresh_interval: serverForm.autoRefreshInterval,
      use_ssl: serverForm.useSSL,
    }})
    saveMsg.server = '✓ Saved'; saveMsg.serverOk = true
    setTimeout(() => { saveMsg.server = '' }, 3000)
  } catch (e) { saveMsg.server = e.body?.message || 'Save failed'; saveMsg.serverOk = false }
  finally { saving.value = false }
}

async function saveCleanup() {
  saving.value = true; saveMsg.cleanup = ''
  try {
    await api.updateConfig({ cleanup: {
      enabled: cleanupForm.enabled,
      disk_path: cleanupForm.diskPath,
      threshold_percent: cleanupForm.thresholdPercent,
      batch_size: cleanupForm.batchSize,
      interval_seconds: cleanupForm.intervalSeconds,
    }})
    saveMsg.cleanup = '✓ Saved'; saveMsg.cleanupOk = true
    setTimeout(() => { saveMsg.cleanup = '' }, 3000)
  } catch (e) { saveMsg.cleanup = e.body?.message || 'Save failed'; saveMsg.cleanupOk = false }
  finally { saving.value = false }
}

const keys = ref([]); const keysLoading = ref(false)
const newKeyName = ref(''); const keyCreating = ref(false)
const newKeyPlaintext = ref(null); const newKeyRevealName = ref('')
const keyCopied = ref(false); const deleteTarget = ref(null); const deleting = ref(false)

async function loadKeys() {
  keysLoading.value = true
  try { keys.value = await api.listKeys() } catch {}
  finally { keysLoading.value = false }
}

async function createKey() {
  if (!newKeyName.value.trim()) return
  keyCreating.value = true
  try {
    const res = await api.createKey(newKeyName.value.trim())
    newKeyPlaintext.value = res.key; newKeyRevealName.value = res.name; newKeyName.value = ''
    await loadKeys()
  } catch (e) { alert(e.body?.message || 'Failed') }
  finally { keyCreating.value = false }
}

async function copyKey() {
  try { await navigator.clipboard.writeText(newKeyPlaintext.value); keyCopied.value = true; setTimeout(()=>{keyCopied.value=false},2000) } catch {}
}

function confirmDelete(name) { deleteTarget.value = name }
async function deleteKey() {
  deleting.value = true
  try { await api.deleteKey(deleteTarget.value); deleteTarget.value = null; await loadKeys() }
  catch (e) { alert(e.body?.message || 'Failed') }
  finally { deleting.value = false }
}

onMounted(() => { loadConfig(); loadKeys() })
</script>

<style scoped>
.admin-layout { display:flex; flex-direction:column; height:100%; overflow:hidden; }
.admin-body { display:flex; flex:1; overflow:hidden; }
.admin-nav { width:180px; flex-shrink:0; background:var(--bg-surface); border-right:1px solid var(--border); padding:.75rem .5rem; display:flex; flex-direction:column; gap:.25rem; overflow-y:auto; }
.nav-tab { display:flex; align-items:center; gap:.5rem; padding:.5rem .75rem; border:none; border-radius:var(--radius); background:none; cursor:pointer; font-size:.875rem; color:var(--text-muted); text-align:left; transition:background .15s,color .15s; width:100%; }
.nav-tab:hover { background:var(--bg-hover); color:var(--text); }
.nav-tab.active { background:var(--bg-selected); color:var(--color-primary); font-weight:600; }
.tab-icon { display:flex; flex-shrink:0; }
.admin-main { flex:1; overflow-y:auto; padding:1.5rem; background:var(--bg); color:var(--text); }
.admin-section { max-width:640px; display:flex; flex-direction:column; gap:1.25rem; }
.section-header { display:flex; flex-direction:column; gap:.25rem; }
.section-header h2 { font-size:1.125rem; font-weight:700; }
.section-desc { font-size:.875rem; color:var(--text-muted); }
.section-desc code { font-family:ui-monospace,monospace; background:var(--bg-hover); padding:.1rem .3rem; border-radius:3px; font-size:.85em; }
.loading-msg { color:var(--text-muted); font-size:.875rem; }
.config-form { display:flex; flex-direction:column; gap:1rem; background:var(--bg-surface); border:1px solid var(--border); border-radius:var(--radius); padding:1.25rem; }
.field-row { display:grid; grid-template-columns:1fr 1fr; gap:.75rem; }
.field-label { display:flex; flex-direction:column; gap:.375rem; font-size:.875rem; font-weight:500; }
.field-input { padding:.4rem .625rem; border:1px solid var(--border); border-radius:var(--radius); background:var(--bg); color:var(--text); font-size:.875rem; width:100%; }
.field-input:focus { outline:2px solid var(--color-primary); outline-offset:-1px; }
.field-input:disabled { opacity:.5; cursor:not-allowed; }
.field-hint { font-size:.75rem; color:var(--text-muted); font-weight:400; }
.toggle-label { display:flex; align-items:center; gap:.5rem; font-size:.875rem; cursor:pointer; }
.toggle-label.big { font-weight:500; }
.toggle-label input { accent-color:var(--color-primary); width:16px; height:16px; cursor:pointer; }
.disabled { opacity:.5; pointer-events:none; }
.form-actions { display:flex; align-items:center; gap:.75rem; padding-top:.25rem; }
.save-msg { font-size:.875rem; }
.save-msg.ok { color:#16a34a; }
.save-msg.err { color:#dc2626; }
.info-grid { background:var(--bg-surface); border:1px solid var(--border); border-radius:var(--radius); overflow:hidden; }
.info-row { display:flex; gap:1rem; padding:.625rem 1rem; border-bottom:1px solid var(--border); }
.info-row:last-child { border-bottom:none; }
.info-key { width:100px; flex-shrink:0; font-size:.8rem; font-weight:600; color:var(--text-muted); text-transform:uppercase; letter-spacing:.04em; padding-top:.1rem; }
.info-val { font-size:.875rem; }
.text-muted { color:var(--text-muted); }
.new-key-form { display:flex; gap:.625rem; align-items:center; flex-wrap:wrap; }
.key-reveal { background:#f0fdf4; border:1px solid #bbf7d0; border-radius:var(--radius); padding:1rem; display:flex; flex-direction:column; gap:.75rem; }
[data-theme="dark"] .key-reveal { background:#052e16; border-color:#166534; }
.key-reveal-header { display:flex; align-items:center; gap:.5rem; flex-wrap:wrap; font-size:.875rem; }
.key-reveal-warn { font-size:.8rem; color:#16a34a; margin-left:auto; }
[data-theme="dark"] .key-reveal-warn { color:#4ade80; }
.key-reveal-value { display:flex; align-items:center; gap:.625rem; background:var(--bg); border:1px solid var(--border); border-radius:var(--radius); padding:.5rem .75rem; overflow:hidden; }
.key-reveal-value code { flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-size:.8rem; }
.copy-btn { background:var(--color-primary); color:#fff; border:none; border-radius:var(--radius); padding:.25rem .625rem; font-size:.8rem; cursor:pointer; flex-shrink:0; }
.btn-sm { font-size:.8rem; padding:.3rem .625rem; }
.empty-keys { font-size:.875rem; color:var(--text-muted); }
.keys-list { list-style:none; background:var(--bg-surface); border:1px solid var(--border); border-radius:var(--radius); overflow:hidden; }
.key-item { display:flex; align-items:center; justify-content:space-between; padding:.75rem 1rem; border-bottom:1px solid var(--border); gap:.75rem; }
.key-item:last-child { border-bottom:none; }
.key-info { display:flex; align-items:center; gap:.75rem; }
.key-name { font-size:.875rem; font-family:ui-monospace,monospace; }
.key-badge { font-size:.7rem; padding:.1rem .375rem; background:var(--bg-hover); border:1px solid var(--border); border-radius:999px; color:var(--text-muted); }
.modal-backdrop { position:fixed; inset:0; background:rgba(0,0,0,.4); z-index:300; display:flex; align-items:center; justify-content:center; padding:1rem; }
.confirm-dialog { background:var(--bg-surface); border:1px solid var(--border); border-radius:var(--radius); padding:1.5rem; max-width:360px; width:100%; display:flex; flex-direction:column; gap:.75rem; box-shadow:0 20px 40px rgba(0,0,0,.2); }
.confirm-dialog h3 { font-size:1rem; font-weight:700; }
.confirm-dialog p { font-size:.875rem; color:var(--text-muted); }
.confirm-actions { display:flex; gap:.625rem; justify-content:flex-end; }
.modal-enter-active,.modal-leave-active { transition:opacity .2s; }
.modal-enter-from,.modal-leave-to { opacity:0; }
@media (max-width:600px) {
  .admin-nav { width:100%; flex-direction:row; border-right:none; border-bottom:1px solid var(--border); overflow-x:auto; padding:.5rem; }
  .nav-tab { flex-shrink:0; }
  .admin-body { flex-direction:column; }
}

.radio-group { display: flex; gap: 1rem; margin-top: .25rem; }
.radio-opt { display: flex; align-items: center; gap: .35rem; font-size: .875rem; cursor: pointer; }
.inline-field { display: flex; align-items: center; gap: .5rem; margin-top: .25rem; }
</style>
