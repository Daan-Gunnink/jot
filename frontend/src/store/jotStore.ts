import type { JSONContent } from "@tiptap/vue-3";
import { defineStore, storeToRefs } from "pinia";
import { ref, computed, type Ref } from "vue";
import { v4 as uuidv4 } from "uuid";
// Import Jot type from db.ts instead of defining it here
import type { Jot } from "../db";
import * as jotService from "../services/jotService";
// Remove unused Pinia persistence types
// import type { StateTree } from "pinia";
// import type { PersistenceOptions } from "pinia-plugin-persistedstate";
import Fuse, { type IFuseOptions } from 'fuse.js'; // <-- Fix import

// Remove local Jot interface definition - use the one from db.ts

// Remove JotState interface if not needed elsewhere, Pinia infers state type

// Remove SerializedJot interface, not needed

// Helper function to limit title length (keep if still used, maybe move to utils)
// function limitTitleLength(title: string, maxLength = 50): string {
//   return title.length > maxLength ? title.substring(0, maxLength) : title;
// }
// --> Assuming title length limiting is handled elsewhere or not needed for now

export const useJotStore = defineStore(
  "jot",
  () => {
    // Remove the local jots ref - data comes from Dexie service
    // const jots = ref<Jot[]>([]);
    const currentJotId = ref<string | null>(null);
    const isLoading = ref<boolean>(true); // Add loading state
    const currentSearchQuery = ref<string>(''); // State for the search input
    const searchResults = ref<Jot[] | null>(null); // State for search results (null = no search active)

    // --- Initialization ---
    // Action to load initial state from Dexie
    const initializeStore = async () => {
      isLoading.value = true;
      try {
        const latestJot = await jotService.getLatestJot();
        currentJotId.value = latestJot?.id ?? null;
        // If no jots exist, maybe create an initial one?
        if (!currentJotId.value) {
           console.log("No initial jots found in Dexie.");
           // Optionally create a default jot here if needed
           // await createJot("My First Jot");
        }
      } catch (error) {
        console.error("Failed to initialize jot store:", error);
        // Handle initialization error (e.g., show message to user)
      } finally {
        isLoading.value = false;
      }
    };

    // --- Reactive Data (Optional, requires dependencies) ---
    // Provides a reactive list directly from Dexie liveQuery
    // Requires npm install @vueuse/rxjs rxjs
    const reactiveJots = jotService.listJotsReactive();


    // --- Actions (now mostly async wrappers for jotService) ---


    const migrateJots = async (): Promise<void> => {
      if (localStorage.getItem("dexieMigrationCompleted")) {
        return;
      }

      const rawData = localStorage.getItem("jot");

      if (!rawData) {
        return;
      }

      try{
        const jotsData = JSON.parse(rawData) as {
          currentJotId: string;
          jots: {
            title: string;
            content: string,
            id: string
          }[];
        };
        for (const jot of jotsData.jots) {
          const parsedContent = JSON.parse(jot.content);
          await jotService.addJot({
            title: jot.title,
            content: parsedContent,
          }, jot.id);
        }
      } catch (error) {
        console.error("Failed to migrate jots:", error);
        return
      }


      localStorage.setItem("dexieMigrationCompleted", "true");
      localStorage.removeItem("jot");
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
        return reactiveJots.value.find(jot => jot.id === currentJotId.value);
    });


    const setCurrentJotId = (id: string | null): void => {
      currentJotId.value = id;
    };

    const isEmpty = computed(() => !reactiveJots.value || reactiveJots.value.length === 0);

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
            // Ensure reactiveJots has loaded its initial value if needed
            // Depending on how listJotsReactive() works, you might need
            // to ensure it's populated before searching. For simplicity,
            // we assume it holds the current state.
            const allJots = reactiveJots.value;

            if (!allJots || allJots.length === 0) {
                searchResults.value = []; // No jots to search
                isLoading.value = false;
                return;
            }

            // 2. Configure Fuse.js
            const fuseOptions: IFuseOptions<Jot> = {
                keys: [
                    { name: 'title', weight: 0.7 }, // Give title slightly more weight
                    { name: 'textContent', weight: 0.3 }
                ],
                includeScore: false, // Set to true if you want score for ranking/debugging
                threshold: 0.4, // Adjust this threshold (0=exact, 1=match anything)
                // Other options like distance, minMatchCharLength can be added here
                // See Fuse.js documentation: https://fusejs.io/api/options.html
            };

            // 3. Create Fuse instance
            // For large datasets, consider creating/updating this instance less frequently
            const fuse = new Fuse(allJots, fuseOptions);

            // 4. Perform the search
            const results = fuse.search(currentSearchQuery.value);

            // 5. Update the searchResults state
            // Results contain { item: Jot, refIndex: number, score?: number }
            // We only need the item itself
            searchResults.value = results.map(result => result.item);

        } catch (error) {
            console.error("Failed to perform search:", error);
            searchResults.value = []; // Indicate error or empty results
        } finally {
            isLoading.value = false;
        }
    };

    // Action to clear the search
    const clearSearch = () => {
        currentSearchQuery.value = '';
        searchResults.value = null;
        isLoading.value = false; // Ensure loading is reset
    };

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

