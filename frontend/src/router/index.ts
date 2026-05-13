import { createRouter, createWebHashHistory } from 'vue-router'
import FileRoot from '@/views/file/FileRoot.vue'
import FileBrowse from '@/views/file/FileBrowse.vue'
import OverviewPage from '@/views/overview/OverviewPage.vue'
import RssPage from '@/views/rss/RssPage.vue'
import CheckinPage from '@/views/checkin/CheckinPage.vue'
import SettingsPage from '@/views/settings/SettingsPage.vue'

const routes = [
  { path: '/', name: 'home', component: FileRoot },
  { path: '/browse/:id', name: 'browse', component: FileBrowse },
  { path: '/overview', name: 'overview', component: OverviewPage },
  { path: '/rss', name: 'rss', component: RssPage },
  { path: '/checkin', name: 'checkin', component: CheckinPage },
  { path: '/settings', name: 'settings', component: SettingsPage },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
