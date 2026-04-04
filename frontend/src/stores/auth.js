import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  getAuthToken,
  setAuthToken,
  clearAuthToken
} from '../services/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(getAuthToken())
  const isAuthenticated = computed(() => token.value !== '')

  function initFromStorage() {
    token.value = getAuthToken()
  }

  async function login(pin) {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pin })
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Invalid PIN')
    }
    const data = await res.json()
    setAuthToken(data.token)
    token.value = data.token
  }

  async function logout() {
    const current = token.value
    clearAuthToken()
    token.value = ''
    if (current) {
      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: { 'X-Auth-Token': current }
      }).catch(() => {})
    }
  }

  return { token, isAuthenticated, initFromStorage, login, logout }
})
