<script setup lang="ts">
import { computed } from "vue";
import { useJotStore } from "../../store/jotStore";
import { useRouter } from "vue-router";
import Logo from "../../assets/Logo.vue";
import JotItem from "./JotItem.vue";

const jotStore = useJotStore();
const router = useRouter();
const jots = computed(() => jotStore.listJots());

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
</script>

<template>
  <div class="w-80 border-r-2 border-r-base-300 flex flex-col h-full">
    <div class="p-2 border-b-2 border-b-base-300 flex flex-row items-center">
      <Logo class="w-10 h-10 fill-neutral" />
      <div class="text-2xl text-neutral font-extrabold">Jot</div>
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
      </button>
    </div>
  </div>
</template>
