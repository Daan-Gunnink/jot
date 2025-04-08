<template>
  <editor-content :editor="editor" class="tiptap" />
</template>

<script lang="ts" setup>
import { Editor, EditorContent } from "@tiptap/vue-3";
import Heading from "@tiptap/extension-heading";
import Bold from "@tiptap/extension-bold";
import History from "@tiptap/extension-history";
import Italic from "@tiptap/extension-italic";
import Blockquote from "@tiptap/extension-blockquote";
import HorizontalRule from "@tiptap/extension-horizontal-rule";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Text from "@tiptap/extension-text";
import ListItem from "@tiptap/extension-list-item";
import OrderedList from "@tiptap/extension-ordered-list";
import BulletList from "@tiptap/extension-bullet-list";
import Placeholder from "@tiptap/extension-placeholder";
import { ref, onMounted, onBeforeUnmount, computed, watch } from "vue";
import { useJotStore } from "../store/jotStore";
import type { JSONContent } from "@tiptap/vue-3";
import type { Jot } from "../store/jotStore";

const editor = ref<Editor>();
const jotStore = useJotStore();

const props = defineProps<{
  jot: Jot;
}>();

let debounceTimeout: number | null = null;

const debounce = (func: (...args: unknown[]) => void, delay: number) => {
  // Clear any existing timeout
  if (debounceTimeout !== null) {
    window.clearTimeout(debounceTimeout);
  }

  // Set a new timeout
  debounceTimeout = window.setTimeout(() => {
    func();
    debounceTimeout = null;
  }, delay);
};

const storeEditorContentWithDebounce = (
  title?: string,
  content?: JSONContent,
) => {
  debounce(() => {
    jotStore.updateJot(props.jot.id, title, content);
  }, 1000);
};

watch(
  () => props.jot.id,
  (newJotId, oldJotId) => {
    if (newJotId !== oldJotId) {
      editor.value?.commands.setContent(props.jot.content);
    }
  },
);

onMounted(() => {
  editor.value = new Editor({
    content: props.jot.content,
    editorProps: {
      attributes: {
        id: "jot-editor",
      },
    },
    extensions: [
      Document,
      Heading,
      Paragraph,
      Text,
      History,
      ListItem,
      OrderedList,
      BulletList,
      Bold,
      Italic,
      Blockquote,
      HorizontalRule,
      TaskItem,
      TaskList,
      Placeholder.configure({
        placeholder: "Write something...",
      }),
    ],
    onUpdate: () => {
      const firstHeading =
        editor.value
          ?.getJSON()
          ?.content?.find((node) => node.type === "heading") ?? undefined;
      storeEditorContentWithDebounce(
        firstHeading?.content?.map((node) => node.text).join(""),
        editor.value?.getJSON(),
      );
    },
  });
});

onBeforeUnmount(() => {
  editor.value?.destroy();
});
</script>

<style scoped>
.tiptap {
  height: 100%;
}

.tiptap :deep(.ProseMirror) {
  height: 100%;
}

.tiptap :deep(.ProseMirror-focused) {
  outline: none;
}

.tiptap :deep(h1) {
  font-size: 2em;
  font-weight: bold;
}

.tiptap :deep(h2) {
  font-size: 1.5em;
  font-weight: bold;
}

.tiptap :deep(h3) {
  font-size: 1.2em;
  font-weight: bold;
}

.tiptap :deep(h4) {
  font-size: 1em;
  font-weight: bold;
}

.tiptap :deep(h5) {
  font-size: 0.8em;
  text-transform: uppercase;
  font-weight: bold;
}

.tiptap :deep(p) {
  font-size: 1em;
  font-weight: normal;
}

.tiptap :deep(strong) {
  font-weight: bold;
}

.tiptap :deep(em) {
  font-style: italic;
}

.tiptap :deep(blockquote) {
  border-left: 4px solid #ddd;
  padding-left: 1em;
}

.tiptap p.is-editor-empty:first-child::before {
  color: #adb5bd;
  content: attr(data-placeholder);
  float: left;
  height: 0;
  pointer-events: none;
}

.tiptap :deep(ol) {
  list-style: decimal;
  padding-left: 1em;
}

.tiptap :deep(ul) {
  list-style: disc;
  padding-left: 1em;
}

.tiptap :deep(ul[data-type="taskList"]) {
  list-style: none;
  padding-left: 0;

  li {
    display: flex;
    align-items: flex-start;

    > label {
      flex: none;
      margin-right: 0.5em;
      user-select: none;

      input {
        height: 1.5rem;
        width: 1.5rem;
        margin-top: 4px;
      }
    }

    > div {
      align-self: center;
      flex: 1 1 auto;
    }
  }
}
</style>
