<template>
  <div class="w-80">
    <Sidebar />
  </div>
  <div class="flex flex-col flex-1 bg-base-300">
    <div class="h-4 bg-base-300"></div>
    <div
      class="absolute top-0 right-0 mr-8 fill-base-300 z-20 pointer-events-auto"
    >
      <div
        class="relative w-12 h-12 bg-base-300 rounded-2xl flex items-center justify-center"
      >
        <button
          class="btn btn-sm btn-ghost rounded-b-xl btn-square"
          @click="toggleDarkMode"
        >
          <SunIcon class="size-4" v-if="isDarkMode" />
          <MoonIcon class="size-4" v-else />
        </button>
      </div>
    </div>

    <div class="grow overflow-y-auto">
      <Editor v-if="jot" :jot="jot" />
      <div
        v-else
        class="flex h-full w-full items-center justify-center flex-col"
      >
        <Placeholder class="w-64 h-64" />
        <div class="text-2xl text-center text-base-content mt-4">
          You don't have any Jots yet
        </div>
        <button @click="createFirstJot()" class="btn btn-base-content">
          Create your first Jot
        </button>
      </div>
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
import { SunIcon, MoonIcon } from "@heroicons/vue/24/outline";
const router = useRouter();
const jotStore = useJotStore();
const isDarkMode = ref(
  localStorage.getItem("theme") === "dark" ||
    (!localStorage.getItem("theme") &&
      window.matchMedia("(prefers-color-scheme: dark)").matches),
);

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

// Toggle between dark and light mode
function toggleDarkMode() {
  isDarkMode.value = !isDarkMode.value;

  if (isDarkMode.value) {
    document.documentElement.setAttribute("data-theme", "dark");
    localStorage.setItem("theme", "dark");
  } else {
    document.documentElement.setAttribute("data-theme", "light");
    localStorage.setItem("theme", "light");
  }
}

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
  // Initialize theme
  if (isDarkMode.value) {
    document.documentElement.setAttribute("data-theme", "dark");
  } else {
    document.documentElement.setAttribute("data-theme", "light");
  }

  // Add keyboard event listener
  document.addEventListener("keydown", handleKeyDown);
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleKeyDown);
});
</script>

<style scoped>
.clip {
  clip-path: inset(0 12px 0 0);
}

.clip-left-polygon {
  clip-path: polygon(
    0% 0%,
    /* top-left */ 10% 10%,
    /* slight curve in */ 8% 90%,
    /* slight curve out */ 0% 100%,
    /* bottom-left */ 100% 100%,
    /* bottom-right */ 100% 0% /* top-right */
  );
}
</style>
