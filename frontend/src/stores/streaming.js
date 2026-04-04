import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiFetch } from '../services/api'

export const useStreamingStore = defineStore('streaming', () => {
  const streams = ref([])
  const loading = ref(false)
  let timer = null

  async function load() {
    loading.value = true
    try {
      const data = await apiFetch('/api/status')
      streams.value = Array.isArray(data?.streams) ? data.streams : []
    } finally {
      loading.value = false
    }
  }

  function startPolling(intervalMs = 4000) {
    stopPolling()
    load()
    timer = window.setInterval(load, intervalMs)
  }

  function stopPolling() {
    if (timer) {
      window.clearInterval(timer)
      timer = null
    }
  }

  return { streams, loading, load, startPolling, stopPolling }
})
