<script setup lang="ts">
import { onMounted, onUnmounted, watch, computed } from 'vue'
import { RouterView, useRouter, useRoute } from 'vue-router'
import { useAuthStore, useWorkspaceStore, useChannelsStore, useWebSocketStore } from '@/stores'
import { useProjectsStore } from '@/stores/projects'
import { useDMStore } from '@/stores/dm'
import Sidebar from '@/components/layout/Sidebar.vue'
import AppRail from '@/components/layout/AppRail.vue'
import SearchBar from '@/components/search/SearchBar.vue'
import ShortcutsModal from '@/components/settings/ShortcutsModal.vue'
import { useIdleDetector } from '@/composables'
import { authApi } from '@/api/auth'
import type { UserStatusValue } from '@/api/auth'

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()
const wsStore = useWebSocketStore()
const projectsStore = useProjectsStore()
const dmStore = useDMStore()
const router = useRouter()
const route = useRoute()

const isProjectRoute = computed(() =>
  route.name === 'project' || route.name === 'project-channel' || route.name === 'project-tasks'
)
const activeProject = computed(() => isProjectRoute.value ? projectsStore.currentProject : null)

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
      const fetchTasks: Promise<unknown>[] = [
        workspaceStore.fetchMembers(newWorkspaceId),
        dmStore.fetchDMs(),
        projectsStore.fetchProjects(newWorkspaceId),
      ]
      if (!isProjectRoute.value) {
        fetchTasks.push(channelsStore.fetchChannels(newWorkspaceId))
      }
      await Promise.all(fetchTasks)
    }
  },
  { immediate: true }
)

// При выходе из проекта обратно в воркспейс — перезагрузить каналы воркспейса
watch(isProjectRoute, async (inProject, wasInProject) => {
  if (wasInProject && !inProject) {
    const wsId = workspaceStore.currentWorkspaceId
    if (wsId) {
      await channelsStore.fetchChannels(wsId)
    }
  }
})
</script>

<template>
  <div class="flex h-screen bg-base">
    <AppRail />
    <Sidebar />

    <main class="flex-1 flex flex-col min-w-0">
      <div class="h-12 flex items-center px-4 border-b border-subtle bg-base flex-shrink-0">
        <template v-if="activeProject">
          <span class="font-semibold text-primary truncate">
            {{ activeProject.name }}
            <span v-if="activeProject.is_archived" class="text-xs text-muted ml-2">(архив)</span>
          </span>
          <div class="flex-1" />
        </template>
        <template v-else>
          <div class="flex-1" />
          <SearchBar />
          <div class="flex-1" />
        </template>
      </div>
      <RouterView />
    </main>

    <ShortcutsModal />
  </div>
</template>

