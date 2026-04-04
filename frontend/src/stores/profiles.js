import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiFetch } from '../services/api'

function createNewProfile(count) {
  const id = `profile-${Date.now()}`
  return {
    id,
    name: `Profile ${count + 1}`,
    stream_key: '',
    stream_url_template: 'rtmp://10.16.0.165:1935/live/%s',
    audio_dir: '',
    enable_video_audio: false,
    video_audio_volume: '1.0',
    ffmpeg_args: '',
    playlist_order: 'normal',
    stream_end_mode: 'forever',
    end_after_minutes: '60',
    video_path: '',
    font_path: '',
    text_x: '30',
    text_y: 'h-th-30',
    enable_playing_label: true,
    now_playing_label: 'Now Playing:',
    enable_next_label: true,
    next_song_label: 'Next song:'
  }
}

export const useProfilesStore = defineStore('profiles', () => {
  const settings = ref({ saved: false, selected_profile: 'default', profiles: [] })
  const runtimeStatus = ref({ current: {}, streams: [] })
  const selectedProfileId = ref('default')

  const saving = ref(false)
  const starting = ref(false)
  const stopping = ref(false)
  const updatingPlaylist = ref(false)

  const profiles = computed(() => settings.value.profiles || [])

  const activeProfile = computed(() =>
    profiles.value.find(p => p.id === selectedProfileId.value) || null
  )

  const playlist = computed(() => {
    if (
      !runtimeStatus.value.current ||
      runtimeStatus.value.current.profileId !== selectedProfileId.value
    ) {
      return []
    }
    return runtimeStatus.value.current.songs || []
  })

  function isRunning(profileId) {
    return (runtimeStatus.value.streams || []).some(
      s => s.profileId === profileId && s.isRunning
    )
  }

  async function loadSettings() {
    const data = await apiFetch('/api/settings')
    settings.value = data
    if (!selectedProfileId.value) {
      selectedProfileId.value = data.selected_profile || 'default'
    }
    if (
      !profiles.value.some(p => p.id === selectedProfileId.value) &&
      profiles.value.length > 0
    ) {
      selectedProfileId.value = profiles.value[0].id
    }
  }

  async function loadStatus() {
    const data = await apiFetch(
      `/api/status?profileId=${encodeURIComponent(selectedProfileId.value)}`
    )
    runtimeStatus.value = data
  }

  async function loadAll() {
    await loadSettings()
    await loadStatus()
  }

  async function saveSettings() {
    if (!activeProfile.value) return
    saving.value = true
    try {
      settings.value.selected_profile = selectedProfileId.value
      await apiFetch('/api/settings', { method: 'POST', body: settings.value })
      await loadSettings()
      return { success: true }
    } finally {
      saving.value = false
    }
  }

  async function startStream() {
    if (!activeProfile.value) return
    starting.value = true
    try {
      const p = activeProfile.value
      await apiFetch('/api/start', {
        method: 'POST',
        body: {
          profileId: p.id,
          streamKey: p.stream_key,
          streamUrlTemplate: p.stream_url_template,
          videoPath: p.video_path,
          enableVideoAudio: p.enable_video_audio,
          videoAudioVolume: p.video_audio_volume,
          audioDir: p.audio_dir,
          ffmpegArgs: p.ffmpeg_args,
          playlistOrder: p.playlist_order,
          streamEndMode: p.stream_end_mode,
          endAfterMinutes: p.end_after_minutes,
          fontPath: p.font_path,
          textX: p.text_x,
          textY: p.text_y,
          enablePlayingLabel: p.enable_playing_label,
          nowPlayingLabel: p.now_playing_label,
          enableNextLabel: p.enable_next_label,
          nextSongLabel: p.next_song_label
        }
      })
      await loadStatus()
      return { success: true }
    } finally {
      starting.value = false
    }
  }

  async function stopStream() {
    if (!activeProfile.value) return
    stopping.value = true
    try {
      await apiFetch('/api/stop', {
        method: 'POST',
        body: { profileId: selectedProfileId.value }
      })
      await loadStatus()
      return { success: true }
    } finally {
      stopping.value = false
    }
  }

  async function updatePlaylist() {
    if (!activeProfile.value) return
    updatingPlaylist.value = true
    try {
      const p = activeProfile.value
      const data = await apiFetch('/api/update-playlist', {
        method: 'POST',
        body: {
          profileId: p.id,
          playlistOrder: p.playlist_order,
          audioDir: p.audio_dir,
          streamEndMode: p.stream_end_mode,
          endAfterMinutes: p.end_after_minutes
        }
      })
      runtimeStatus.value.current = {
        ...(runtimeStatus.value.current || {}),
        profileId: p.id,
        songs: data.playlist || []
      }
      return { success: true, songs: data.songs || 0 }
    } finally {
      updatingPlaylist.value = false
    }
  }

  function addProfile() {
    const profile = createNewProfile(profiles.value.length)
    settings.value.profiles.push(profile)
    selectedProfileId.value = profile.id
  }

  function deleteProfile() {
    if (profiles.value.length <= 1 || !activeProfile.value) return
    settings.value.profiles = settings.value.profiles.filter(
      p => p.id !== selectedProfileId.value
    )
    if (!settings.value.profiles.some(p => p.id === selectedProfileId.value)) {
      selectedProfileId.value = settings.value.profiles[0].id
    }
  }

  return {
    settings,
    runtimeStatus,
    selectedProfileId,
    saving,
    starting,
    stopping,
    updatingPlaylist,
    profiles,
    activeProfile,
    playlist,
    isRunning,
    loadSettings,
    loadStatus,
    loadAll,
    saveSettings,
    startStream,
    stopStream,
    updatePlaylist,
    addProfile,
    deleteProfile
  }
})
