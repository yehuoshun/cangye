<template>
  <div class="app-layout">
    <SidebarNav />
    <div class="app-main">
      <div class="app-header">
        <Breadcrumb />
        <div style="flex:1" />
        <div class="view-toggle" v-if="showViewToggle">
          <button :class="{ active: viewMode === 'grid' }" @click="viewMode='grid'">▦</button>
          <button :class="{ active: viewMode === 'list' }" @click="viewMode='list'">☰</button>
        </div>
      </div>
      <div class="app-content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, provide, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import SidebarNav from '@/components/SidebarNav.vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { useTheme } from '@/composables/useTheme'

const route = useRoute()
const { load: loadTheme } = useTheme()
const viewMode = ref('grid')

const showViewToggle = ref(false)

provide('viewMode', viewMode)

onMounted(() => {
  loadTheme()
})
</script>
