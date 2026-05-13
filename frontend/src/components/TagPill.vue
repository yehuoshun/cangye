<template>
  <span class="tag-pill" :class="typeClass" :style="customStyle">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  type?: 'local' | 'web' | 'default'
  color?: string
}>()

const typeClass = computed(() => props.type || 'default')

const customStyle = computed(() => {
  if (props.color && props.color !== 'gray') {
    return { background: props.color + '33', color: props.color }
  }
  return undefined
})
</script>
