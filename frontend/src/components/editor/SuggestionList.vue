<template>
  <div
    v-if="items.length > 0"
    ref="suggestionListRef"
    class="min-w-48 z-50 p-1 bg-base-200 text-base-content rounded-md shadow-lg border border-base-300 max-h-60 overflow-y-auto"
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
import type { Jot } from "../../db";
const props = defineProps({
  items: {
    type: Array as PropType<Jot[]>,
    required: true,
  },
  command: {
    type: Function as PropType<(item: Jot) => void>,
    required: true,
  },
  clientRect: {
    type: Object as PropType<DOMRect | undefined>,
    default: undefined,
  },
  style: {
    type: Object as PropType<CSSProperties>,
    default: () => ({
      position: "absolute",
    }),
  },
});

const selectedIndex = ref(0);
const suggestionListRef = ref<HTMLDivElement | null>(null);

const selectItem = (index: number) => {
  const item = props.items[index];
  if (item) {
    props.command(item);
  }
};

const handleGlobalKeyDown = (event: KeyboardEvent) => {
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
    event.preventDefault();
    const newIndex =
      (selectedIndex.value + props.items.length - 1) % props.items.length;
    selectedIndex.value = newIndex;
    nextTick(scrollToSelectedItem);
  } else if (event.key === "ArrowDown") {
    event.preventDefault();
    const newIndex = (selectedIndex.value + 1) % props.items.length;
    selectedIndex.value = newIndex;
    nextTick(scrollToSelectedItem);
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleGlobalKeyDown);
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleGlobalKeyDown);
});

const onKeyDown = ({ event }: { event: KeyboardEvent }): boolean => {
  if (props.items.length === 0) return false;

  if (event.key === "Enter" || event.key === "Tab") {
    event.preventDefault();
    selectItem(selectedIndex.value);
    return true;
  }
  return false;
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
);

defineExpose({
  onKeyDown,
});
</script>
