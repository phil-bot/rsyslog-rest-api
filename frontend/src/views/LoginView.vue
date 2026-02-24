<template>
  <div class="login-page">
    <div class="login-card">
      <img :src="logoSrc" alt="rsyslox" class="logo" />

      <form @submit.prevent="submit">
        <div class="field">
          <label for="pw">Admin Password</label>
          <input
            id="pw"
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="••••••••••••"
            required
            autofocus
          />
        </div>

        <p v-if="error" class="error-msg">{{ error }}</p>

        <button type="submit" class="btn btn-primary submit-btn" :disabled="loading">
          {{ loading ? 'Logging in…' : 'Log in' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const theme  = inject('theme')
const logoSrc = computed(() => theme?.value === 'dark' ? '/logo-dark.svg' : '/logo-light.svg')

const router = useRouter()
const route  = useRoute()
const auth   = useAuthStore()

const password = ref('')
const error    = ref('')
const loading  = ref(false)

async function submit() {
  error.value   = ''
  loading.value = true
  try {
    const res = await api.login(password.value)
    auth.setSession(res.token, 'admin')
    router.push(route.query.redirect || '/logs')
  } catch (e) {
    error.value = e.body?.message || 'Incorrect password'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  padding: 1.5rem;
}

.login-card {
  width: 100%;
  max-width: 340px;
  padding: 2rem 2rem 2.5rem;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: calc(var(--radius) * 2);
  box-shadow: 0 4px 24px rgba(0,0,0,.07);
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

.logo {
  height: 36px;
  width: auto;
  display: block;
}

form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: .375rem;
}

.field label {
  font-size: .8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .05em;
  color: var(--text-muted);
}

.field input {
  padding: .625rem .75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-mono);
  font-size: .875rem;
  letter-spacing: .05em;
  transition: border-color .15s;
  width: 100%;
}
.field input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(2,132,199,.15);
}

.error-msg {
  font-size: .8rem;
  color: #dc2626;
  background: #fef2f2;
  border: 1px solid #fca5a5;
  border-radius: var(--radius);
  padding: .5rem .625rem;
}
[data-theme="dark"] .error-msg {
  background: #2d1212;
  border-color: #7f1d1d;
  color: #fca5a5;
}

.submit-btn {
  width: 100%;
  justify-content: center;
  padding: .625rem;
  font-size: .9rem;
  margin-top: .25rem;
}
</style>
