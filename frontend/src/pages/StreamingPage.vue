<template>
  <q-page class="q-py-md">
    <q-card class="page-card q-pa-md q-pa-lg-md">
      <div class="row items-center q-col-gutter-md q-mb-md">
        <div class="col">
          <div class="text-h5 text-weight-bold">Streaming Status</div>
          <div class="text-caption">
            Current live profiles and playback progress.
          </div>
        </div>
        <div class="col-auto">
          <q-btn color="primary" icon="refresh" label="Refresh" @click="load" />
        </div>
      </div>

      <q-banner
        v-if="streams.length === 0"
        rounded
        class="bg-grey-2 text-grey-8"
      >
        No active streams.
      </q-banner>

      <div v-else class="q-gutter-md">
        <q-card v-for="stream in streams" :key="stream.profileId" flat bordered>
          <q-card-section>
            <div class="row items-center q-col-gutter-md">
              <div class="col">
                <div class="text-h6">{{ stream.profileId }}</div>
                <div class="text-caption">
                  Started: {{ formatDate(stream.startedAt) }}
                </div>
              </div>
              <div class="col-auto">
                <q-chip color="positive" text-color="white" label="LIVE" />
              </div>
            </div>
            <q-separator class="q-my-sm" />
            <div>
              <strong>Now Playing:</strong> {{ stream.currentSong || "-" }}
            </div>
            <div>
              <strong>Progress:</strong> {{ stream.songIndex }} /
              {{ stream.songTotal }}
            </div>
          </q-card-section>
        </q-card>
      </div>
    </q-card>
  </q-page>
</template>

<script setup>
import { onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { storeToRefs } from "pinia";
import { useStreamingStore } from "../stores/streaming";

const router = useRouter();
const $q = useQuasar();
const streamingStore = useStreamingStore();

const { streams } = storeToRefs(streamingStore);

function formatDate(value) {
  if (!value || value === "0001-01-01T00:00:00Z") return "-";
  return new Date(value).toLocaleString();
}

async function load() {
  try {
    await streamingStore.load();
  } catch (error) {
    if (error.message === "Unauthorized") {
      router.push("/login");
      return;
    }
    $q.notify({
      type: "negative",
      message: error.message || "Unable to load status",
    });
  }
}

onMounted(() => streamingStore.startPolling(4000));
onUnmounted(() => streamingStore.stopPolling());
</script>
