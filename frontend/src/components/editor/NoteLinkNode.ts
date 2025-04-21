import { Node, mergeAttributes } from "@tiptap/core";
import { VueNodeViewRenderer } from "@tiptap/vue-3";
import NoteLinkComponent from "./NoteLinkComponent.vue";
import Suggestion from "@tiptap/suggestion";
import { noteSuggestionOptions } from "./suggestionUtils";

export interface NoteLinkOptions {
  HTMLAttributes: Record<string, any>;
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    noteLink: {
      setNoteLink: (attributes: { jotId: string; label: string }) => ReturnType;
    };
  }
}

export const NoteLinkNode = Node.create<NoteLinkOptions>({
  name: "noteLink",
  group: "inline",
  inline: true,
  atom: true,
  selectable: true,

  addOptions() {
    return {
      HTMLAttributes: {
        class: "note-link",
      },
    };
  },

  addAttributes() {
    return {
      jotId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-jot-id"),
        renderHTML: (attributes) => ({ "data-jot-id": attributes.jotId }),
      },
      label: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-label"),
        renderHTML: (attributes) => ({ "data-label": attributes.label }),
      },
    };
  },

  parseHTML() {
    return [{ tag: "span[data-jot-id]" }];
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "span",
      mergeAttributes(this.options.HTMLAttributes, HTMLAttributes),
    ];
  },

  addCommands() {
    return {
      setNoteLink:
        (attributes) =>
        ({ commands }) => {
          return commands.insertContent({
            type: this.name,
            attrs: attributes,
          });
        },
    };
  },

  addNodeView() {
    return VueNodeViewRenderer(NoteLinkComponent);
  },

  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor,
        ...noteSuggestionOptions,
      }),
    ];
  },
});
