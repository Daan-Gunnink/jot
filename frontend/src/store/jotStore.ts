import type { JSONContent } from "@tiptap/vue-3";
import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { StateTree } from "pinia";
import type { PersistenceOptions } from "pinia-plugin-persistedstate";
import { v4 as uuidv4 } from "uuid";

export interface Jot {
  id: string;
  title: string;
  content: JSONContent;
  createdAt: Date;
  updatedAt: Date;
}

// Type for our store state
interface JotState {
  jots: Jot[];
  currentJotId: string | null;
  revisionsMap: Map<string, Jot[]>;
  redoMap: Map<string, Jot[]>;
}

interface SerializedJot {
  id: string;
  title: string;
  content: string; // Content is serialized as a string
  createdAt: string;
  updatedAt: string;
}

// Helper function to limit title length (was previously undefined in the code)
function limitTitleLength(title: string, maxLength = 50): string {
  return title.length > maxLength ? title.substring(0, maxLength) : title;
}

export const useJotStore = defineStore(
  "jot",
  () => {
    const jots = ref<Jot[]>([]);
    const currentJotId = ref<string | null>(null);

    const createJot = (
      title: string = "Untitled Jot",
      content?: JSONContent,
    ): string => {
      const jot: Jot = {
        id: uuidv4(),
        title: title,
        content: content ?? { type: "doc", content: [] },
        createdAt: new Date(),
        updatedAt: new Date(),
      };
      jots.value.push(jot);
      currentJotId.value = jot.id;
      return jot.id;
    };

    const updateJot = (
      id: string,
      title?: string,
      content?: JSONContent,
    ): void => {
      const jot = jots.value.find((jot) => jot.id === id);
      if (jot) {
        jot.title = title ? limitTitleLength(title) : jot.title;
        jot.content = content ?? jot.content;
        jot.updatedAt = new Date();
      }
    };

    const deleteJot = (id: string): string | null => {
      const isCurrentJot = currentJotId.value === id;
      jots.value = jots.value.filter((jot) => jot.id !== id);

      // If we just deleted the current jot, we need to select a new one
      if (isCurrentJot) {
        const latestJot = getLatestJot();
        currentJotId.value = latestJot?.id || null;
        return currentJotId.value;
      }
      return null;
    };

    const getJotById = (id: string): Jot | undefined => {
      return jots.value.find((jot) => jot.id === id);
    };

    const getLatestJot = (): Jot | undefined => {
      if (jots.value.length === 0) return undefined;

      return jots.value.sort(
        (a, b) => b.updatedAt.getTime() - a.updatedAt.getTime(),
      )[0];
    };

    const listJots = (): Jot[] => {
      return jots.value.sort(
        (a, b) => b.updatedAt.getTime() - a.updatedAt.getTime(),
      );
    };

    const setCurrentJotId = (id: string | null): void => {
      currentJotId.value = id;
    };

    const isEmpty = computed(() => jots.value.length === 0);

    return {
      jots,
      currentJotId,
      createJot,
      updateJot,
      deleteJot,
      getJotById,
      getLatestJot,
      listJots,
      setCurrentJotId,
      isEmpty,
    };
  },
  {
    persist: {
      serializer: {
        serialize: (state: StateTree): string => {
          const serializedJots = state.jots.map((jot: Jot) => {
            return {
              ...jot,
              content: JSON.stringify(jot.content),
              createdAt: jot.createdAt.toISOString(),
              updatedAt: jot.updatedAt.toISOString(),
            } as SerializedJot;
          });

          return JSON.stringify({
            jots: serializedJots,
            currentJotId: state.currentJotId,
          });
        },
        deserialize: (serializedState: string): StateTree => {
          const parsedData = JSON.parse(serializedState);
          const jots: Jot[] = [];

          // If latest revisions exist in the parsed data
          if (parsedData.jots) {
            // Loop through each entry in the latest revisions array
            parsedData.jots.forEach((item: SerializedJot) => {
              // Convert the stringified content back to objects and dates back to Date objects
              const processedRevision = {
                ...item,
                content: JSON.parse(item.content), // Parse the stringified content
                createdAt: new Date(item.createdAt),
                updatedAt: new Date(item.updatedAt),
              };

              // Set the processed revision as the only revision for this id in the map
              jots.push(processedRevision);
            });
          }

          return {
            ...parsedData,
            jots,
            currentJotId: parsedData.currentJotId || null,
          };
        },
      },
    } as PersistenceOptions<JotState>,
  },
);
