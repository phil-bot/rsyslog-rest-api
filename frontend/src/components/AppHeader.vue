<template>
  <header class="app-header">
    <div class="header-left">
      <a href="/logs" class="logo-link">
        <img :src="logoSrc" alt="rsyslox" class="logo-img" />
      </a>
      <button class="icon-btn filter-toggle" @click="$emit('toggle-sidebar')" :title="t('nav.toggle_sidebar')">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/>
        </svg>
      </button>
    </div>

    <nav class="header-nav">
      <RouterLink to="/logs" class="nav-item" :class="{ active: route.path === '/logs' }">
        {{ t('nav.logs') }}
      </RouterLink>
      <span class="nav-item nav-item--soon" :title="t('nav.statistics_soon')">
        {{ t('nav.statistics') }}
      </span>
    </nav>

    <div class="header-right">
      <!-- Docs -->
      <a href="/docs/" target="_blank" rel="noopener" class="icon-btn" :title="t('nav.docs')">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
        </svg>
        <svg width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="opacity:.5">
          <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/>
        </svg>
      </a>

      <!-- Theme toggle -->
      <button class="icon-btn" @click="toggleTheme()"
        :title="theme === 'dark' ? t('nav.toggle_theme_dark') : t('nav.toggle_theme_light')">
        <svg v-if="theme === 'dark'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="5"/>
          <line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
          <line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
        </svg>
        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
      </button>

      <!-- Admin / Settings — icon button, right side -->

      <!-- User / Account menu -->
      <div class="user-menu" ref="userMenuRef">
        <button class="icon-btn" @click="userMenuOpen = !userMenuOpen" :title="t('nav.account')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
          </svg>
        </button>
        <div v-if="userMenuOpen" class="dropdown">
          <RouterLink v-if="auth.isAdmin" to="/admin" class="dropdown-item" @click="userMenuOpen = false">
            {{ t('nav.settings') }}
          </RouterLink>
          <div v-if="auth.isAdmin" class="dropdown-divider"></div>
          <button class="dropdown-item" @click="logout">{{ t('nav.logout') }}</button>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, inject, computed } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import { onClickOutside } from '@/composables/useClickOutside'
import { useLocale } from '@/composables/useLocale'

const { t } = useLocale()

const theme       = inject('theme')
const toggleTheme = inject('toggleTheme')
const logoSrc     = computed(() => theme.value === 'dark' ? '/logo-dark.svg' : '/logo-light.svg')

defineEmits(['toggle-sidebar'])

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

const userMenuOpen = ref(false)
const userMenuRef  = ref(null)
onClickOutside(userMenuRef, () => { userMenuOpen.value = false })

async function logout() {
  try { await api.logout() } catch {}
  auth.clearSession()
  router.push('/login')
}
</script>

<style scoped>
.app-header {
  height: var(--header-height);
  display: flex; align-items: center;
  padding: 0 .75rem;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border);
  position: sticky; top: 0; z-index: 100;
  flex-shrink: 0; gap: .5rem;
}
.header-left  { display: flex; align-items: center; gap: .375rem; flex-shrink: 0; }
.header-nav   { display: flex; align-items: center; gap: .125rem; margin-left: .75rem; flex-shrink: 0; }
.header-right { display: flex; align-items: center; gap: .125rem; margin-left: auto; flex-shrink: 0; }

.logo-link { display: flex; align-items: center; }
.logo-img  { height: 40px; width: auto; }

.nav-item {
  padding: .375rem .7rem; border-radius: var(--radius);
  font-size: .9rem; font-weight: 500; color: var(--text-muted);
  text-decoration: none; white-space: nowrap; transition: background .15s, color .15s;
}
.nav-item:hover  { background: var(--bg-hover); color: var(--text); }
.nav-item.active { color: var(--color-primary); background: var(--bg-selected); }

.icon-btn {
  display: flex; align-items: center; justify-content: center; gap: .2rem;
  padding: .375rem; background: none; border: none; border-radius: var(--radius);
  cursor: pointer; color: var(--text-muted); text-decoration: none;
  transition: background .15s, color .15s; flex-shrink: 0;
}
.icon-btn:hover       { background: var(--bg-hover); color: var(--text); }
.icon-btn--active     { color: var(--color-primary); background: var(--bg-selected); }

.user-menu { position: relative; flex-shrink: 0; }
.dropdown {
  position: absolute; top: calc(100% + 4px); right: 0;
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: var(--radius); box-shadow: var(--shadow);
  min-width: 130px; z-index: 200;
}
.dropdown-item { display: block; width: 100%; text-align: left;
  padding: .5rem .75rem; background: none; border: none;
  cursor: pointer; font-size: .875rem; color: var(--text);
  text-decoration: none;
}
.dropdown-item:hover { background: var(--bg-hover); }
.dropdown-divider { height: 1px; background: var(--border); margin: .25rem 0; }

.filter-toggle { margin-left: .25rem; }

.nav-item--soon {
  color: var(--text-muted); opacity: .4;
  cursor: default; pointer-events: none;
  font-style: italic;
}

@media (max-width: 380px) { .header-nav { display: none; } }
</style>
