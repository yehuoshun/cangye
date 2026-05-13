<template>
  <div class="preview-panel" v-if="file">
    <div class="preview-header">
      {{ file.name }}
    </div>

    <!-- Image -->
    <div class="preview-image" v-if="isImage">
      <img :src="thumbnailUrl" alt="" />
    </div>

    <!-- Video -->
    <div v-else-if="isVideo" class="preview-video">
      <video controls autoplay muted style="max-width:100%; border-radius:6px;">
        <source :src="contentUrl" />
      </video>
    </div>

    <!-- Audio -->
    <div v-else-if="isAudio" class="preview-audio">
      <audio controls style="width:100%">
        <source :src="contentUrl" />
      </audio>
    </div>

    <!-- Text -->
    <div v-else-if="isText" class="preview-text">
      <div v-if="loading">加载中...</div>
      <pre v-else>{{ textContent }}</pre>
    </div>

    <!-- Other -->
    <div v-else style="text-align:center; padding:32px; color:var(--text-muted)">
      <div style="font-size:32px; margin-bottom:12px">📎</div>
      <div>暂不支持预览此文件类型</div>
      <button class="btn btn-sm" style="margin-top:12px" @click="openExternal">系统程序打开</button>
    </div>

    <div style="margin-top:12px; padding-top:12px; border-top:1px solid var(--border)">
      <div style="font-size:12px; color:var(--text-secondary)">
        <div>路径: {{ file.path }}</div>
        <div>大小: {{ formatSize(file.size) }}</div>
        <div v-if="file.mod_time">修改: {{ file.mod_time }}</div>
      </div>
    </div>
  </div>
  <div class="preview-panel" v-else style="display:flex; align-items:center; justify-content:center; color:var(--text-muted)">
    点击文件预览
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { api } from '@/api'
import type { FileEntry } from '@/api'

const props = defineProps<{
  file: FileEntry | null
}>()

const textContent = ref('')
const loading = ref(false)

const isImage = computed(() => props.file?.mime_type?.startsWith('image/'))
const isVideo = computed(() => props.file?.mime_type?.startsWith('video/'))
const isAudio = computed(() => props.file?.mime_type?.startsWith('audio/'))
const isText = computed(() => props.file?.mime_type?.startsWith('text/') || props.file?.mime_type === 'application/json')

const thumbnailUrl = computed(() => {
  if (!props.file) return ''
  return api.preview.thumbnail(props.file.path)
})

const contentUrl = computed(() => {
  if (!props.file) return ''
  return `/api/preview/content?path=${encodeURIComponent(props.file.path)}`
})

async function loadText() {
  if (!props.file) return
  loading.value = true
  try {
    textContent.value = await api.preview.content(props.file.path)
  } catch (e: any) {
    textContent.value = '加载失败: ' + e.message
  } finally {
    loading.value = false
  }
}

function openExternal() {
  if (props.file) {
    api.preview.openExternal(props.file.path)
  }
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

watch(() => props.file, (f) => {
  if (f && isText.value) {
    loadText()
  }
})
</script>
