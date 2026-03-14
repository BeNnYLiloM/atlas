<script setup lang="ts">
import { watch } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useChannelsStore, useProjectsStore, useWorkspaceStore } from '@/stores'
import { projectsApi } from '@/api'

const route = useRoute()
const router = useRouter()
const channelsStore = useChannelsStore()
const projectsStore = useProjectsStore()
const workspaceStore = useWorkspaceStore()

// Следим за парой [projectId, workspaceId] — при перезагрузке страницы
// workspaceId появляется чуть позже, этот watch подхватит оба изменения
watch(
  () => [route.params.projectId as string, workspaceStore.currentWorkspaceId] as const,
  async ([projectId, workspaceId]) => {
    if (!projectId || !workspaceId) return

    projectsStore.currentProjectId = projectId

    // Если список проектов ещё не загружен — загружаем
    if (projectsStore.projects.length === 0) {
      await projectsStore.fetchProjects(workspaceId)
    }

    // Если проект всё ещё не найден (например нет доступа через обычный список) —
    // запрашиваем напрямую чтобы получить workspaceId
    const inStore = projectsStore.projects.find((p) => p.id === projectId)
    if (!inStore) {
      try {
        await projectsApi.get(projectId)
      } catch {
        return
      }
    }

    await Promise.all([
      channelsStore.fetchProjectChannels(workspaceId, projectId),
      projectsStore.fetchMembers(projectId),
    ])

    // Редирект в первый текстовый канал если channelId не задан и мы не на задачах
    if (!route.params.channelId && route.name !== 'project-tasks' && channelsStore.textChannels.length > 0) {
      const first = channelsStore.textChannels[0]
      channelsStore.setCurrentChannel(first.id)
      router.replace({ name: 'project-channel', params: { projectId, channelId: first.id } })
    }
  },
  { immediate: true }
)
</script>

<template>
  <RouterView />
</template>
