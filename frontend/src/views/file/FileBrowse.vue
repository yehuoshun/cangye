<template>
  <div>
    <div style="display:flex; align-items:center; gap:12px; margin-bottom:16px; flex-wrap:wrap">
      <h2 style="font-size:18px; font-weight:600; color:var(--text-primary)">{{ collection?.name || '浏览' }}</h2>
      <button class="btn btn-sm" @click="showPathModal = true">📂 添加路径</button>
      <button class="btn btn-sm" @click="showFileModal = true">📄 添加文件</button>
      <button class="btn btn-sm" @click="scanAll">🔄 扫描</button>
      <button class="btn btn-sm" @click="goBack">← 返回</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <template v-if="!loading">
      <!-- Paths -->
      <div v-if="paths.length > 0" style="margin-bottom:12px">
        <div v-for="p in paths" :key="p.id" class="card" style="display:flex; align-items:center; gap:8px; margin-bottom:6px">
          <span style="color:var(--text-accent); font-size:12px">{{ p.path }}</span>
          <span class="tag-pill local" v-if="p.auto_scan">自动扫描</span>
          <div style="flex:1" />
          <button class="btn-icon" @click="scanPath(p.id)" title="扫描">🔄</button>
          <button class="btn-icon" @click="deletePath(p.id)" title="删除">🗑️</button>
        </div>
      </div>

      <!-- Files -->
      <div v-if="viewMode === 'grid'">
        <GridView
          :items="files"
          @select="selectFile"
          @contextmenu="showContextMenu"
        />
      </div>
      <table v-else class="file-list">
        <thead>
          <tr>
            <th class="file-icon-cell"></th>
            <th>名称</th>
            <th>大小</th>
            <th>来源</th>
            <th>修改时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="f in files" :key="f.id"
            @click="selectFile(f)"
            @contextmenu.prevent="showContextMenu($event, f)"
          >
            <td class="file-icon-cell">{{ f.icon || '📄' }}</td>
            <td>{{ f.name }}</td>
            <td>{{ formatSize(f.size) }}</td>
            <td>
              <span class="tag-pill" :class="f.source === 'scan' ? 'local' : 'default'">{{ f.source }}</span>
              <span v-if="f.prefix" class="tag-pill web" style="margin-left:4px">{{ f.prefix }}</span>
            </td>
            <td>{{ f.mod_time ? new Date(f.mod_time).toLocaleString('zh-CN') : '-' }}</td>
            <td @click.stop>
              <button class="btn-icon" @click.stop="deleteFile(f)" title="删除">🗑️</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="files.length === 0" class="empty-state">
        <div class="icon">📂</div>
        <p>此文件夹暂无文件</p>
        <button class="btn btn-primary" style="margin-top:12px" @click="showFileModal = true">添加虚拟文件</button>
        <button class="btn" style="margin-top:8px" @click="scanAll">扫描路径</button>
      </div>
    </template>

    <!-- Context Menu -->
    <ContextMenu
      :visible="contextMenu.visible"
      :x="contextMenu.x"
      :y="contextMenu.y"
      @action="onContextAction"
    />

    <!-- Preview Panel (loaded inline if selected) -->
    <PreviewPanel
      v-if="selectedEntry"
      :file="selectedEntry"
      style="position:fixed; right:0; top:48px; bottom:0; z-index:10; box-shadow:-4px 0 16px rgba(0,0,0,0.3)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api'
import type { Collection, CollectionPath, FileEntry } from '@/api'
import GridView from '@/components/GridView.vue'
import ContextMenu from '@/components/ContextMenu.vue'
import PreviewPanel from '@/components/PreviewPanel.vue'
import VirtualFileModal from '@/components/VirtualFileModal.vue'

const route = useRoute()
const router = useRouter()
const viewMode = inject('viewMode') as Ref<string>

const collectionId = route.params.id as string

const collection = ref<Collection | null>(null)
const paths = ref<CollectionPath[]>([])
const files = ref<FileEntry[]>([])
const loading = ref(false)
const selectedEntry = ref<FileEntry | null>(null)

const showPathModal = ref(false)
const showFileModal = ref(false)

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  target: null as any,
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const [col, pathsData, browseData] = await Promise.all([
      api.collections.get(collectionId),
      api.paths.list(collectionId),
      api.browse.list(collectionId),
    ])
    collection.value = col
    paths.value = pathsData
    files.value = browseData
  } catch (e: any) {
    console.error('load failed', e)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/')
}

function selectFile(file: FileEntry) {
  selectedEntry.value = selectedEntry.value?.id === file.id ? null : file
}

function showContextMenu(e: MouseEvent, item: any) {
  contextMenu.value = { visible: true, x: e.clientX, y: e.clientY, target: item }
}

function onContextAction(action: string) {
  const item = contextMenu.value.target
  if (!item) return
  contextMenu.value.visible = false

  switch (action) {
    case 'preview':
      selectedEntry.value = item
      break
    case 'open':
      api.preview.openExternal(item.path)
      break
    case 'copy-path':
      navigator.clipboard.writeText(item.path)
      break
  }
}

async function scanAll() {
  for (const p of paths.value) {
    try {
      await api.paths.scan(p.id)
    } catch (e) {
      console.error('scan failed', p.path, e)
    }
  }
  load()
}

async function scanPath(pathId: string) {
  try {
    await api.paths.scan(pathId)
    load()
  } catch (e: any) {
    alert('扫描失败: ' + e.message)
  }
}

async function deletePath(pathId: string) {
  try {
    await api.paths.delete(pathId)
    load()
  } catch (e: any) {
    alert('删除失败: ' + e.message)
  }
}

async function deleteFile(f: FileEntry) {
  if (!confirm(`确定移除"${f.name}"？`)) return
  try {
    if (f.source === 'scan') return // can't remove scanned
    await api.vfiles.delete(f.id)
    load()
  } catch (e: any) {
    alert('删除失败: ' + e.message)
  }
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>
