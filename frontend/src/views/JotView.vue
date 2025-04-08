<template>
  <div class="w-80">
    <Sidebar />
  </div>
  <div class="grow px-4 pt-2 overflow-y-auto">
    <Editor v-if="jot" :jot="jot" />
    <div v-else class="flex h-full w-full items-center justify-center flex-col">
      <Placeholder class="w-64 h-64" />
      <div class="text-2xl text-center text-neutral mt-4">
        You don't have any Jots yet
      </div>
      <button @click="createFirstJot()" class="btn btn-neutral mt-4">
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
  computed,
  watch,
  ref,
  onBeforeMount,
} from "vue";
import { useJotStore } from "../store/jotStore";
import Placeholder from "../assets/Placeholder.vue";
import Editor from "../components/Editor.vue";
import type { Jot } from "../store/jotStore";

const router = useRouter();
const jotStore = useJotStore();

const route = useRoute();
const jotId = ref<string | undefined>(route.params.id as string);
const jot = ref<Jot | undefined>(
  jotId.value ? jotStore.getJotById(jotId.value) : undefined,
);

function createFirstJot() {
  const id = jotStore.createJot();
  router.push(`/jot/${id}`);
}

onBeforeMount(() => {
  if (!route.params.id) {
    const latestJot = jotStore.getLatestJot();
    router.push(`/jot/${latestJot?.id}`);
    return;
  }
});

watch(route, () => {
  jotId.value = route.params.id as string;
  jot.value = jotStore.getJotById(jotId.value);
});

// Watch for changes in the jotId and update editor content
watch(
  jotId,
  (newId, oldId) => {
    if (newId !== oldId) {
      if (newId) {
        jot.value = jotStore.getJotById(newId);
      } else {
        jot.value = undefined;
      }
    }
  },
  { immediate: false },
);

const handleKeyDown = (event: KeyboardEvent) => {
  if (
    (event.metaKey || event.ctrlKey) &&
    (event.altKey || event.ctrlKey || event.key === "n") &&
    event.key === "n"
  ) {
    event.preventDefault(); // Prevent default browser behavior
    event.stopPropagation();
    const id = jotStore.createJot();
    router.push(`/jot/${id}`);
  }
};

onMounted(() => {
  window.addEventListener("keydown", handleKeyDown);
});

// Clean up event listeners when component unmounts
onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleKeyDown);
});
</script>
