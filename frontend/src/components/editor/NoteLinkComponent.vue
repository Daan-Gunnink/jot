<template>
  <NodeViewWrapper
    as="span"
    :class="[
      'note-link-component',
      'text-primary',
      'hover:underline',
      'cursor-pointer',
      'py-0.5',
      'rounded',
      { 'ProseMirror-selectednode': selected },
    ]"
    @click="handleClick"
  >
    {{ title }}
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { nodeViewProps, NodeViewWrapper } from "@tiptap/vue-3";
import { useRouter } from "vue-router";
import { useJotStore } from "../../store/jotStore";
import { computed } from "vue";
const props = defineProps(nodeViewProps);
const router = useRouter();
const jotStore = useJotStore();

const title = computed(() => {
  return jotStore?.jotsTitleMap?.get(props.node.attrs.jotId) ?? "";
});

const handleClick = () => {
  const jotId = props.node.attrs.jotId;
  if (jotId) {
    console.log(`Navigating to jot: /jot/${jotId}`);
    router.push(`/jot/${jotId}`);
  }
};
</script>
