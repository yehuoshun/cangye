<template>
  <div class="file-grid">
    <div
      v-for="item in items"
      :key="item.id"
      class="file-grid-item"
      @click="$emit('select', item)"
      @contextmenu.prevent="$emit('contextmenu', $event, item)"
    >
      <span class="file-icon">{{ item.icon || '📄' }}</span>
      <div class="file-name" :title="item.name">{{ item.name }}</div>
      <div class="file-info">{{ formatSize(item.size) }}</div>
      <div class="file-info" v-if="item.prefix">
        <span class="tag-pill" :class="item.prefix_type === 'local' ? 'local' : 'web'">
          {{ item.prefix }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FileEntry, Collection } from '@/api'

defineProps<{
  items: (FileEntry | Collection)[]
}>()

defineEmits<{
  select: [item: FileEntry | Collection]
  contextmenu: [event: MouseEvent, item: FileEntry | Collection]
}>()

function formatSize(bytes: number): string {
  if (bytes === 0) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>
