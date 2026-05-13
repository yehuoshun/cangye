<template>
  <div>
    <h2 style="font-size:18px; font-weight:600; margin-bottom:16px">总览</h2>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else style="display:grid; grid-template-columns:repeat(auto-fill, minmax(200px,1fr)); gap:12px">
      <div class="card" style="text-align:center; padding:20px">
        <div style="font-size:36px; margin-bottom:8px">📁</div>
        <div style="font-size:24px; font-weight:700; color:var(--text-accent)">{{ stats.root_collections || 0 }}</div>
        <div style="font-size:12px; color:var(--text-muted); margin-top:4px">根文件夹</div>
      </div>
      <div class="card" style="text-align:center; padding:20px">
        <div style="font-size:36px; margin-bottom:8px">📂</div>
        <div style="font-size:24px; font-weight:700; color:var(--text-accent)">{{ stats.sub_collections || 0 }}</div>
        <div style="font-size:12px; color:var(--text-muted); margin-top:4px">子文件夹</div>
      </div>
      <div class="card" style="text-align:center; padding:20px">
        <div style="font-size:36px; margin-bottom:8px">🔗</div>
        <div style="font-size:24px; font-weight:700; color:var(--text-accent)">{{ stats.paths || 0 }}</div>
        <div style="font-size:12px; color:var(--text-muted); margin-top:4px">扫描路径</div>
      </div>
      <div class="card" style="text-align:center; padding:20px">
        <div style="font-size:36px; margin-bottom:8px">📄</div>
        <div style="font-size:24px; font-weight:700; color:var(--text-accent)">{{ stats.virtual_files || 0 }}</div>
        <div style="font-size:12px; color:var(--text-muted); margin-top:4px">虚拟文件</div>
      </div>
      <div class="card" style="text-align:center; padding:20px">
        <div style="font-size:36px; margin-bottom:8px">📑</div>
        <div style="font-size:24px; font-weight:700; color:var(--text-accent)">{{ stats.scanned_files || 0 }}</div>
        <div style="font-size:12px; color:var(--text-muted); margin-top:4px">扫描文件数</div>
      </div>
      <div class="card" style="text-align:center; padding:20px">
        <div style="font-size:36px; margin-bottom:8px">🏷️</div>
        <div style="font-size:24px; font-weight:700; color:var(--text-accent)">{{ stats.tags || 0 }}</div>
        <div style="font-size:12px; color:var(--text-muted); margin-top:4px">标签</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const loading = ref(false)
const stats = ref<Record<string, number>>({})

onMounted(async () => {
  loading.value = true
  try {
    const res = await api.overview.stats()
    stats.value = res.stats
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})
</script>
