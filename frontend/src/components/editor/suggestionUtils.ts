import { VueRenderer } from "@tiptap/vue-3";
import {
  type SuggestionOptions,
  type SuggestionProps,
  type SuggestionKeyDownProps,
} from "@tiptap/suggestion";
import Fuse, { type IFuseOptions } from "fuse.js";
import type { Jot } from "../../db";
import { db } from "../../db";
import SuggestionList from "./SuggestionList.vue";
import { ref, watch } from "vue";
const suggestionState = ref<{
  items: Jot[];
  command: ((item: Jot) => void) | null;
  clientRect?: DOMRect | null;
  show: boolean;
}>({ items: [], command: null, show: false });

const queryRef = ref("");

const setSuggestionState = (
  newState: Partial<typeof suggestionState.value>,
) => {
  suggestionState.value = { ...suggestionState.value, ...newState };
};

export const useSharedSuggestionState = () => {
  return suggestionState;
};

export const noteSuggestionOptions: Omit<SuggestionOptions, "editor"> = {
  items: async ({ query }) => {
    const allJots = await db.jots.orderBy("updatedAt").reverse().toArray();
    queryRef.value = query;

    if (!allJots || allJots.length === 0) {
      return [];
    }

    if (query.length === 0) {
      return allJots.slice(0, 10);
    }

    const fuseOptions: IFuseOptions<Jot> = {
      keys: ["title"],
      threshold: 0.4,
    };
    const fuse = new Fuse(allJots, fuseOptions);
    const results = fuse.search(query);

    return results.map((result) => result.item).slice(0, 10);
  },

  char: "@",
  allowSpaces: true,
  startOfLine: false,

  command: ({ editor, range, props }) => {
    const { id, title } = props.item;

    const newRange = {
      from: range.from,
      to: range.to + queryRef.value.length,
    };

    editor
      .chain()
      .focus()
      .deleteRange(newRange)
      .insertContent([
        {
          type: "noteLink",
          attrs: { jotId: id, label: title },
        },
        { type: "text", text: " " },
      ])
      .run();
    setSuggestionState({ show: false });
  },

  render: () => {
    let component: VueRenderer | null = null;

    let stopWatch: (() => void) | null = null;

    return {
      onStart: (props: SuggestionProps<Jot>) => {
        component = new VueRenderer(SuggestionList, {
          props: {
            items: props.items,
            command: (item: Jot) => props.command({ ...props, item }),
            style: { position: "absolute" },
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

        stopWatch = watch(
          suggestionState,
          (newState) => {
            const newRect = newState.clientRect;
            component?.updateProps({
              items: newState.items,
              command: newState.command,
              style: newRect
                ? {
                    position: "absolute",
                    left: `${newRect.left}px`,
                    top: `${newRect.bottom + window.scrollY}px`,
                  }
                : { position: "absolute", visibility: "hidden" },
            });
          },
          { deep: true },
        );
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
        const { event } = props;

        if (event.key === "Escape") {
          setSuggestionState({ show: false });
          return true;
        }

        if (event.key === "Enter" || event.key === "Tab") {
          if (component?.ref?.onKeyDown) {
            return component.ref.onKeyDown({ event });
          }
        }

        return false;
      },

      onExit: () => {
        stopWatch?.();
        component?.destroy();
        setSuggestionState({ items: [], command: null, show: false });
      },
    };
  },
};
