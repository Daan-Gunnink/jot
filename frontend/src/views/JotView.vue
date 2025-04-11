<template>
  <div class="w-80">
    <Sidebar />
  </div>
  <div class="grow px-4 pt-2 overflow-y-auto">
    <Editor v-if="jot" :jot="jot" />
    <div v-else class="flex h-full w-full items-center justify-center flex-col">
      <Placeholder class="w-64 h-64" />
      <div class="text-2xl text-center text-base-content mt-4">
        You don't have any Jots yet
      </div>
      <button @click="createFirstJot()" class="btn btn-base-content mt-4">
        Create your first Jot
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import Sidebar from "../components/sidebar/Sidebar.vue";
import { useRouter, useRoute } from "vue-router";
import {
  onMounted,
  onBeforeUnmount,
  watch,
  ref,
  computed,
  onBeforeMount,
} from "vue";
import { useJotStore } from "../store/jotStore";
import Placeholder from "../assets/Placeholder.vue";
import Editor from "../components/Editor.vue";
import type { Jot } from "../store/jotStore";

const router = useRouter();
const jotStore = useJotStore();

const route = useRoute();
const jotId = computed(() => {
  return route.params.id ? String(route.params.id) : jotStore.currentJotId;
});

const jot = computed(() => {
  if (!jotId.value) return undefined;
  return jotStore.getJotById(jotId.value as string);
});

function createFirstJot() {
  const id = jotStore.createJot();
  router.push(`/jot/${id}`);
}

function loadJotState() {
  if (jotStore.isEmpty) {
    return;
  }

  if (!jotId.value || !jot.value) {
    // If there's no valid ID or the jot doesn't exist, navigate to the latest jot
    const latestJot = jotStore.getLatestJot();
    if (latestJot) {
      router.push(`/jot/${latestJot.id}`);
    }
    return;
  }

  jotStore.setCurrentJotId(jotId.value as string);
}

onBeforeMount(() => {
  loadJotState();
});

// Watch for route changes to update the state
watch(route, () => {
  loadJotState();
});

// Watch for changes in the store's jots array to handle deletion
watch(
  () => jotStore.jots.length,
  () => {
    loadJotState();
  },
);

// Watch for changes in the currentJotId to update the view
watch(
  () => jotStore.currentJotId,
  (newId) => {
    if (newId && newId !== jotId.value) {
      router.push(`/jot/${newId}`);
    }
  },
);

const handleKeyDown = (event: KeyboardEvent) => {
  if (
    (event.metaKey || event.ctrlKey) &&
    (event.altKey || event.ctrlKey || event.key === "n") &&
    event.key === "n"
  ) {
    event.preventDefault();
    const id = jotStore.createJot();
    router.push(`/jot/${id}`);
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleKeyDown);
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleKeyDown);
});
</script>
