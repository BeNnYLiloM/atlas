<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { tasksApi, KANBAN_COLUMNS, TASK_STATUS_LABELS } from '@/api/tasks'
import type { Task, TaskStatus } from '@/api/tasks'
import { useWorkspaceStore } from '@/stores/workspace'
import { useProjectsStore } from '@/stores/projects'
import TaskCard from '@/components/tasks/TaskCard.vue'
import TaskCreationModal from '@/components/tasks/TaskCreationModal.vue'

const route = useRoute()
const workspaceStore = useWorkspaceStore()
const projectsStore = useProjectsStore()

const projectId = computed(() => route.params.projectId as string | undefined || undefined)

const currentProject = computed(() =>
  projectId.value ? projectsStore.projects.find(p => p.id === projectId.value) : null
)

const tasks = ref<Task[]>([])
const loading = ref(false)
const showCreateModal = ref(false)

async function loadTasks() {
  if (!workspaceStore.currentWorkspaceId) return
  loading.value = true
  try {
    tasks.value = await tasksApi.list(workspaceStore.currentWorkspaceId, {
      projectId: projectId.value,
    })
  } finally {
    loading.value = false
  }
}

function getColumnTasks(status: TaskStatus): Task[] {
  return tasks.value.filter(t => t.status === status)
}

watch(projectId, loadTasks)
onMounted(loadTasks)
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-subtle flex-shrink-0">
      <div>
        <h1 class="text-lg font-semibold text-primary">
          Задачи
        </h1>
        <p
          v-if="currentProject"
          class="text-xs text-muted mt-0.5"
        >
          {{ currentProject.name }}
        </p>
      </div>
      <button
        class="btn btn-primary"
        @click="showCreateModal = true"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        Добавить задачу
      </button>
    </div>

    <!-- Loading -->
    <div
      v-if="loading"
      class="flex-1 flex items-center justify-center"
    >
      <div class="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
    </div>

    <!-- Kanban board -->
    <div
      v-else
      class="flex-1 overflow-x-auto p-6"
    >
      <div class="flex gap-4 h-full min-w-max">
        <div
          v-for="status in KANBAN_COLUMNS"
          :key="status"
          class="w-72 flex flex-col"
        >
          <!-- Column header -->
          <div class="flex items-center justify-between mb-2">
            <h2 class="text-sm font-semibold text-secondary uppercase tracking-wide text-xs">
              {{ TASK_STATUS_LABELS[status] }}
            </h2>
            <span class="text-xs text-muted bg-elevated border border-default px-2 py-0.5 rounded-full">{{ getColumnTasks(status).length }}</span>
          </div>

          <!-- Cards -->
          <div class="flex-1 space-y-2 min-h-16 rounded-xl p-2 border border-default bg-surface">
            <TaskCard
              v-for="task in getColumnTasks(status)"
              :key="task.id"
              :task="task"
              @updated="loadTasks"
              @deleted="loadTasks"
            />

            <!-- Empty state -->
            <div
              v-if="getColumnTasks(status).length === 0"
              class="text-center py-8 text-xs text-faint"
            >
              Нет задач
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <TaskCreationModal
      v-if="showCreateModal"
      :project-id="projectId"
      @close="showCreateModal = false"
      @created="loadTasks"
    />
  </div>
</template>

