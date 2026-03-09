import { defineStore } from 'pinia'
import { ref } from 'vue'
import apiClient from '@/api/client'
import type { Message } from '@/types'

export interface SearchResult {
  message: Message
  rank: number
  highlight: string
}

export interface SearchResponse {
  results: SearchResult[]
  total: number
  limit: number
  offset: number
}

export const useSearchStore = defineStore('search', () => {
  const query = ref('')
  const results = ref<SearchResult[]>([])
  const total = ref(0)
  const loading = ref(false)
  const isOpen = ref(false)
  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  function open() {
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
    query.value = ''
    results.value = []
    total.value = 0
  }

  async function search(q: string, workspaceId?: string) {
    query.value = q

    if (!q.trim()) {
      results.value = []
      total.value = 0
      return
    }

    if (debounceTimer) clearTimeout(debounceTimer)

    debounceTimer = setTimeout(async () => {
      loading.value = true
      try {
        const params = new URLSearchParams({ q })
        if (workspaceId) params.append('workspace_id', workspaceId)

        const response = await apiClient.get<{ data: SearchResponse }>(`/search?${params}`)
        results.value = response.data.data.results ?? []
        total.value = response.data.data.total
      } catch (e) {
        console.error('[Search] Error:', e)
        results.value = []
      } finally {
        loading.value = false
      }
    }, 300)
  }

  function $reset() {
    query.value = ''
    results.value = []
    total.value = 0
    loading.value = false
    isOpen.value = false
  }

  return {
    query,
    results,
    total,
    loading,
    isOpen,
    open,
    close,
    search,
    $reset,
  }
})
