<template>
  <table class="file-list">
    <thead>
      <tr>
        <th class="file-icon-cell"></th>
        <th class="sortable" @click="$emit('sort', 'name')">名称</th>
        <th class="sortable" @click="$emit('sort', 'size')">大小</th>
        <th>类型</th>
        <th v-if="showPrefix">来源</th>
        <th>修改时间</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="item in items"
        :key="item.id"
        @click="$emit('select', item)"
        @contextmenu.prevent="$emit('contextmenu', $event, item)"
      >
        <td class="file-icon-cell">{{ item.icon || '📄' }}</td>
        <td>{{ item.name }}</td>
        <td>{{ formatSize(item.size) }}</td>
        <td>
          <span class="tag-pill" :class="typeClass(item)">{{ item.source }}</span>
          <template v-if="item.prefix">
            <span class="tag-pill" :class="item.prefix_type === 'local' ? 'local' : 'web'" style="margin-left:4px">
              {{ item.prefix }}
            </span>
          </template>
        </td>
        <td v-if="showPrefix">
          <span class="tag-pill" v-if="item.prefix" :class="item.prefix_type === 'local' ? 'local' : 'web'">
            {{ item.prefix }}
          </span>
        </td>
        <td>{{ formatTime(item.mod_time) }}</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup lang="ts">
import type { FileEntry, Collection } from '@/api'

const props = defineProps<{
  items: (FileEntry | Collection)[]
  showPrefix?: boolean
}>()

defineEmits<{
  select: [item: FileEntry | Collection]
  contextmenu: [event: MouseEvent, item: FileEntry | Collection]
  sort: [field: string]
}>()

function formatSize(bytes: number): string {
  if (bytes === 0) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatTime(t: string): string {
  if (!t) return '-'
  try {
    const d = new Date(t)
    return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN')
  } catch {
    return t
  }
}

function typeClass(item: FileEntry | Collection): string {
  if ('source' in item) {
    return item.source === 'scan' ? 'local' : 'default'
  }
  return 'default'
}
</script>
