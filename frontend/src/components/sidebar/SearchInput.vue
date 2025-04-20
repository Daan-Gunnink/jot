<template>
  <label class="input gap-2">
    <MagnifyingGlassIcon class="w-5 h-5" />
    <input
      ref="searchInput"
      type="search"
      class="grow peer"
      placeholder="Search"
      v-model="searchQuery"
    />
    <span
      class="flex gap-1 items-center invisible opacity-0 peer-focus:visible peer-focus:opacity-100 duration-300 transition-opacity"
    >
      <kbd class="kbd kbd-sm">⌘</kbd>
      <kbd class="kbd kbd-sm">K</kbd>
    </span>
  </label>
</template>

<script setup lang="ts">
import { MagnifyingGlassIcon } from "@heroicons/vue/24/outline";
import { XMarkIcon } from "@heroicons/vue/20/solid";
import { ref, onMounted, onUnmounted, watch } from "vue";
import { useJotStore } from "../../store/jotStore";

const jotStore = useJotStore();
const searchInput = ref<HTMLInputElement | null>(null);
const searchQuery = ref("");
let debounceTimer: number | undefined;

watch(searchQuery, (newValue) => {
  clearTimeout(debounceTimer);
  debounceTimer = window.setTimeout(() => {
    console.log("Performing search:", newValue);
    jotStore.performSearch(newValue);
  }, 300);
});

const clearSearchInput = () => {
  searchQuery.value = "";
  searchInput.value?.focus();
};

const handleKeyDown = (event: KeyboardEvent) => {
  if ((event.metaKey || event.ctrlKey) && event.key === "k") {
    event.preventDefault();
    searchInput.value?.focus();
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleKeyDown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeyDown);
});
</script>
