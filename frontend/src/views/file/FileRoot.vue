<template>
  <div>
    <div style="display:flex; align-items:center; gap:12px; margin-bottom:16px">
      <h2 style="font-size:18px; font-weight:600; color:var(--text-primary)">我的文件夹</h2>
      <button class="btn btn-primary btn-sm" @click="showCreate = true">+ 新建</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <template v-if="!loading">
      <!-- Grid view -->
      <div v-if="viewMode === 'grid'" class="file-grid">
        <div
          v-for="col in collections"
          :key="col.id"
          class="file-grid-item"
          @click="openCollection(col)"
          @contextmenu.prevent="showContextMenu($event, col)"
        >
          <span class="file-icon">{{ col.icon }}</span>
          <div class="file-name" :title="col.name">{{ col.name }}</div>
          <div class="file-info">
            <span v-if="col.path_count">{{ col.path_count }} 路径</span>
            <span v-if="col.file_count"> · {{ col.file_count }} 文件</span>
          </div>
        </div>
      </div>

      <!-- List view -->
      <table v-else class="file-list">
        <thead>
          <tr>
            <th></th>
            <th>名称</th>
            <th>路径数</th>
            <th>文件数</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="col in collections"
            :key="col.id"
            @click="openCollection(col)"
            @contextmenu.prevent="showContextMenu($event, col)"
          >
            <td style="width:28px; text-align:center">{{ col.icon }}</td>
            <td>{{ col.name }}</td>
            <td>{{ col.path_count || 0 }}</td>
            <td>{{ col.file_count || 0 }}</td>
            <td @click.stop>
              <button class="btn-icon" @click.stop="editCollection(col)" title="编辑">✏️</button>
              <button class="btn-icon" @click.stop="deleteCollection(col)" title="删除">🗑️</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="collections.length === 0" class="empty-state">
        <div class="icon">📁</div>
        <p>还没有虚拟文件夹</p>
        <button class="btn btn-primary" style="margin-top:12px" @click="showCreate = true">创建第一个文件夹</button>
      </div>
    </template>

    <!-- Context Menu -->
    <ContextMenu
      :visible="contextMenu.visible"
      :x="contextMenu.x"
      :y="contextMenu.y"
      @action="onContextAction"
    />

    <!-- Create/Edit Modal -->
    <CollectionModal
      v-if="showCreate || editingCollection"
      :collection="editingCollection"
      @close="closeModal"
      @saved="onSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api'
import type { Collection } from '@/api'
import ContextMenu from '@/components/ContextMenu.vue'
import CollectionModal from '@/components/CollectionModal.vue'

const router = useRouter()
const viewMode = inject('viewMode') as Ref<string>

const collections = ref<Collection[]>([])
const loading = ref(false)
const showCreate = ref(false)
const editingCollection = ref<Collection | null>(null)

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  target: null as Collection | null,
})

onMounted(load)

document.addEventListener('click', () => {
  contextMenu.value.visible = false
})

async function load() {
  loading.value = true
  try {
    collections.value = await api.collections.list()
  } catch (e: any) {
    console.error('load failed', e)
  } finally {
    loading.value = false
  }
}

function openCollection(col: Collection) {
  router.push(`/browse/${col.id}`)
}

function editCollection(col: Collection) {
  editingCollection.value = col
}

async function deleteCollection(col: Collection) {
  if (!confirm(`确定删除"${col.name}"？\n所有关联的路径和文件将被删除。`)) return
  try {
    await api.collections.delete(col.id)
    load()
  } catch (e: any) {
    alert('删除失败: ' + e.message)
  }
}

function showContextMenu(e: MouseEvent, col: Collection) {
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
    target: col,
  }
}

function onContextAction(action: string) {
  const col = contextMenu.value.target
  if (!col) return
  contextMenu.value.visible = false

  switch (action) {
    case 'preview':
      openCollection(col)
      break
    case 'open':
      // TODO
      break
    case 'copy-path':
      navigator.clipboard.writeText(col.id)
      break
    case 'properties':
      editCollection(col)
      break
  }
}

function closeModal() {
  showCreate.value = false
  editingCollection.value = null
}

function onSaved() {
  closeModal()
  load()
}
</script>
