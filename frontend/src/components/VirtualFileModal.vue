<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <h3>{{ isEditing ? '编辑虚拟文件' : '新建虚拟文件' }}</h3>

      <div class="form-group">
        <label>路径</label>
        <input v-model="form.path" placeholder="例如: D:\files\doc.pdf" />
        <div style="font-size:11px; color:var(--text-muted); margin-top:4px">
          支持前缀: 115: D:\path, tg: https://t.me/...
        </div>
      </div>

      <div class="form-group">
        <label>显示名称 (可选)</label>
        <input v-model="form.display_name" placeholder="留空则使用文件名" />
      </div>

      <div class="modal-actions">
        <button class="btn" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="save">{{ isEditing ? '保存' : '创建' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { api } from '@/api'
import type { VirtualFile } from '@/api'

const props = defineProps<{
  collectionId: string
  vfile?: VirtualFile | null
}>()

const emit = defineEmits<{
  close: []
  saved: [file: VirtualFile]
}>()

const isEditing = !!props.vfile

const form = reactive({
  path: props.vfile?.path || '',
  display_name: props.vfile?.display_name || '',
})

async function save() {
  try {
    if (isEditing && props.vfile) {
      const res = await api.vfiles.update(props.vfile.id, {
        path: form.path,
        display_name: form.display_name || null,
      })
      emit('saved', res as any)
    } else {
      const res = await api.vfiles.create(props.collectionId, {
        path: form.path,
        display_name: form.display_name || null,
      })
      emit('saved', res)
    }
  } catch (e: any) {
    alert('操作失败: ' + e.message)
  }
}
</script>
