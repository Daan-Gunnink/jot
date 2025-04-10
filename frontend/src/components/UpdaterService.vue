<template>
  <div
    v-if="showUpdateBanner"
    class="absolute bottom-0 right-0 border-base-300 border shadow-md rounded-lg mb-8 mr-8 z-30 flex flex-row items-center justify-between py-1 px-2 gap-1"
  >
    <span class="text-base-content text-xs"
      >New version {{ latestVersion }} available!</span
    >
    <div class="flex flex-row items-center gap">
      <button
        class="btn btn-ghost btn-xs"
        @click="downloadAndInstall"
        :disabled="isUpdating"
      >
        {{ isUpdating ? "Updating..." : "Update Now" }}
      </button>
      <button class="btn btn-ghost btn-xs" @click="dismissUpdate">x</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import {
  CheckForUpdates,
  DownloadAndInstallUpdate,
} from "../../wailsjs/go/main/App";

const showUpdateBanner = ref(false);
const updateMessage = ref("");
const latestVersion = ref("");
const isUpdating = ref(false);

async function checkForUpdates() {
  try {
    const result = await CheckForUpdates();
    // If the result contains "Update available" then show the banner
    if (result.includes("Update available")) {
      showUpdateBanner.value = true;
      updateMessage.value = result;

      // Extract the version number from the message
      const match = result.match(/Version ([\d.]+) is available/);
      if (match && match[1]) {
        latestVersion.value = match[1];
      }
    }
  } catch (error) {
    console.error("Error checking for updates:", error);
  }
}

async function downloadAndInstall() {
  isUpdating.value = true;
  try {
    const result = await DownloadAndInstallUpdate();
    updateMessage.value = result;
  } catch (error) {
    console.error("Error updating:", error);
    updateMessage.value = `Update failed: ${error}`;
  } finally {
    isUpdating.value = false;
  }
}

function dismissUpdate() {
  showUpdateBanner.value = false;
}

// Check for updates when the component is mounted
onMounted(() => {
  // Wait a few seconds before checking for updates
  setTimeout(checkForUpdates, 3000);

  // Schedule periodic update checks (every hour)
  setInterval(checkForUpdates, 60 * 60 * 1000);
});
</script>
