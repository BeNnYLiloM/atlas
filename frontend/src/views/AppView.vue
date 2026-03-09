<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { useAuthStore, useWorkspaceStore, useChannelsStore, useWebSocketStore } from '@/stores'
import Sidebar from '@/components/layout/Sidebar.vue'
import SearchBar from '@/components/search/SearchBar.vue'
import ShortcutsModal from '@/components/settings/ShortcutsModal.vue'
import { useUIStore } from '@/stores/ui'

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()
const wsStore = useWebSocketStore()
const uiStore = useUIStore()
const router = useRouter()

// Инициализируем тему при загрузке
uiStore.initTheme()

function onGlobalKeydown(e: KeyboardEvent) {
  // Alt+↑/↓ - переключение каналов
  if (e.altKey && (e.key === 'ArrowUp' || e.key === 'ArrowDown')) {
    e.preventDefault()
    const textChans = channelsStore.textChannels
    if (textChans.length === 0) return

    const currentIdx = textChans.findIndex(c => c.id === channelsStore.currentChannelId)
    let nextIdx: number

    if (e.key === 'ArrowUp') {
      nextIdx = currentIdx <= 0 ? textChans.length - 1 : currentIdx - 1
    } else {
      nextIdx = currentIdx >= textChans.length - 1 ? 0 : currentIdx + 1
    }

    const nextChannel = textChans[nextIdx]
    channelsStore.setCurrentChannel(nextChannel.id)
    router.push(`/channels/${nextChannel.id}`)
  }
}

onMounted(async () => {
  await authStore.fetchUser()
  await workspaceStore.fetchWorkspaces()
  wsStore.connect()
  document.addEventListener('keydown', onGlobalKeydown)
})

onUnmounted(() => {
  wsStore.disconnect()
  document.removeEventListener('keydown', onGlobalKeydown)
})

// При смене воркспейса загружаем каналы и подписываемся на workspace
watch(
  () => workspaceStore.currentWorkspaceId,
  async (newWorkspaceId, oldWorkspaceId) => {
    // Отписываемся от старого workspace
    if (oldWorkspaceId) {
      wsStore.unsubscribeFromWorkspace(oldWorkspaceId)
    }
    
    // Подписываемся на новый workspace и загружаем каналы
    if (newWorkspaceId) {
      wsStore.subscribeToWorkspace(newWorkspaceId)
      await Promise.all([
        channelsStore.fetchChannels(newWorkspaceId),
        workspaceStore.fetchMembers(newWorkspaceId),
      ])
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="flex h-screen bg-dark-950">
    <!-- Sidebar -->
    <Sidebar />

    <!-- Main content -->
    <main class="flex-1 flex flex-col min-w-0">
      <!-- Top bar with search -->
      <div class="h-12 flex items-center px-4 border-b border-dark-800 bg-dark-950 flex-shrink-0">
        <div class="flex-1" />
        <SearchBar />
        <div class="flex-1" />
      </div>
      <RouterView />
    </main>

    <!-- Глобальные модалки -->
    <ShortcutsModal />
  </div>
</template>

