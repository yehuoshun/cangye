import { ref, watch } from 'vue'
import { api } from '@/api'

const theme = ref('dark')

export function useTheme() {
  async function load() {
    try {
      const res = await api.settings.get('theme')
      if (res.value) {
        theme.value = res.value
      }
    } catch {
      // default dark
    }
    applyTheme()
  }

  function applyTheme() {
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  async function toggle() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
    applyTheme()
    await api.settings.set('theme', theme.value)
  }

  async function set(t: string) {
    theme.value = t
    applyTheme()
    await api.settings.set('theme', t)
  }

  return { theme, load, toggle, set }
}
