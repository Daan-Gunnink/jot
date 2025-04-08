<script setup lang="ts">
import {
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogOverlay,
  AlertDialogPortal,
  AlertDialogRoot,
  AlertDialogTitle,
} from 'reka-ui'


const emit = defineEmits<{
  (e: 'action'): void
  (e: 'cancel'): void
}>()

defineProps<{
  open: boolean
  title: string
  description: string
  actionDescription?: string
}>()


</script>

<template>
  <AlertDialogRoot :open="open">
    <AlertDialogPortal>
      <AlertDialogOverlay class="bg-neutral-100/90 dark:bg-blackA9 data-[state=open]:animate-overlayShow fixed inset-0 z-30" />
      <AlertDialogContent
        class="z-[100] text-sm data-[state=open]:animate-contentShow fixed top-[50%] left-[50%] max-h-[85vh] w-[90vw] max-w-[500px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-base-100 p-[25px] shadow-3xl focus:outline-none"
      >
        <AlertDialogTitle class="text-mauve12 m-0 text-[17px] font-semibold">
          {{ title }}
        </AlertDialogTitle>
        <AlertDialogDescription class="text-mauve11 mt-4 mb-5 text-sm leading-normal">
          {{ description }}
        </AlertDialogDescription>
        <div class="flex justify-end gap-4">
          <AlertDialogCancel
            class="btn btn-ghost"
            @click="emit('cancel')"
          >
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            class="btn btn-error btn-soft"
            @click="emit('action')"
          >
            {{ actionDescription }}
          </AlertDialogAction>
        </div>
      </AlertDialogContent>
    </AlertDialogPortal>
  </AlertDialogRoot>
</template>