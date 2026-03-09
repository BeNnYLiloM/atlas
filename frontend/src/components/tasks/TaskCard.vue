<script setup lang="ts">
import { tasksApi, TASK_PRIORITY_COLORS, TASK_STATUS_LABELS } from '@/api/tasks'
import type { Task, TaskStatus } from '@/api/tasks'

const props = defineProps<{
  task: Task
}>()

const emit = defineEmits<{
  updated: []
  deleted: []
}>()

async function changeStatus(status: TaskStatus) {
  await tasksApi.update(props.task.id, { status })
  emit('updated')
}

async function deleteTask() {
  await tasksApi.delete(props.task.id)
  emit('deleted')
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
}
</script>

<template>
  <div class="bg-dark-800 border border-dark-700 rounded-xl p-3 hover:border-dark-600 transition-colors group">
    <div class="flex items-start justify-between gap-2">
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-dark-100 line-clamp-2">{{ task.title }}</p>
        <p v-if="task.description" class="text-xs text-dark-500 mt-1 line-clamp-1">{{ task.description }}</p>
      </div>
      <button
        class="opacity-0 group-hover:opacity-100 text-dark-600 hover:text-red-400 transition-all"
        aria-label="Удалить задачу"
        @click="deleteTask"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="flex items-center gap-2 mt-2">
      <!-- Priority badge -->
      <span :class="['text-xs font-medium', TASK_PRIORITY_COLORS[task.priority]]">
        {{ { low: '↓', medium: '→', high: '↑', urgent: '⚡' }[task.priority] }}
      </span>

      <!-- Due date -->
      <span v-if="task.due_date" class="text-xs text-dark-500">
        {{ formatDate(task.due_date) }}
      </span>

      <!-- Status change -->
      <div class="ml-auto flex gap-1">
        <button
          v-if="task.status !== 'done'"
          class="text-xs text-dark-500 hover:text-green-400 transition-colors"
          title="Отметить готово"
          @click="changeStatus('done')"
        >✓</button>
        <button
          v-if="task.status === 'todo'"
          class="text-xs text-dark-500 hover:text-blue-400 transition-colors"
          title="В работу"
          @click="changeStatus('in_progress')"
        >▶</button>
      </div>
    </div>
  </div>
</template>
