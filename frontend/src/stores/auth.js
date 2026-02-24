import { ref, computed } from 'vue'

// Simple reactive auth state.
// No localStorage — token lives in memory only (more secure, invalidated on tab close).
const token = ref(sessionStorage.getItem('rsyslox_token') || null)
const role = ref(sessionStorage.getItem('rsyslox_role') || null)

export function useAuthStore() {
  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => role.value === 'admin')

  function setSession(newToken, newRole) {
    token.value = newToken
    role.value = newRole
    sessionStorage.setItem('rsyslox_token', newToken)
    sessionStorage.setItem('rsyslox_role', newRole)
  }

  function clearSession() {
    token.value = null
    role.value = null
    sessionStorage.removeItem('rsyslox_token')
    sessionStorage.removeItem('rsyslox_role')
  }

  function getToken() {
    return token.value
  }

  return { isAuthenticated, isAdmin, setSession, clearSession, getToken }
}
