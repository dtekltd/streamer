<template>
  <q-layout view="hHh lpR fFf" class="app-shell">
    <q-header v-if="showHeader" elevated class="topbar">
      <q-toolbar>
        <q-toolbar-title class="brand-title">Go Stream Control</q-toolbar-title>
        <q-tabs
          dense
          indicator-color="accent"
          active-color="accent"
          class="text-white"
        >
          <q-route-tab to="/profiles" label="Profiles" />
          <q-route-tab to="/streaming" label="Streaming" />
        </q-tabs>
        <q-space />
        <q-btn flat color="white" label="Logout" @click="onLogout" />
      </q-toolbar>
    </q-header>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup>
import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "./stores/auth";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const showHeader = computed(() => route.path !== "/login");

async function onLogout() {
  await authStore.logout();
  router.push("/login");
}
</script>
