<template>
  <aside class="app-sidebar">
    <div class="sidebar-header">
      🍃 藏叶
    </div>
    <ul class="sidebar-nav">
      <li :class="{ active: isActive('/') }" @click="go('/')">
        <span class="icon">📁</span> 我的文件夹
      </li>
      <li :class="{ active: isActive('/overview') }" @click="go('/overview')">
        <span class="icon">📊</span> 总览
      </li>
      <li :class="{ active: isActive('/rss') }" @click="go('/rss')">
        <span class="icon">📡</span> RSS
      </li>
      <li :class="{ active: isActive('/checkin') }" @click="go('/checkin')">
        <span class="icon">✅</span> 签到
      </li>
      <li :class="{ active: isActive('/settings') }" @click="go('/settings')">
        <span class="icon">⚙️</span> 设置
      </li>
    </ul>
    <div style="flex:1" />
    <div style="padding:12px; font-size:11px; color:var(--text-muted); text-align:center">
      Cangye v0.1.0
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

function isActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function go(path: string) {
  router.push(path)
}
</script>
