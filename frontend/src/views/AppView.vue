<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import { useAuthStore, useWorkspaceStore, useChannelsStore, useWebSocketStore } from '@/stores'
import Sidebar from '@/components/layout/Sidebar.vue'
import SearchBar from '@/components/search/SearchBar.vue'
import ShortcutsModal from '@/components/settings/ShortcutsModal.vue'
import { useIdleDetector } from '@/composables'
import { authApi } from '@/api/auth'
import type { UserStatusValue } from '@/api/auth'

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()
const wsStore = useWebSocketStore()
const router = useRouter()

// Статусы при которых AFK не вмешивается — пользователь сам выбрал
const MANUAL_STATUSES: UserStatusValue[] = ['dnd', 'offline']

// Статус до перехода в AFK — чтобы восстановить при возврате
let statusBeforeIdle: UserStatusValue | null = null

async function applyStatus(status: UserStatusValue) {
  try {
    const updated = await authApi.updateStatus(status, authStore.user?.custom_status ?? null)
    authStore.user = updated
  } catch { /* тихо */ }
}

useIdleDetector({
  onIdle() {
    const current = authStore.user?.status as UserStatusValue | undefined
    // Не трогаем если пользователь вручную поставил dnd/offline
    if (!current || MANUAL_STATUSES.includes(current)) return
    statusBeforeIdle = current
    applyStatus('away')
  },
  onActive() {
    if (!statusBeforeIdle) return
    // Восстанавливаем только если текущий статус всё ещё "away" (AFK)
    // — пользователь мог вручную сменить его пока был idle
    const current = authStore.user?.status as UserStatusValue | undefined
    if (current === 'away') {
      applyStatus(statusBeforeIdle)
    }
    statusBeforeIdle = null
  },
})

function onGlobalKeydown(e: KeyboardEvent) {
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
  if (!authStore.user && authStore.token) {
    await authStore.fetchUser()
  }
  if (!authStore.isAuthenticated) {
    return
  }

  await workspaceStore.fetchWorkspaces()
  wsStore.connect()
  document.addEventListener('keydown', onGlobalKeydown)
})

onUnmounted(() => {
  wsStore.disconnect()
  document.removeEventListener('keydown', onGlobalKeydown)
})

watch(
  () => workspaceStore.currentWorkspaceId,
  async (newWorkspaceId, oldWorkspaceId) => {
    if (oldWorkspaceId) {
      wsStore.unsubscribeFromWorkspace(oldWorkspaceId)
    }

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
  <div class="flex h-screen bg-base">
    <Sidebar />

    <main class="flex-1 flex flex-col min-w-0">
      <div class="h-12 flex items-center px-4 border-b border-subtle bg-base flex-shrink-0">
        <div class="flex-1" />
        <SearchBar />
        <div class="flex-1" />
      </div>
      <RouterView />
    </main>

    <ShortcutsModal />
  </div>
</template>

