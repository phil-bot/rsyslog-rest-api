const BASE = ''

// Serialize params supporting arrays: { Severity: [3,4] } → "Severity=3&Severity=4"
function toQueryString(params) {
  const parts = []
  for (const [key, val] of Object.entries(params)) {
    if (val === undefined || val === null) continue
    if (Array.isArray(val)) {
      val.forEach(v => parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(v)}`))
    } else {
      parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(val)}`)
    }
  }
  return parts.join('&')
}

async function request(path, options = {}) {
  // Read token from the key the auth store actually uses
  const token   = sessionStorage.getItem('rsyslox_token')
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['X-Session-Token'] = token

  const res = await fetch(BASE + path, { ...options, headers })

  // 401 on protected endpoints → clear session and navigate to login.
  // Exception: the login endpoint itself — a 401 there just means wrong password,
  // not an expired session. Let the caller handle it as a normal error.
  if (res.status === 401 && !path.startsWith('/api/admin/login')) {
    sessionStorage.removeItem('rsyslox_token')
    sessionStorage.removeItem('rsyslox_role')
    const redirect = encodeURIComponent(window.location.pathname)
    window.location.href = `/login?redirect=${redirect}`
    throw new Error('Session expired')
  }

  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try { const e = await res.json(); msg = e.message || e.error || msg } catch {}
    throw new Error(msg)
  }

  return res.json()
}

export const api = {
  login:  (password) =>
    request('/api/admin/login', { method: 'POST', body: JSON.stringify({ password }) }),

  logout: () =>
    request('/api/admin/logout', { method: 'POST' }),

  setup: (payload) =>
    request('/api/setup', { method: 'POST', body: JSON.stringify(payload) }),

  getLogs: (params) =>
    request('/api/logs?' + toQueryString(params)),

  getMeta: () =>
    request('/api/meta'),

  getMetaColumn: (column, params = {}) =>
    request('/api/meta/' + column + (Object.keys(params).length ? '?' + toQueryString(params) : '')),

  getConfig: () =>
    request('/api/admin/config'),

  updateConfig: (patch) =>
    request('/api/admin/config', { method: 'PATCH', body: JSON.stringify(patch) }),

  getKeys: () =>
    request('/api/admin/keys'),

  createKey: (name) =>
    request('/api/admin/keys', { method: 'POST', body: JSON.stringify({ name }) }),

  deleteKey: (name) =>
    request('/api/admin/keys/' + encodeURIComponent(name), { method: 'DELETE' }),

  health: () =>
    fetch(BASE + '/health').then(r => r.json()),  // public, no auth, no redirect
}
