<template>
    <div class="group relative flex" :aria-label="label">
        <div class="absolute pointer-events-none bg-base-300 w-max rounded-md py-1 px-2 line-clamp-1 group-hover:opacity-100 border-black/30 border opacity-0 self-center transition-opacity duration-300 delay-0 group-hover:delay-1000 shadow-md max-h-10 flex items-center justify-center" :class="placementClass">
            <slot name="content" /> <div class="text-xs ml-2">{{ label }}</div>
        </div>
        <slot />
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const placementClass = computed(() => {
    switch (true) {
        case props.placement === "top":
            return "bottom-full mb-2";
        case props.placement === "bottom":
            return "top-full mt-2";
        case props.placement === "left":
            return "right-full mr-2";
        case props.placement === "right":
        default:
            return "left-full ml-2";
    }
});

const props = defineProps<{
    label: string;
    placement?: "top" | "bottom" | "left" | "right";
}>();

</script>
