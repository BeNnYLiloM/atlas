<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWorkspaceStore } from '@/stores'
import { useProjectsStore } from '@/stores/projects'
import { useNavigationStore } from '@/stores/navigation'
import { useDMStore } from '@/stores/dm'
import { Avatar, Modal, Input, Button } from '@/components/ui'

const router = useRouter()
const route = useRoute()
const workspaceStore = useWorkspaceStore()
const projectsStore = useProjectsStore()
const navigationStore = useNavigationStore()
const dmStore = useDMStore()

const showWorkspaceDropdown = ref(false)
const showCreateModal = ref(false)
const newWorkspaceName = ref('')
const creating = ref(false)

const currentProjectId = computed(() => navigationStore.activeProjectId)

const projects = computed(() => projectsStore.projects)

function selectWorkspace(id: string) {
  workspaceStore.setCurrentWorkspace(id)
  showWorkspaceDropdown.value = false
  navigationStore.setSection('channels')
  router.push('/channels')
}

function goToChannels() {
  navigationStore.setSection('channels')
  router.push('/channels')
}

function goToDM() {
  navigationStore.setSection('dm')
  router.push('/dm')
}

function goToProject(projectId: string) {
  navigationStore.setProject(projectId)
  router.push({ name: 'project', params: { projectId } })
}

async function createWorkspace() {
  if (!newWorkspaceName.value.trim()) return
  creating.value = true
  try {
    await workspaceStore.createWorkspace({ name: newWorkspaceName.value.trim() })
    showCreateModal.value = false
    newWorkspaceName.value = ''
  } finally {
    creating.value = false
  }
}

function closeDropdown(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('[data-workspace-switcher]')) {
    showWorkspaceDropdown.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', closeDropdown))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeDropdown))

// Синхронизируем activeSection с текущим маршрутом при навигации браузерными кнопками
const isDM = computed(() => route.name === 'dm' || route.name === 'dm-channel')
const isProject = computed(() =>
  route.name === 'project' || route.name === 'project-channel' || route.name === 'project-tasks',
)
</script>

<template>
  <nav class="w-14 flex flex-col items-center py-2 gap-1 bg-[var(--bg-elevated)] border-r border-subtle shrink-0">
    <!-- Workspace switcher -->
    <div class="relative mb-1" data-workspace-switcher>
      <button
        class="w-10 h-10 rounded-xl overflow-hidden bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] flex items-center justify-center text-white font-bold shadow-md hover:rounded-2xl transition-all duration-150"
        :title="workspaceStore.currentWorkspace?.name"
        @click="showWorkspaceDropdown = !showWorkspaceDropdown"
      >
        <img
          v-if="workspaceStore.currentWorkspace?.icon_url"
          :src="workspaceStore.currentWorkspace.icon_url"
          class="w-full h-full object-cover"
          alt=""
        >
        <span v-else>{{ workspaceStore.currentWorkspace?.name?.[0]?.toUpperCase() || 'A' }}</span>
      </button>

      <!-- Workspace dropdown -->
      <Transition
        enter-active-class="transition-all duration-150 origin-top-left"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition-all duration-100 origin-top-left"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-if="showWorkspaceDropdown"
          class="absolute top-0 left-full ml-2 w-56 rounded-xl bg-overlay border border-default shadow-xl z-50 overflow-hidden"
        >
          <div class="px-3 py-2 border-b border-default">
            <p class="text-xs font-semibold text-subtle uppercase tracking-wider">Воркспейсы</p>
          </div>
          <div class="max-h-64 overflow-y-auto py-1">
            <button
              v-for="ws in workspaceStore.workspaces"
              :key="ws.id"
              class="flex items-center gap-3 w-full px-3 py-2 text-sm text-secondary hover:bg-elevated transition-colors"
              :class="ws.id === workspaceStore.currentWorkspaceId ? 'text-primary font-medium' : ''"
              @click="selectWorkspace(ws.id)"
            >
              <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-[var(--accent-dim)] to-[var(--accent-dim)] flex items-center justify-center text-white text-xs font-bold shrink-0 overflow-hidden">
                <img v-if="ws.icon_url" :src="ws.icon_url" class="w-full h-full object-cover" alt="">
                <span v-else>{{ ws.name[0].toUpperCase() }}</span>
              </div>
              <span class="truncate">{{ ws.name }}</span>
              <svg v-if="ws.id === workspaceStore.currentWorkspaceId" class="w-3.5 h-3.5 text-accent ml-auto shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
          <div class="border-t border-default p-2">
            <button
              class="flex items-center gap-2 w-full px-2 py-1.5 text-xs text-muted hover:text-primary hover:bg-elevated rounded-lg transition-colors"
              @click="showCreateModal = true; showWorkspaceDropdown = false"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              Создать воркспейс
            </button>
          </div>
        </div>
      </Transition>
    </div>

    <div class="w-8 h-px bg-subtle my-1" />

    <!-- DM -->
    <button
      class="rail-btn"
      :class="isDM ? 'rail-btn--active' : ''"
      title="Личные сообщения"
      @click="goToDM"
    >
      <!-- Иконка чата -->
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>
      <!-- Unread badge -->
      <span
        v-if="dmStore.totalUnread > 0 && !isDM"
        class="absolute -top-1 -right-1 min-w-[16px] h-4 px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center"
      >
        {{ dmStore.totalUnread > 99 ? '99+' : dmStore.totalUnread }}
      </span>
    </button>

    <!-- Каналы -->
    <button
      class="rail-btn"
      :class="!isDM && !isProject ? 'rail-btn--active' : ''"
      title="Каналы воркспейса"
      @click="goToChannels"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
      </svg>
    </button>

    <!-- Разделитель + проекты -->
    <template v-if="projects.length > 0">
      <div class="w-8 h-px bg-subtle my-1" />

      <button
        v-for="project in projects"
        :key="project.id"
        class="rail-btn relative"
        :class="isProject && currentProjectId === project.id ? 'rail-btn--active' : ''"
        :title="project.name"
        @click="goToProject(project.id)"
      >
        <Avatar :name="project.name" :src="project.icon_url ?? undefined" size="sm" class="!w-8 !h-8 !rounded-xl hover:!rounded-2xl transition-all duration-150" />
      </button>
    </template>
  </nav>

  <!-- Create workspace modal -->
  <Modal :open="showCreateModal" title="Создать воркспейс" @close="showCreateModal = false">
    <form class="space-y-4" @submit.prevent="createWorkspace">
      <Input v-model="newWorkspaceName" label="Название" placeholder="Моя компания" />
      <div class="flex gap-3 pt-2">
        <Button variant="secondary" class="flex-1" @click="showCreateModal = false">Отмена</Button>
        <Button type="submit" :loading="creating" class="flex-1">Создать</Button>
      </div>
    </form>
  </Modal>
</template>

<style scoped>
.rail-btn {
  @apply relative w-10 h-10 flex items-center justify-center rounded-xl text-muted hover:text-primary hover:bg-surface transition-all duration-150 hover:rounded-2xl;
}

.rail-btn--active {
  @apply text-primary bg-surface rounded-2xl;
}
</style>
