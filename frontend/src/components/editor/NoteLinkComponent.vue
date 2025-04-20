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
    {{ node.attrs.label }}
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3';
import { useRouter } from 'vue-router';

const props = defineProps(nodeViewProps);
const router = useRouter();

const handleClick = () => {
  const jotId = props.node.attrs.jotId;
  if (jotId) {
    console.log(`Navigating to jot: /jot/${jotId}`);
    router.push(`/jot/${jotId}`);
  }
};
</script>