import { VueRenderer } from '@tiptap/vue-3';
// Import specific suggestion types
import { Suggestion, type SuggestionOptions, type SuggestionProps, type SuggestionKeyDownProps } from '@tiptap/suggestion';
import Fuse, { type IFuseOptions } from 'fuse.js'; // Correct Fuse import
import type { Jot } from '../../db';
import { db } from '../../db'; // Import db directly for data fetching
import SuggestionList from './SuggestionList.vue';
import { ref, watch, type Ref } from 'vue';

// --- Reactive State for Suggestion Component ---
const suggestionState = ref<{
    items: Jot[];
    command: ((item: Jot) => void) | null;
    clientRect?: DOMRect | null; // Allow null for clientRect
    show: boolean;
}>({ items: [], command: null, show: false });

const setSuggestionState = (newState: Partial<typeof suggestionState.value>) => {
    suggestionState.value = { ...suggestionState.value, ...newState };
};

export const useSharedSuggestionState = () => {
    return suggestionState;
};
// ---

// TipTap Suggestion Plugin Configuration
export const noteSuggestionOptions: Omit<SuggestionOptions, 'editor'> = {
    items: async ({ query }) => {
        // Fetch sorted notes
        const allJots = await db.jots.orderBy('updatedAt').reverse().toArray();

        if (!allJots || allJots.length === 0) {
            return [];
        }

        // Handle empty query: return recent/all notes immediately after trigger char
        if (query.length === 0) { // Check specifically for empty string
            return allJots.slice(0, 10); // Return top 10 most recently updated
        }

        // If query exists, filter with Fuse.js
        const fuseOptions: IFuseOptions<Jot> = {
            keys: ['title'],
            threshold: 0.4,
        };
        const fuse = new Fuse(allJots, fuseOptions);
        const results = fuse.search(query);

        return results.map(result => result.item).slice(0, 10);
    },

    // Keep the original trigger character configuration
    char: '@',
    allowSpaces: true,
    startOfLine: false,

    command: ({ editor, range, props }) => {
        const { id, title } = props.item;
        editor
            .chain()
            .focus()
            .insertContentAt(range, [
                 {
                    type: 'noteLink',
                    attrs: { jotId: id, label: title },
                },
                { type: 'text', text: ' ' }
            ])
            .run();
        setSuggestionState({ show: false });
    },

    render: () => {
        let component: VueRenderer | null = null;

        return {
            onStart: (props: SuggestionProps<Jot>) => {
                component = new VueRenderer(SuggestionList, {
                    props: {
                        items: props.items,
                        command: (item: Jot) => props.command({ ...props, item }),
                        style: { position: 'absolute' },
                    },
                    editor: props.editor,
                });

                const rect = props.clientRect ? props.clientRect() : null;
                setSuggestionState({
                    items: props.items,
                    command: (item: Jot) => props.command({ ...props, item }),
                    clientRect: rect,
                    show: true,
                });

                watch(suggestionState, (newState) => {
                    const newRect = newState.clientRect;
                    component?.updateProps({
                        items: newState.items,
                        style: newRect
                            ? {
                                position: 'absolute',
                                left: `${newRect.left}px`,
                                top: `${newRect.bottom + window.scrollY}px`,
                              }
                            : { position: 'absolute', visibility: 'hidden' },
                    });
                }, { deep: true });
            },

            onUpdate: (props: SuggestionProps<Jot>) => {
                const rect = props.clientRect ? props.clientRect() : null;
                setSuggestionState({
                    items: props.items,
                    clientRect: rect,
                    show: props.items.length > 0,
                });
            },

            onKeyDown: (props: SuggestionKeyDownProps): boolean => {
                 const { view, event } = props;

                 // Handle Escape locally
                 if (event.key === 'Escape') {
                     setSuggestionState({ show: false });
                     return true;
                 }

                 // Delegate ONLY Enter/Tab to the component
                 if (event.key === 'Enter' || event.key === 'Tab') {    
                     // Check component exists and has the method
                     if (component?.ref?.onKeyDown) {
                         return component.ref.onKeyDown({ event });
                     }
                 }

                 return false;
            },

            onExit: () => {
                component?.destroy();
                setSuggestionState({ items: [], command: null, show: false });
            },
        };
    },
};

// Removed placeholder jotService as data is fetched directly now 