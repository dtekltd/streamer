const AUTH_TOKEN_KEY = 'streamer_auth_token'

export function getAuthToken() {
  return window.localStorage.getItem(AUTH_TOKEN_KEY) || ''
}

export function hasAuthToken() {
  return getAuthToken() !== ''
}

export function setAuthToken(token) {
  if (token) {
    window.localStorage.setItem(AUTH_TOKEN_KEY, token)
  }
}

export function clearAuthToken() {
  window.localStorage.removeItem(AUTH_TOKEN_KEY)
}

export async function login(pin) {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pin })
  })

  if (!res.ok) {
    let message = 'Login failed'
    try {
      const data = await res.json()
      message = data.error || message
    } catch {
      message = await res.text()
    }
    throw new Error(message)
  }

  const data = await res.json()
  setAuthToken(data.token)
  return data
}

export async function logout() {
  const token = getAuthToken()
  await fetch('/api/auth/logout', {
    method: 'POST',
    headers: {
      'X-Auth-Token': token
    }
  })
}
