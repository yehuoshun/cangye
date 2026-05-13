import { ref } from 'vue'
import { api } from '@/api'

export function useApi() {
  return { api }
}

export function useAsync<T>(fetcher: () => Promise<T>) {
  const data = ref<T | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      data.value = await fetcher()
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { data, loading, error, load }
}
