import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/logs' },
    {
      path: '/setup',
      name: 'setup',
      component: () => import('@/views/SetupView.vue'),
      meta: { public: true },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/logs',
      name: 'logs',
      component: () => import('@/views/LogsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAdmin: true },
    },
    { path: '/:pathMatch(.*)*', redirect: '/logs' },
  ],
})

// Fetches /health and returns whether the server is in setup mode.
// Result is NOT cached — the health endpoint is cheap and we need
// the live value after the setup wizard writes config.toml.
async function isSetupMode() {
  try {
    const res = await fetch('/health')
    const data = await res.json()
    return data.setup_mode === true
  } catch {
    return false
  }
}

router.beforeEach(async (to) => {
  const setupMode = await isSetupMode()

  if (setupMode) {
    if (to.name !== 'setup') return { name: 'setup' }
    return true
  }

  // Normal operation
  if (to.meta.public) return true

  const auth = useAuthStore()
  if (!auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.requiresAdmin && !auth.isAdmin) {
    return { name: 'logs' }
  }

  return true
})

export default router
