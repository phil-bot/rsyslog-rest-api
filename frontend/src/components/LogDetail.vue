<template>
  <!-- Desktop: sidebar -->
  <aside v-if="!isMobile" class="detail-sidebar" :class="{ open: !!entry }">
    <div v-if="entry" class="detail-inner">
      <div class="detail-header">
        <span :class="'sev-badge sev-' + sevVal">{{ sevLabel }}</span>
        <button class="close-btn" @click="$emit('close')" title="Close">✕</button>
      </div>
      <DetailContent :entry="entry" />
    </div>
  </aside>

  <!-- Mobile: modal overlay -->
  <Teleport v-else to="body">
    <Transition name="modal">
      <div v-if="entry" class="modal-backdrop" @click.self="$emit('close')">
        <div class="modal-panel">
          <div class="detail-header">
            <span :class="'sev-badge sev-' + sevVal">{{ sevLabel }}</span>
            <button class="close-btn" @click="$emit('close')" title="Close">✕</button>
          </div>
          <DetailContent :entry="entry" />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { SEVERITY_LABELS } from '@/stores/logs'
import DetailContent from './LogDetailContent.vue'

const props = defineProps({
  entry: { type: Object, default: null }
})
defineEmits(['close'])

const isMobile = ref(window.innerWidth < 768)
function onResize() { isMobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const sevVal = computed(() => {
  if (!props.entry) return 6
  if (typeof props.entry.Severity === 'number') return props.entry.Severity
  if (typeof props.entry.Priority === 'number') return props.entry.Priority % 8
  return 6
})

const sevLabel = computed(() => SEVERITY_LABELS[sevVal.value] ?? 'Info')
</script>

<style scoped>
/* Desktop sidebar */
.detail-sidebar {
  width: 0;
  flex-shrink: 0;
  overflow: hidden;
  border-left: 0 solid var(--border);
  background: var(--bg-surface);
  transition: width .25s, border-width .25s;
}
.detail-sidebar.open {
  width: 380px;
  border-left-width: 1px;
  overflow-y: auto;
}

@media (max-width: 1100px) {
  .detail-sidebar.open { width: 300px; }
}

.detail-inner { padding: 1rem; min-width: 280px; }

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.close-btn {
  background: none; border: none; cursor: pointer;
  color: var(--text-muted); font-size: 1rem; padding: .25rem;
  border-radius: var(--radius);
}
.close-btn:hover { background: var(--bg-hover); color: var(--text); }

/* Mobile modal */
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,.4);
  z-index: 300;
  display: flex;
  align-items: flex-end;
  justify-content: stretch;
}
.modal-panel {
  width: 100%;
  max-height: 80vh;
  overflow-y: auto;
  background: var(--bg-surface);
  border-top: 1px solid var(--border);
  border-radius: 12px 12px 0 0;
  padding: 1rem;
}

.modal-enter-active, .modal-leave-active { transition: opacity .2s, transform .25s; }
.modal-enter-from, .modal-leave-to { opacity: 0; transform: translateY(40px); }

.sev-badge {
  display: inline-block;
  padding: .2rem .5rem;
  border-radius: 4px;
  font-size: .8rem;
  font-weight: 700;
  color: #fff;
}
.sev-0 { background: var(--sev-0); }
.sev-1 { background: var(--sev-1); }
.sev-2 { background: var(--sev-2); }
.sev-3 { background: var(--sev-3); }
.sev-4 { background: var(--sev-4); }
.sev-5 { background: var(--sev-5); }
.sev-6 { background: var(--sev-6); }
.sev-7 { background: var(--sev-7); color: var(--text); }
</style>
