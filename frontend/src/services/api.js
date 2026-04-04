import { clearAuthToken, getAuthToken } from './auth'

export async function apiFetch(path, options = {}) {
  const token = getAuthToken()
  const headers = {
    ...(options.headers || {}),
    'X-Auth-Token': token
  }

  let body = options.body
  if (body && typeof body === 'object' && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(body)
  }

  const res = await fetch(path, {
    ...options,
    headers,
    body
  })

  if (res.status === 401) {
    clearAuthToken()
    throw new Error('Unauthorized')
  }

  const text = await res.text()
  if (!res.ok) {
    throw new Error(text || 'Request failed')
  }

  if (!text) {
    return null
  }

  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}
