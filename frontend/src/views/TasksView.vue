<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { tasksApi, KANBAN_COLUMNS, TASK_STATUS_LABELS } from '@/api/tasks'
import type { Task, TaskStatus } from '@/api/tasks'
import { useWorkspaceStore } from '@/stores/workspace'
import TaskCard from '@/components/tasks/TaskCard.vue'
import TaskCreationModal from '@/components/tasks/TaskCreationModal.vue'

const workspaceStore = useWorkspaceStore()
const tasks = ref<Task[]>([])
const loading = ref(false)
const showCreateModal = ref(false)

async function loadTasks() {
  if (!workspaceStore.currentWorkspaceId) return
  loading.value = true
  try {
    tasks.value = await tasksApi.list(workspaceStore.currentWorkspaceId)
  } finally {
    loading.value = false
  }
}

function getColumnTasks(status: TaskStatus): Task[] {
  return tasks.value.filter(t => t.status === status)
}

onMounted(loadTasks)
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-dark-800 flex-shrink-0">
      <h1 class="text-lg font-semibold text-dark-100">Задачи</h1>
      <button
        class="flex items-center gap-2 px-3 py-1.5 bg-atlas-600 hover:bg-atlas-500 text-white rounded-lg text-sm font-medium transition-colors"
        @click="showCreateModal = true"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Добавить задачу
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="w-6 h-6 border-2 border-atlas-500 border-t-transparent rounded-full animate-spin" />
    </div>

    <!-- Kanban board -->
    <div v-else class="flex-1 overflow-x-auto p-6">
      <div class="flex gap-4 h-full min-w-max">
        <div
          v-for="status in KANBAN_COLUMNS"
          :key="status"
          class="w-72 flex flex-col"
        >
          <!-- Column header -->
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-dark-300">{{ TASK_STATUS_LABELS[status] }}</h2>
            <span class="text-xs text-dark-500 bg-dark-800 px-2 py-0.5 rounded-full">{{ getColumnTasks(status).length }}</span>
          </div>

          <!-- Cards -->
          <div class="flex-1 space-y-2 min-h-16 bg-dark-900/50 rounded-xl p-2">
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
              class="text-center py-8 text-xs text-dark-600"
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
      @close="showCreateModal = false"
      @created="loadTasks"
    />
  </div>
</template>
