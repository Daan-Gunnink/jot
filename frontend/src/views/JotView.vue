<template>
  <Sidebar />
  <div class="flex flex-col flex-1 bg-base-300">
    <div v-if="isSidebarOpen" class="h-4 bg-base-300"></div>
    <div
      v-if="isSidebarOpen"
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
import { useUIStore } from "../store/uiStore";
import Placeholder from "../assets/Placeholder.vue";
import Editor from "../components/Editor.vue";
import { SunIcon, MoonIcon } from "@heroicons/vue/24/outline";
import type { Jot } from "../db";
const router = useRouter();
const jotStore = useJotStore();
const uiStore = useUIStore();

const isDarkMode = ref(
  localStorage.getItem("theme") === "dark" ||
    (!localStorage.getItem("theme") &&
      window.matchMedia("(prefers-color-scheme: dark)").matches),
);

const isSidebarOpen = computed(() => uiStore.isSidebarOpen);

const route = useRoute();
const jotId = computed(() => {
  return route.params.id ? String(route.params.id) : uiStore.lastSelectedJotId;
});

const jot = ref<Jot | undefined>(undefined);

watch(
  () => jotId.value,
  async () => {
    jot.value = await jotStore.getJotById(jotId.value as string);
    uiStore.setLastSelectedJotId(jotId.value as string);
  },
  { immediate: true },
);

async function createFirstJot() {
  const id = await jotStore.createJot();
  router.push(`/jot/${id}`);
}

async function loadJotState() {
  if (jotStore.isEmpty) {
    return;
  }

  if (!jotId.value) {
    const latestJot = await jotStore.getLatestJot();
    if (latestJot) {
      router.push(`/jot/${latestJot.id}`);
    }
    return;
  } else {
    router.push(`/jot/${jotId.value}`);
  }

  jotStore.setCurrentJotId(jotId.value as string);
}

onBeforeMount(async () => {
  await loadJotState();
});

watch(route, async () => {
  await loadJotState();
});

watch(
  () => jotStore.reactiveJots?.length,
  () => {
    loadJotState();
  },
);

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

const handleKeyDown = async (event: KeyboardEvent) => {
  if (
    (event.metaKey || event.ctrlKey) &&
    (event.altKey || event.ctrlKey || event.key === "n") &&
    event.key === "n"
  ) {
    event.preventDefault();
    const id = await jotStore.createJot();
    router.push(`/jot/${id}`);
  }

  if ((event.metaKey || event.ctrlKey) && event.key === "/") {
    event.preventDefault();
    uiStore.toggleSidebar();
  }
};

onMounted(() => {
  if (isDarkMode.value) {
    document.documentElement.setAttribute("data-theme", "dark");
  } else {
    document.documentElement.setAttribute("data-theme", "light");
  }

  document.addEventListener("keydown", handleKeyDown);
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleKeyDown);
});
</script>
