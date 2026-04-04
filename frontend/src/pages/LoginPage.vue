<template>
  <q-page class="row items-center justify-center">
    <q-card class="page-card" style="width: min(460px, 96vw)">
      <q-card-section>
        <div class="text-overline text-primary">Secure Access</div>
        <div class="text-h4 text-weight-bold">Sign In With PIN</div>
        <div class="text-caption q-mt-sm">
          Default PIN is <span class="mono">gostream</span> unless changed via
          <span class="mono">--pin</span>.
        </div>
      </q-card-section>

      <q-card-section>
        <q-input
          v-model="pin"
          outlined
          type="password"
          label="PIN Code"
          autocomplete="off"
          @keyup.enter="submit"
        />
      </q-card-section>

      <q-card-actions align="right" class="q-pa-md">
        <q-btn
          :loading="loading"
          color="primary"
          label="Login"
          @click="submit"
        />
      </q-card-actions>
    </q-card>
  </q-page>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const $q = useQuasar();
const authStore = useAuthStore();

const pin = ref("");
const loading = ref(false);

async function submit() {
  loading.value = true;
  try {
    await authStore.login(pin.value);
    router.push("/profiles");
  } catch (error) {
    $q.notify({
      type: "negative",
      message: error.message || "Invalid PIN",
    });
  } finally {
    loading.value = false;
  }
}
</script>
