<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { storeToRefs } from "pinia";
import { useJotStore } from "../../store/jotStore";
import { useUIStore } from "../../store/uiStore";
import { useRouter } from "vue-router";
import Logo from "../../assets/Logo.vue";
import JotItem from "./JotItem.vue";
import { Environment } from "../../../wailsjs/runtime";
import { Bars3Icon, PlusIcon } from "@heroicons/vue/24/outline";
import SearchInput from "./SearchInput.vue";
import { useVirtualizer } from "@tanstack/vue-virtual";

const jotStore = useJotStore();
const uiStore = useUIStore();
const router = useRouter();

const { reactiveJots, searchResults, currentSearchQuery, isLoading } =
  storeToRefs(jotStore);

const system = ref<"darwin" | "windows" | null>(null);

const displayedJots = computed(() => {
  if (searchResults.value !== null) {
    return searchResults.value;
  }
  return reactiveJots.value ?? [];
});

const noResultsFound = computed(() => {
  return jotStore.currentSearchQuery && jotStore.searchResults?.length === 0;
});

const isSidebarOpen = computed(() => uiStore.isSidebarOpen);

// --- Virtualization Setup ---
const parentRef = ref<HTMLElement | null>(null);

const rowVirtualizer = useVirtualizer(
  computed(() => ({
    count: displayedJots.value.length,
    getScrollElement: () => parentRef.value,
    estimateSize: () => 64,
  })),
);

const virtualItems = computed(() => rowVirtualizer.value.getVirtualItems());
const totalSize = computed(() => rowVirtualizer.value.getTotalSize());
// --- End Virtualization Setup ---

function handleJotDelete(id: string) {
  const newJotId = jotStore.deleteJot(id);
  if (newJotId) {
    router.push(`/jot/${newJotId}`);
  } else {
    router.push("/");
  }
}

async function createNewJot() {
  const id = await jotStore.createJot();
  router.push(`/jot/${id}`);
}

function toggleSidebar() {
  uiStore.toggleSidebar();
}

onMounted(async () => {
  try {
    const env = await Environment();
    if (env.platform === "darwin") {
      system.value = "darwin";
    } else if (env.platform === "windows") {
      system.value = "windows";
    }
  } catch (error) {
    const ua = navigator.userAgent;
    const macOSRegex = /(macintosh|macintel|macppc|mac68k|macos)/i;
    system.value = macOSRegex.test(ua) ? "darwin" : null;
  }
});
</script>

<template>
  <div class="flex flex-col gap-2 absolute top-0 left-0 mt-2 ml-2 z-30">
    <button
      v-if="!isSidebarOpen"
      class="btn btn-ghost btn-square"
      @click="toggleSidebar"
    >
      <Bars3Icon class="w-6 h-6 fill-base-content" />
    </button>
    <button
      v-if="!isSidebarOpen"
      class="btn btn-ghost btn-square"
      @click="createNewJot"
    >
      <PlusIcon class="w-6 h-6 fill-base-content" />
    </button>
  </div>

  <div v-if="isSidebarOpen" class="w-80 flex flex-col h-full bg-base-300">
    <div class="p-2 border-b-2 border-b-base-300 flex flex-row items-center">
      <button class="btn btn-ghost btn-square" @click="toggleSidebar">
        <Bars3Icon class="w-6 h-6 fill-base-content" />
      </button>
      <Logo class="w-10 h-10 fill-base-content" />
      <div class="text-2xl text-base-content font-extrabold">Jot</div>
    </div>
    <div class="p-2">
      <SearchInput />
    </div>
    <div ref="parentRef" class="flex-1 overflow-auto ml-2">
      <div
        v-if="isLoading && currentSearchQuery"
        class="text-center p-4 text-base-content/50"
      >
        Searching...
      </div>
      <div
        v-else-if="noResultsFound"
        class="text-center p-4 text-base-content/50"
      >
        No results found for "{{ currentSearchQuery }}"
      </div>
      <div
        v-else
        :style="{
          height: `${totalSize}px`,
          width: '100%',
          position: 'relative',
        }"
      >
        <div
          v-for="virtualRow in virtualItems"
          :key="String(virtualRow.key)"
          :style="{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: `${virtualRow.size}px`,
            transform: `translateY(${virtualRow.start}px)`,
          }"
        >
          <JotItem
            :jot="displayedJots[virtualRow.index]"
            @onDelete="handleJotDelete(displayedJots[virtualRow.index].id)"
          />
        </div>
      </div>
    </div>
    <div class="p-4 border-t-2 border-t-base-300">
      <button @click="createNewJot" class="btn btn-neutral w-full">
        New Jot
        <span
          v-if="system === 'darwin'"
          class="ml-1 text-xs font-bold text-neutral-content/60"
          >⌘ + N</span
        >
        <span
          v-if="system === 'windows'"
          class="ml-1 text-xs font-bold text-neutral-content/60"
          >Ctrl + N</span
        >
      </button>
    </div>
  </div>
</template>
