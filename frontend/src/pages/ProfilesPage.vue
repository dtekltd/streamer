<template>
  <q-page class="q-pt-md">
    <div class="row q-col-gutter-md">
      <div class="col-12 col-md-3">
        <q-card class="page-card q-pa-md full-height">
          <div class="text-h6 text-weight-bold">Profiles</div>
          <div class="text-caption q-mb-md">
            Manage stream profile settings and live status.
          </div>

          <q-list bordered separator class="rounded-borders">
            <q-item
              v-for="profile in profiles"
              :key="profile.id"
              clickable
              :active="profile.id === selectedProfileId"
              @click="selectedProfileId = profile.id"
            >
              <q-item-section>
                <q-item-label>{{ profile.name || profile.id }}</q-item-label>
                <q-item-label caption class="mono">{{
                  profile.id
                }}</q-item-label>
              </q-item-section>
              <q-item-section side>
                <q-chip
                  :color="isRunning(profile.id) ? 'positive' : 'grey-7'"
                  text-color="white"
                  :label="isRunning(profile.id) ? 'LIVE' : 'OFFLINE'"
                  size="sm"
                />
              </q-item-section>
            </q-item>
          </q-list>

          <div class="row q-col-gutter-sm q-mt-md">
            <div class="col">
              <q-btn
                color="secondary"
                icon="add"
                label="Add"
                class="full-width"
                @click="addProfile"
              />
            </div>
            <div class="col">
              <q-btn
                color="negative"
                icon="delete"
                label="Delete"
                class="full-width"
                :disable="profiles.length <= 1"
                @click="deleteProfile"
              />
            </div>
          </div>
        </q-card>
      </div>

      <div class="col-12 col-md-4">
        <q-card class="page-card q-pa-md">
          <div class="row items-center q-col-gutter-sm q-mb-md">
            <div class="col">
              <div class="text-h5 text-weight-bold">
                {{ activeProfile?.name || "Profile" }}
              </div>
              <div class="text-caption">
                Edit settings while streaming. Updates apply on next start.
              </div>
            </div>
            <div class="col-auto">
              <q-btn
                flat
                color="primary"
                icon="refresh"
                label="Reload"
                @click="loadAll"
              />
            </div>
          </div>

          <q-form v-if="activeProfile" class="" @submit.prevent="saveSettings">
            <div class="row q-col-gutter-sm">
              <div class="col-12 col-md-6">
                <q-input
                  v-model="activeProfile.name"
                  outlined
                  label="Profile Name"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-input
                  v-model="activeProfile.stream_key"
                  outlined
                  label="Stream Key"
                />
              </div>
            </div>

            <q-input
              v-model="activeProfile.stream_url_template"
              outlined
              label="Stream URL Template"
              class="q-mt-sm"
            />

            <div class="row q-col-gutter-sm q-pt-sm">
              <div class="col-12 col-md-6">
                <q-input
                  v-model="activeProfile.audio_dir"
                  outlined
                  label="Audio Directory"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-select
                  v-model="activeProfile.playlist_order"
                  outlined
                  emit-value
                  map-options
                  :options="playlistOrderOptions"
                  label="Playlist Order"
                />
              </div>
            </div>

            <div class="row q-col-gutter-sm q-pt-sm">
              <div class="col-12 col-md-6">
                <q-select
                  v-model="activeProfile.stream_end_mode"
                  outlined
                  emit-value
                  map-options
                  :options="streamEndOptions"
                  label="Stream End Mode"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-input
                  v-model="activeProfile.end_after_minutes"
                  outlined
                  label="End After Minutes"
                />
              </div>
            </div>

            <q-input
              v-model="activeProfile.video_path"
              outlined
              label="Background Video Path"
              class="q-pt-sm"
            />

            <div class="row q-col-gutter-sm q-pt-sm">
              <div class="col-12 col-md-6">
                <q-toggle
                  v-model="activeProfile.enable_video_audio"
                  label="Enable Video Audio"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-input
                  v-model="activeProfile.video_audio_volume"
                  outlined
                  label="Video Audio Volume"
                />
              </div>
            </div>

            <q-expansion-item
              icon="text_fields"
              label="Text Overlay"
              class="bg-grey-3 q-mt-sm"
              dense
            >
              <div class="q-pa-sm">
                <q-input
                  v-model="activeProfile.font_path"
                  outlined
                  label="Font Path"
                  class="bg-white"
                />
                <div class="row q-col-gutter-sm q-pt-sm">
                  <div class="col-12 col-md-6">
                    <q-input
                      v-model="activeProfile.text_x"
                      outlined
                      label="Text X"
                      class="bg-white"
                    />
                  </div>
                  <div class="col-12 col-md-6">
                    <q-input
                      v-model="activeProfile.text_y"
                      outlined
                      label="Text Y"
                      class="bg-white"
                    />
                  </div>
                </div>
                <q-toggle
                  v-model="activeProfile.enable_playing_label"
                  label="Show Now Playing Label"
                />
                <q-input
                  v-model="activeProfile.now_playing_label"
                  outlined
                  label="Now Playing Label"
                  class="bg-white"
                />
                <q-toggle
                  v-model="activeProfile.enable_next_label"
                  label="Show Next Song Label"
                />
                <q-input
                  v-model="activeProfile.next_song_label"
                  outlined
                  label="Next Song Label"
                  class="bg-white"
                />
              </div>
            </q-expansion-item>

            <q-expansion-item
              icon="terminal"
              label="FFmpeg Args"
              class="bg-grey-3 q-mt-sm"
              dense
            >
              <div class="q-pa-sm">
                <q-input
                  v-model="activeProfile.ffmpeg_args"
                  type="textarea"
                  outlined
                  autogrow
                  label="One argument per line"
                  class="bg-white"
                />
              </div>
            </q-expansion-item>

            <div class="row q-col-gutter-sm q-pt-sm">
              <div class="col-12 col-md-6">
                <q-btn
                  color="primary"
                  icon="save"
                  label="Save Profile"
                  class="full-width"
                  :loading="saving"
                  @click="saveSettings"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-btn
                  color="secondary"
                  icon="playlist_add_check"
                  label="Update Playlist"
                  class="full-width"
                  :loading="updatingPlaylist"
                  @click="updatePlaylist"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-btn
                  color="positive"
                  icon="play_arrow"
                  label="Start"
                  class="full-width"
                  :loading="starting"
                  @click="startStream"
                />
              </div>
              <div class="col-12 col-md-6">
                <q-btn
                  color="negative"
                  icon="stop"
                  label="Stop"
                  class="full-width"
                  :loading="stopping"
                  @click="stopStream"
                />
              </div>
            </div>
          </q-form>
        </q-card>
      </div>
      <div class="col-12 col-md-5">
        <q-card class="page-card q-pa-md">
          <div class="text-subtitle1 text-weight-bold">Playlist Preview</div>
          <div class="text-caption q-mb-sm">
            Current runtime playlist for selected profile.
          </div>
          <q-banner
            v-if="playlist.length === 0"
            class="bg-grey-2 text-grey-8"
            rounded
          >
            No playlist loaded for selected profile.
          </q-banner>
          <q-list v-else bordered separator>
            <q-item v-for="song in playlist" :key="song.display" dense>
              <q-item-section>{{ song.display }}</q-item-section>
              <q-item-section side class="mono">{{
                song.duration
              }}</q-item-section>
            </q-item>
          </q-list>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<script setup>
import { onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { storeToRefs } from "pinia";
import { useProfilesStore } from "../stores/profiles";

const router = useRouter();
const $q = useQuasar();
const profilesStore = useProfilesStore();

const {
  selectedProfileId,
  profiles,
  activeProfile,
  playlist,
  saving,
  starting,
  stopping,
  updatingPlaylist,
} = storeToRefs(profilesStore);

const { isRunning, addProfile, deleteProfile } = profilesStore;

const playlistOrderOptions = [
  { label: "normal", value: "normal" },
  { label: "a-z", value: "a-z" },
  { label: "z-a", value: "z-a" },
  { label: "shuffle", value: "shuffle" },
];

const streamEndOptions = [
  { label: "forever", value: "forever" },
  { label: "duration", value: "duration" },
  { label: "all_songs", value: "all_songs" },
];

function handleError(error, fallback) {
  if (error.message === "Unauthorized") {
    router.push("/login");
    return;
  }
  $q.notify({ type: "negative", message: error.message || fallback });
}

async function loadAll() {
  try {
    await profilesStore.loadAll();
  } catch (error) {
    handleError(error, "Failed to load profile data");
  }
}

async function saveSettings() {
  try {
    await profilesStore.saveSettings();
    $q.notify({ type: "positive", message: "Settings saved" });
  } catch (error) {
    handleError(error, "Unable to save settings");
  }
}

async function startStream() {
  try {
    await profilesStore.startStream();
    $q.notify({ type: "positive", message: "Stream starting" });
  } catch (error) {
    handleError(error, "Failed to start stream");
  }
}

async function stopStream() {
  try {
    await profilesStore.stopStream();
    $q.notify({ type: "positive", message: "Stream stopping" });
  } catch (error) {
    handleError(error, "Failed to stop stream");
  }
}

async function updatePlaylist() {
  try {
    const result = await profilesStore.updatePlaylist();
    $q.notify({
      type: "positive",
      message: `Playlist updated (${result?.songs ?? 0} songs)`,
    });
  } catch (error) {
    handleError(error, "Failed to update playlist");
  }
}

let timer = null;

onMounted(async () => {
  await loadAll();
  timer = window.setInterval(() => {
    profilesStore.loadStatus().catch(() => {});
  }, 4000);
});

onUnmounted(() => {
  if (timer) window.clearInterval(timer);
});
</script>
