<template>
    <div class="w-80">
        <Sidebar />
    </div>
    <div class="grow pl-4 pt-2">
        <TipTap />
    </div>
</template>

<script setup lang="ts">
import Sidebar from '../components/Sidebar.vue'
import TipTap from '../components/TipTap.vue'
import { useRouter, useRoute } from 'vue-router'
import { onMounted, onBeforeUnmount } from 'vue'
import { useJotStore } from '../store/jotStore'

const currentJotId = useRoute().params.id as string
const router = useRouter()
const jotStore = useJotStore()
const handleKeyDown = (event: KeyboardEvent) => {
    if ((event.metaKey || event.ctrlKey) && (event.altKey || event.ctrlKey) && event.key === 'n') {
        event.preventDefault() // Prevent default browser behavior
        event.stopPropagation()
        const id = jotStore.createJot()
        router.push(`/jot/${id}`)
    }
}

onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
})

// Clean up event listeners when component unmounts
onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyDown)
})
</script>