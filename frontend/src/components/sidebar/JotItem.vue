<template>
  <ContextMenuRoot class="w-full h-full">
    <ContextMenuTrigger as-child class="">
      <div class="w-full h-full hover:bg-base-200 items-center justify-center p-4" :class="{
        'bg-base-100': jotId === props.jot.id
      }">
        <RouterLink :key="props.jot.id" :to="`/jot/${props.jot.id}`">
          <h2 class="text-lg font-bold text-ellipsis overflow-hidden line-clamp-1 select-none">
            {{ jot.title }}
          </h2>
          <p class="text-sm text-base-content/70 text-end select-none">
            {{ formattedDate }}
          </p>
        </RouterLink>
      </div>
    </ContextMenuTrigger>

    <ContextMenuPortal>
      <ContextMenuContent
        class="min-w-[220px] z-30 bg-base-200 border border-base-100 shadow-3xl hover:bg-neutral text-base-content hover:text-primary-content rounded-md p-[5px]"
        :side-offset="5">
        <ContextMenuItem value="Delete"
          class="group text-xs leading-none text-grass11 rounded-[3px] flex items-center h-4 px-1 relative pl-2 select-none outline-none"
          @click="onDisplayDeleteConfirmation">
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenuPortal>
  </ContextMenuRoot>
  <AlertModal :open="displayDeleteConfirmation" :title="`Delete ${props.jot.title}?`"
    :description="`Are you sure you want to delete this Jot? This action cannot be undone.`" actionDescription="Delete"
    @action="handleDelete" @cancel="displayDeleteConfirmation = false" />
</template>

<script setup lang="ts">
import type { Jot } from "../../store/jotStore";
import dayjs from "dayjs";
import { computed, ref } from "vue";
import AlertModal from "../AlertModal.vue";
import {
  ContextMenuRoot,
  ContextMenuTrigger,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuPortal,
} from "reka-ui";
const props = defineProps<{
  jot: Jot;
}>();

import { useRoute } from "vue-router";

const route = useRoute();
const jotId = computed(() => route.params.id);

const displayDeleteConfirmation = ref(false);

const emit = defineEmits<{
  (e: "onDelete"): void;
}>();

function onDisplayDeleteConfirmation() {
  displayDeleteConfirmation.value = true;
}

function handleDelete() {
  emit("onDelete");
  displayDeleteConfirmation.value = false;
}

const formattedDate = computed(() => {
  const today = dayjs();
  const dayJSDate = dayjs(props.jot.updatedAt);

  if (dayJSDate.isSame(today, "day")) {
    return dayJSDate.format("HH:mm");
  }

  if (dayJSDate.isSame(today, "week")) {
    return dayJSDate.format("dddd");
  }

  if (dayJSDate.isSame(today, "month")) {
    return dayJSDate.format("DD MMM");
  }

  if (dayJSDate.isSame(today, "year")) {
    return dayJSDate.format("DD MMM");
  }

  return dayJSDate.format("DD MMM YYYY");
});
</script>
