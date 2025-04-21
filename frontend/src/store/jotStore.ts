import type { JSONContent } from "@tiptap/vue-3";
import { defineStore } from "pinia";
import { ref, computed, type Ref } from "vue";
import { v4 as uuidv4 } from "uuid";
import type { Jot } from "../db";
import * as jotService from "../services/jotService";
import Fuse, { type IFuseOptions } from "fuse.js";

// Helper function to limit title length (keep if still used, maybe move to utils)
// function limitTitleLength(title: string, maxLength = 50): string {
//   return title.length > maxLength ? title.substring(0, maxLength) : title;
// }
// --> Assuming title length limiting is handled elsewhere or not needed for now

export const useJotStore = defineStore(
  "jot",
  () => {
    const currentJotId = ref<string | null>(null);
    const isLoading = ref<boolean>(true);
    const currentSearchQuery = ref<string>("");
    const searchResults = ref<Jot[] | null>(null);

    const initializeStore = async () => {
      isLoading.value = true;
      try {
        const latestJot = await jotService.getLatestJot();
        currentJotId.value = latestJot?.id ?? null;
      } catch (error) {
        console.error("Failed to initialize jot store:", error);
      } finally {
        isLoading.value = false;
      }
    };

    const reactiveJots = jotService.listJotsReactive();

    const migrateJots = async (): Promise<void> => {
      if (localStorage.getItem("dexieMigrationCompleted")) {
        return;
      }

      const rawData = localStorage.getItem("jot");

      if (!rawData) {
        return;
      }

      try {
        const jotsData = JSON.parse(rawData) as {
          currentJotId: string;
          jots: {
            title: string;
            content: string;
            id: string;
          }[];
        };
        for (const jot of jotsData.jots) {
          const parsedContent = JSON.parse(jot.content);
          await jotService.addJot(
            {
              title: jot.title,
              content: parsedContent,
            },
            jot.id,
          );
        }
      } catch (error) {
        console.error("Failed to migrate jots:", error);
        return;
      }

      localStorage.setItem("dexieMigrationCompleted", "true");
    };

    const createJot = async (
      title: string = "Untitled Jot",
      content?: JSONContent,
    ): Promise<string> => {
      isLoading.value = true;
      const newId = uuidv4(); // Generate ID here
      const jotData = {
        title: title,
        content: content ?? { type: "doc", content: [] },
      };
      try {
        const newJot = await jotService.addJot(jotData, newId);
        currentJotId.value = newJot.id;
        return newJot.id;
      } catch (error) {
        console.error("Failed to create jot:", error);
        throw error; // Re-throw for component handling
      } finally {
        isLoading.value = false;
      }
    };

    const updateJot = async (
      id: string,
      title?: string,
      content?: JSONContent,
    ): Promise<void> => {
      // Avoid unnecessary updates if title/content are undefined
      if (title === undefined && content === undefined) {
        return;
      }
      isLoading.value = true;
      const updateData = { title, content };
      try {
        const updatedJot = await jotService.updateJot(id, updateData);
        if (!updatedJot) {
          console.warn(`Jot with id ${id} not found for update.`);
        }
        // No need to update local state, Dexie handles it
      } catch (error) {
        console.error("Failed to update jot:", error);
        throw error;
      } finally {
        isLoading.value = false;
      }
    };

    const deleteJot = async (id: string): Promise<string | null> => {
      isLoading.value = true;
      const isCurrentJot = currentJotId.value === id;
      try {
        await jotService.deleteJot(id);

        // If we deleted the current jot, select the new latest one
        if (isCurrentJot) {
          const latestJot = await jotService.getLatestJot();
          currentJotId.value = latestJot?.id ?? null;
          return currentJotId.value;
        }
        return null; // Return null if a different jot was deleted
      } catch (error) {
        console.error("Failed to delete jot:", error);
        throw error;
      } finally {
        isLoading.value = false;
      }
    };

    // Getters now call the service
    const getJotById = async (id: string): Promise<Jot | undefined> => {
      if (!id) return undefined;
      // Consider adding caching here if called frequently
      return await jotService.getJotById(id);
    };

    // Use computed property for current Jot object based on ID and reactive list
    const currentJot: Ref<Jot | undefined> = computed(() => {
      if (!currentJotId.value || !reactiveJots.value) {
        return undefined;
      }
      return reactiveJots.value.find((jot) => jot.id === currentJotId.value);
    });

    const setCurrentJotId = (id: string | null): void => {
      currentJotId.value = id;
    };

    const isEmpty = computed(
      () => !reactiveJots.value || reactiveJots.value.length === 0,
    );

    const getLatestJot = async (): Promise<Jot | undefined> => {
      return await jotService.getLatestJot();
    };

    // --- Search Logic ---
    const performSearch = async (query: string) => {
      currentSearchQuery.value = query.trim();
      if (!currentSearchQuery.value) {
        searchResults.value = null; // Clear results if query is empty
        isLoading.value = false; // Ensure loading is false
        return;
      }

      isLoading.value = true;
      try {
        const allJots = reactiveJots.value;

        if (!allJots || allJots.length === 0) {
          searchResults.value = []; // No jots to search
          isLoading.value = false;
          return;
        }

        // 2. Configure Fuse.js
        const fuseOptions: IFuseOptions<Jot> = {
          keys: [
            { name: "title", weight: 0.7 },
            { name: "textContent", weight: 0.6 },
          ],
          includeScore: false,
          threshold: 0.4,
        };

        const fuse = new Fuse(allJots, fuseOptions);

        // 4. Perform the search
        const results = fuse.search(currentSearchQuery.value);

        searchResults.value = results.map((result) => result.item);
      } catch (error) {
        console.error("Failed to perform search:", error);
        searchResults.value = []; // Indicate error or empty results
      } finally {
        isLoading.value = false;
      }
    };

    // Action to clear the search
    const clearSearch = () => {
      currentSearchQuery.value = "";
      searchResults.value = null;
      isLoading.value = false; // Ensure loading is reset
    };

    const jotsTitleMap = computed(() => {
      const map = new Map<string, string>();
      reactiveJots.value?.forEach((jot) => {
        if (jot && jot.id && jot.title !== undefined && jot.title !== null) {
          map.set(jot.id, jot.title);
        }
      });

      return map;
    });

    // --- Initialization Call ---
    // Call initializeStore when the store is created
    initializeStore();

    return {
      // State
      currentJotId,
      isLoading,
      reactiveJots, // Expose reactive jots
      currentSearchQuery,
      searchResults,

      // Getters
      currentJot,
      isEmpty,
      getLatestJot, // Keep if needed elsewhere
      jotsTitleMap,

      // Actions
      migrateJots,
      createJot,
      updateJot,
      deleteJot,
      getJotById,
      setCurrentJotId,
      performSearch,
      clearSearch,
      // Don't forget initializeStore if it needs to be called externally
    };
  },
  // Pinia Persist configuration (if you were using it)
  // {
  //   persist: { ... }
  // }
);
