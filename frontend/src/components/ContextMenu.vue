<template>
  <div
    v-if="visible"
    class="context-menu"
    :style="{ left: x + 'px', top: y + 'px' }"
    @click.stop
  >
    <div class="context-menu-item" @click="emit('action', 'preview')">
      👁️ 预览
    </div>
    <div class="context-menu-item" @click="emit('action', 'open')">
      📂 系统程序打开
    </div>
    <div class="context-menu-divider"></div>
    <div class="context-menu-item" @click="emit('action', 'copy-path')">
      📋 复制路径
    </div>
    <div class="context-menu-item" @click="emit('action', 'properties')">
      ℹ️ 属性
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
  x: number
  y: number
}>()

const emit = defineEmits<{
  action: [action: string]
}>()
</script>
