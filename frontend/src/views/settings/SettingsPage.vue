<template>
  <div>
    <h2 style="font-size:18px; font-weight:600; margin-bottom:16px">设置</h2>

    <!-- Layout -->
    <div class="settings-section">
      <h3>布局</h3>
      <div class="settings-row">
        <label>导航布局</label>
        <select v-model="layout" @change="saveLayout" style="width:auto">
          <option value="sidebar">侧边栏</option>
          <option value="top">顶部导航</option>
          <option value="split">双栏对开</option>
        </select>
      </div>
    </div>

    <!-- Theme -->
    <div class="settings-section">
      <h3>主题</h3>
      <div class="settings-row">
        <label>界面主题</label>
        <select v-model="theme" @change="saveTheme" style="width:auto">
          <option value="dark">暗色</option>
          <option value="light">亮色</option>
        </select>
      </div>
    </div>

    <!-- Prefixes -->
    <div class="settings-section">
      <h3>路径前缀映射</h3>
      <div v-for="p in prefixes" :key="p.prefix" class="card" style="margin-bottom:8px">
        <div style="display:flex; align-items:center; gap:8px">
          <span class="tag-pill" :class="p.type === 'local' ? 'local' : 'web'" style="font-size:13px">
            {{ p.prefix }}
          </span>
          <span class="tag-pill default" style="font-size:11px">{{ p.type === 'local' ? '本地' : '网页' }}</span>
          <div style="flex:1" />
          <button class="btn btn-sm" @click="editPrefix(p)">编辑</button>
        </div>
        <div v-if="p.map_path" style="font-size:12px; color:var(--text-muted); margin-top:4px">
          映射: {{ p.map_path }}
        </div>
        <div v-if="p.url_template" style="font-size:12px; color:var(--text-muted); margin-top:2px">
          URL模板: {{ p.url_template }}
        </div>
      </div>

      <!-- Edit prefix modal -->
      <div class="modal-overlay" v-if="editingPrefix" @click.self="editingPrefix=null">
        <div class="modal">
          <h3>编辑前缀: {{ editingPrefix.prefix }}</h3>
          <div class="form-group">
            <label>类型</label>
            <select v-model="editForm.type" style="width:auto">
              <option value="local">本地</option>
              <option value="web">网页</option>
            </select>
          </div>
          <div class="form-group">
            <label>本地映射路径</label>
            <input v-model="editForm.map_path" placeholder="例如: D:\115\download" />
          </div>
          <div class="form-group">
            <label>URL 模板</label>
            <input v-model="editForm.url_template" placeholder="例如: https://115.com/?pickcode={path}" />
          </div>
          <div class="modal-actions">
            <button class="btn" @click="editingPrefix=null">取消</button>
            <button class="btn btn-primary" @click="savePrefix">保存</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Tags -->
    <div class="settings-section">
      <h3>标签管理</h3>
      <div style="display:flex; flex-wrap:wrap; gap:6px">
        <span v-for="tag in tags" :key="tag.id" class="tag-pill default">
          {{ tag.name }}
          <span class="remove" style="cursor:pointer; margin-left:2px; opacity:0.6" @click="deleteTag(tag)">✕</span>
        </span>
      </div>
      <div style="display:flex; gap:6px; margin-top:8px">
        <input v-model="newTagName" placeholder="新标签名称" style="width:200px" @keydown.enter="createTag" />
        <button class="btn btn-sm btn-primary" @click="createTag">添加</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import type { Prefix, Tag } from '@/api'
import { useTheme } from '@/composables/useTheme'

const { theme, set: setTheme } = useTheme()
const layout = ref('sidebar')
const prefixes = ref<Prefix[]>([])
const tags = ref<Tag[]>([])
const newTagName = ref('')
const editingPrefix = ref<Prefix | null>(null)
const editForm = ref({ type: 'local', map_path: '', url_template: '' })

onMounted(load)

async function load() {
  try {
    const [layoutRes, prefixesRes, tagsRes] = await Promise.all([
      api.settings.get('layout'),
      api.prefixes.list(),
      api.tags.search(''),
    ])
    if (layoutRes.value) layout.value = layoutRes.value
    prefixes.value = prefixesRes
    tags.value = tagsRes
  } catch (e) {
    console.error(e)
  }
}

async function saveLayout() {
  await api.settings.set('layout', layout.value)
}

async function saveTheme() {
  await setTheme(theme.value)
}

async function createTag() {
  if (!newTagName.value.trim()) return
  try {
    const tag = await api.tags.create({ name: newTagName.value.trim() })
    tags.value.push(tag)
    newTagName.value = ''
  } catch (e: any) {
    alert('创建失败: ' + e.message)
  }
}

async function deleteTag(tag: Tag) {
  if (!confirm(`确定删除标签"${tag.name}"？`)) return
  try {
    await api.tags.delete(tag.id)
    tags.value = tags.value.filter(t => t.id !== tag.id)
  } catch (e: any) {
    alert('删除失败: ' + e.message)
  }
}

function editPrefix(p: Prefix) {
  editingPrefix.value = p
  editForm.value = {
    type: p.type,
    map_path: p.map_path,
    url_template: p.url_template,
  }
}

async function savePrefix() {
  if (!editingPrefix.value) return
  try {
    await api.prefixes.update(editingPrefix.value.prefix, editForm.value)
    editingPrefix.value = null
    load()
  } catch (e: any) {
    alert('保存失败: ' + e.message)
  }
}
</script>
