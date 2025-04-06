<script setup lang="ts">
import { computed } from 'vue'
import { useJotStore } from '../store/jotStore'

const jotStore = useJotStore()

const jots = computed(() => jotStore.listJots())

</script>

<template>
    <div class="w-80 border-r-2 border-r-base-300 flex flex-col h-full">
        <div class="p-4 border-b-2 border-b-base-300 flex flex-row">
            <div class="text-4xl text-neutral font-extrabold">JOT</div>
            <div class="text-4xl text-neutral-content">thing</div>
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