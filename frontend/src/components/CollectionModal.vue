<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <h3>{{ isEditing ? '编辑文件夹' : '新建文件夹' }}</h3>

      <div class="form-group">
        <label>名称</label>
        <input v-model="form.name" placeholder="文件夹名称" />
      </div>

      <div class="form-group">
        <label>图标</label>
        <input v-model="form.icon" placeholder="📁" style="width:60px" />
      </div>

      <div class="form-group" v-if="showParent">
        <label>父级文件夹 ID (可选)</label>
        <input v-model="form.parent_id" placeholder="留空则为根级" />
      </div>

      <div class="modal-actions">
        <button class="btn" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="save">{{ isEditing ? '保存' : '创建' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import { api } from '@/api'
import type { Collection } from '@/api'

const props = defineProps<{
  collection?: Collection | null
  showParent?: boolean
}>()

const emit = defineEmits<{
  close: []
  saved: [collection: Collection]
}>()

const isEditing = !!props.collection

const form = reactive({
  name: props.collection?.name || '',
  icon: props.collection?.icon || '📁',
  parent_id: props.collection?.parent_id || null as string | null,
})

watch(() => props.collection, (c) => {
  if (c) {
    form.name = c.name
    form.icon = c.icon
    form.parent_id = c.parent_id
  }
})

async function save() {
  try {
    if (isEditing && props.collection) {
      const res = await api.collections.update(props.collection.id, {
        name: form.name,
        icon: form.icon,
      })
      emit('saved', res)
    } else {
      const res = await api.collections.create({
        name: form.name,
        icon: form.icon,
        parent_id: form.parent_id,
      })
      emit('saved', res)
    }
  } catch (e: any) {
    alert('操作失败: ' + e.message)
  }
}
</script>
