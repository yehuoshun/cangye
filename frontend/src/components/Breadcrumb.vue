<template>
  <div class="breadcrumb">
    <router-link to="/">首页</router-link>
    <template v-for="(crumb, i) in crumbs" :key="i">
      <span class="sep">/</span>
      <router-link v-if="crumb.path" :to="crumb.path">{{ crumb.label }}</router-link>
      <span v-else>{{ crumb.label }}</span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const crumbs = computed(() => {
  const list: { label: string; path?: string }[] = []

  if (route.path.startsWith('/browse/')) {
    list.push({ label: '浏览文件夹' })
  } else if (route.path.startsWith('/overview')) {
    list.push({ label: '总览' })
  } else if (route.path.startsWith('/rss')) {
    list.push({ label: 'RSS' })
  } else if (route.path.startsWith('/checkin')) {
    list.push({ label: '签到' })
  } else if (route.path.startsWith('/settings')) {
    list.push({ label: '设置' })
  }

  return list
})
</script>
