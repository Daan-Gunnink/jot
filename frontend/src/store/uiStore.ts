import { defineStore } from "pinia";
import { ref } from "vue";

export const useUIStore = defineStore(
  "ui",
  () => {
    const isSidebarOpen = ref<boolean>(false);

    const toggleSidebar = () => {
      isSidebarOpen.value = !isSidebarOpen.value;
    };

    return {
      isSidebarOpen,
      toggleSidebar,
    };
  },
  {
    persist: {
      key: "ui-state",
    },
  },
);
