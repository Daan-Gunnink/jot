import { defineStore } from "pinia";
import { ref } from "vue";

export const useUIStore = defineStore(
  "ui",
  () => {
    const isSidebarOpen = ref<boolean>(false);
    const lastSelectedJotId = ref<string | null>(null);

    const toggleSidebar = () => {
      isSidebarOpen.value = !isSidebarOpen.value;
    };

    const setLastSelectedJotId = (id: string) => {
      lastSelectedJotId.value = id;
    };

    return {
      isSidebarOpen,
      toggleSidebar,
      lastSelectedJotId,
      setLastSelectedJotId,
    };
  },
  {
    persist: {
      key: "ui-state",
    },
  },
);
