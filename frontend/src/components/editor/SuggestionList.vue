<template>
  <div
    v-if="items.length > 0"
    ref="suggestionListRef"
    :key="forceRerenderKey"
    class="suggestion-list z-50 p-1 bg-base-200 text-base-content rounded-md shadow-lg border border-base-300 max-h-60 overflow-y-auto"
    :style="style"
  >
    <div
      v-for="(item, index) in items"
      :key="item.id"
      :class="
        index === selectedIndex
          ? 'bg-primary text-primary-content'
          : 'hover:bg-base-300'
      "
      class="suggestion-item p-1 rounded cursor-pointer"
      @click="() => selectItem(index)"
    >
      {{ item.title }}
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  watch,
  nextTick,
  onMounted,
  onBeforeUnmount,
  type PropType,
  type CSSProperties,
} from "vue";
import type { Jot } from "../../db"; // Import Jot type

const props = defineProps({
  items: {
    type: Array as PropType<Jot[]>,
    required: true,
  },
  command: {
    type: Function as PropType<(item: Jot) => void>,
    required: true,
  },
  // Optional: Pass clientRect for positioning if needed
  clientRect: {
    type: Object as PropType<DOMRect | undefined>,
    default: undefined,
  },
  // Optional: Pass style object directly if preferred
  style: {
    type: Object as PropType<CSSProperties>,
    default: () => ({
      position: "absolute" /* Add default positioning if needed */,
    }),
  },
});

const selectedIndex = ref(0);
const suggestionListRef = ref<HTMLDivElement | null>(null);
const forceRerenderKey = ref(0); // Keep force re-render key just in case

const selectItem = (index: number) => {
  const item = props.items[index];
  if (item) {
    props.command(item);
  }
};

// Global Key Listener Logic
const handleGlobalKeyDown = (event: KeyboardEvent) => {
  // Only handle arrows if the list is visible (items exist)
  if (props.items.length === 0) {
    return;
  }

  console.log(
    "Global listener sees:",
    event.key,
    "Current Index:",
    selectedIndex.value,
  );

  if (event.key === "ArrowUp") {
    event.preventDefault(); // Prevent default editor/browser behavior
    const newIndex =
      (selectedIndex.value + props.items.length - 1) % props.items.length;
    console.log("Global ArrowUp: New Index:", newIndex);
    selectedIndex.value = newIndex;
    forceRerenderKey.value++;
    nextTick(scrollToSelectedItem);
  } else if (event.key === "ArrowDown") {
    event.preventDefault(); // Prevent default editor/browser behavior
    const newIndex = (selectedIndex.value + 1) % props.items.length;
    console.log("Global ArrowDown: New Index:", newIndex);
    selectedIndex.value = newIndex;
    forceRerenderKey.value++;
    nextTick(scrollToSelectedItem);
  }
  // Enter/Tab are handled via TipTap delegation
};

onMounted(() => {
  console.log("SuggestionList Mounted - Adding global listener");
  document.addEventListener("keydown", handleGlobalKeyDown);
});

onBeforeUnmount(() => {
  console.log("SuggestionList Unmounted - Removing global listener");
  document.removeEventListener("keydown", handleGlobalKeyDown);
});

// Keep ONLY onKeyDown exposed for Enter/Tab delegation
const onKeyDown = ({ event }: { event: KeyboardEvent }): boolean => {
  console.log("SuggestionList onKeyDown (Enter/Tab):", event.key);
  if (props.items.length === 0) return false; // Ignore if no items

  if (event.key === "Enter" || event.key === "Tab") {
    event.preventDefault();
    selectItem(selectedIndex.value);
    return true; // Handled
  }
  return false; // Not handled by this specific method
};

const scrollToSelectedItem = () => {
  const listEl = suggestionListRef.value;
  if (!listEl) return;
  const selectedEl = listEl?.children[selectedIndex.value] as HTMLElement;
  if (!selectedEl) return;
  const listRect = listEl.getBoundingClientRect();
  const selectedRect = selectedEl.getBoundingClientRect();
  if (selectedRect.bottom > listRect.bottom) {
    listEl.scrollTop += selectedRect.bottom - listRect.bottom;
  } else if (selectedRect.top < listRect.top) {
    listEl.scrollTop -= listRect.top - selectedRect.top;
  }
};

watch(
  () => props.items,
  (newItems, oldItems) => {
    // Reset index only if the list actually changes content or appears/disappears
    if (
      newItems.length !== oldItems?.length ||
      newItems[0]?.id !== oldItems?.[0]?.id
    ) {
      selectedIndex.value = 0;
      if (suggestionListRef.value) {
        suggestionListRef.value.scrollTop = 0;
      }
    }
  },
  { deep: false },
); // Don't need deep watch here

// Expose only onKeyDown for Enter/Tab
defineExpose({
  onKeyDown,
});
</script>

<style scoped>
.suggestion-list {
  min-width: 200px;
  /* Ensure a minimum width */
}

.suggestion-item {
  /* Add any specific styles for items */
}
</style>
