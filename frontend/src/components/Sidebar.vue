<script setup lang="ts">
import { computed } from 'vue'
import { useJotStore } from '../store/jotStore'
import Logo from '../assets/Logo.vue'
const jotStore = useJotStore()

const jots = computed(() => jotStore.listJots())

</script>

<template>
    <div class="w-80 border-r-2 border-r-base-300 flex flex-col h-full">
        <div class="p-2 border-b-2 border-b-base-300 flex flex-row items-center">
            <Logo class="w-10 h-10 fill-neutral " />
            <div class="text-2xl text-neutral font-extrabold">Jot</div>
        </div>
        <div class="flex flex-col flex-1 overflow-auto">
            <RouterLink v-for="jot in jots" :key="jot.id" :to="`/jot/${jot.id}`" class="p-4 w-full hover:bg-base-300"
                activeClass="bg-base-200" exactActiveClass="bg-base-200">
                <h2 class="text-lg font-bold">{{ jot.title }}</h2>
            </RouterLink>
        </div>
        <div class="p-4 border-t-2 border-t-base-300">
            <button @click="jotStore.createJot()" class="btn btn-neutral w-full">
                New Jot
            </button>
        </div>
    </div>
</template>