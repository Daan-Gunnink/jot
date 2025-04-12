<script setup lang="ts">
import { computed, onBeforeMount, onMounted, ref } from "vue";
import { useJotStore } from "../../store/jotStore";
import { useRouter } from "vue-router";
import Logo from "../../assets/Logo.vue";
import JotItem from "./JotItem.vue";
import { Environment } from "../../../wailsjs/runtime";

const jotStore = useJotStore();
const router = useRouter();
const jots = computed(() => jotStore.listJots());
const system = ref<"darwin" | "windows" | null>(null);

function handleJotDelete(id: string) {
  const newJotId = jotStore.deleteJot(id);
  if (newJotId) {
    router.push(`/jot/${newJotId}`);
  } else {
    router.push("/");
  }
}

function createNewJot() {
  const id = jotStore.createJot();
  router.push(`/jot/${id}`);
}

onMounted(async () => {
  try {
    const env = await Environment();
    if (env.platform === "darwin") {
      system.value = "darwin";
    } else if (env.platform === "windows") {
      system.value = "windows";
    }
  } catch (error) {
    const ua = navigator.userAgent;
    const macOSRegex = /(macintosh|macintel|macppc|mac68k|macos)/i;
    system.value = macOSRegex.test(ua) ? "darwin" : null;
  }
});
</script>

<template>
  <div class="w-80 border-r-2 border-r-base-300 flex flex-col h-full rounden bg-base-300">
    <div class="p-2 border-b-2 border-b-base-300 flex flex-row items-center">
      <Logo class="w-10 h-10 fill-base-content" />
      <div class="text-2xl text-base-content font-extrabold">Jot</div>
    </div>
    <div class="flex flex-col flex-1 overflow-auto">
      <JotItem
        v-for="jot in jots"
        :key="jot.id"
        :jot="jot"
        @onDelete="handleJotDelete(jot.id)"
      />
    </div>
    <div class="p-4 border-t-2 border-t-base-300">
      <button @click="createNewJot" class="btn btn-neutral w-full">
        New Jot
        <span
          v-if="system === 'darwin'"
          class="ml-1 text-xs font-bold text-neutral-content/60"
          >⌘ + N</span
        >
        <span
          v-if="system === 'windows'"
          class="ml-1 text-xs font-bold text-neutral-content/60"
          >Ctrl + N</span
        >
      </button>
    </div>
  </div>
</template>
