import { Node, mergeAttributes } from '@tiptap/core';
import { VueNodeViewRenderer } from '@tiptap/vue-3';
import NoteLinkComponent from './NoteLinkComponent.vue'; // We'll create this component later
// Import Suggestion utility and our options
import { Suggestion } from '@tiptap/suggestion';
import { noteSuggestionOptions } from './suggestionUtils'; // Assuming options are defined here

export interface NoteLinkOptions {
  HTMLAttributes: Record<string, any>;
}

// Declare module augmentation for accessing node attributes
declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    noteLink: {
      /**
       * Add a note link
       */
      setNoteLink: (attributes: { jotId: string; label: string }) => ReturnType;
    };
  }
}

export const NoteLinkNode = Node.create<NoteLinkOptions>({
  name: 'noteLink',
  group: 'inline', // Behaves like text
  inline: true,
  atom: true, // Treat as a single unit, not editable internally by default
  selectable: true,

  addOptions() {
    return {
      HTMLAttributes: {
        class: 'note-link', // Add a default class for styling
      },
    };
  },

  addAttributes() {
    return {
      jotId: {
        default: null,
        parseHTML: (element) => element.getAttribute('data-jot-id'),
        renderHTML: (attributes) => ({ 'data-jot-id': attributes.jotId }),
      },
      label: {
        default: null,
        parseHTML: (element) => element.getAttribute('data-label'),
        renderHTML: (attributes) => ({ 'data-label': attributes.label }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'span[data-jot-id]' }]; // How to recognize this node in pasted/loaded HTML
  },

  renderHTML({ HTMLAttributes }) {
    // How to render the node back to basic HTML (used for saving/copying)
    // We use a span with data attributes. The actual interactive rendering is done via VueNodeViewRenderer.
    return ['span', mergeAttributes(this.options.HTMLAttributes, HTMLAttributes)];
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
    // Use a Vue component for rich rendering and interaction within the editor
    return VueNodeViewRenderer(NoteLinkComponent);
  },

  // Add the Prosemirror Suggestion plugin via this hook
  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor, // Pass the editor instance
        ...noteSuggestionOptions, // Spread our suggestion config
      }),
    ];
  },
}); 