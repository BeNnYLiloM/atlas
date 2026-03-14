<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectsStore, useWorkspaceStore, useAuthStore } from '@/stores'

const router = useRouter()
const projectsStore = useProjectsStore()
const workspaceStore = useWorkspaceStore()
const authStore = useAuthStore()

const currentMember = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return null
  const members = workspaceStore.membersMap[wsId] ?? []
  return members.find(m => m.user_id === authStore.user?.id) ?? null
})

const canCreateProject = computed(() => {
  if (currentMember.value?.role === 'owner') return true
  return false
})

function openProject(projectId: string) {
  projectsStore.currentProjectId = projectId
  router.push({ name: 'project', params: { projectId } })
}
</script>

<template>
  <div v-if="projectsStore.projects.length > 0 || canCreateProject">
    <div class="px-3 pt-4 pb-1 flex items-center justify-between">
      <p class="text-xs font-semibold text-muted uppercase tracking-wider">
        Проекты
      </p>
    </div>

    <button
      v-for="project in projectsStore.projects"
      :key="project.id"
      class="w-full text-left px-2 py-1.5 text-sm rounded-lg transition-all flex items-center gap-2 group"
      :class="projectsStore.currentProjectId === project.id
        ? 'bg-accent/20 text-primary'
        : 'text-secondary hover:bg-elevated hover:text-primary'"
      @click="openProject(project.id)"
    >
      <span
        class="w-10 h-10 rounded-md flex-shrink-0 overflow-hidden flex items-center justify-center text-lg font-semibold transition-transform group-hover:scale-105"
        :style="!project.icon_url ? { backgroundColor: 'var(--accent-600-hex)', color: 'var(--accent-300-hex)' } : {}"
      >
        <img
          v-if="project.icon_url"
          :src="project.icon_url"
          :alt="project.name"
          class="w-full h-full object-cover"
        >
        <template v-else>
          {{ project.name[0]?.toUpperCase() }}
        </template>
      </span>
      <span class="truncate flex-1">
        {{ project.name }}
      </span>
      <span
        v-if="project.is_archived"
        class="text-xs text-muted"
      >
        (архив)
      </span>
    </button>

    <div
      v-if="projectsStore.projects.length === 0"
      class="px-3 py-1 text-xs text-muted"
    >
      Нет проектов
    </div>
  </div>
</template>
