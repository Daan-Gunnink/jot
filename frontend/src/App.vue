<script setup lang="ts">
import { onMounted } from "vue";
import UpdaterService from "./components/UpdaterService.vue";
import { useJotStore } from "./store/jotStore";

const jotStore = useJotStore();

onMounted(async () => {
  await jotStore.migrateJots();
});
</script>

<template>
  <main class="h-screen w-screen flex">
    <UpdaterService />
    <RouterView />
  </main>
</template>

<style scoped>
header {
  line-height: 1.5;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>
