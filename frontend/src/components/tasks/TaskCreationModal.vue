<script setup lang="ts">
import { ref } from 'vue'
import { tasksApi } from '@/api/tasks'
import type { TaskPriority } from '@/api/tasks'
import { useWorkspaceStore } from '@/stores/workspace'

const props = defineProps<{
  messageId?: string
  initialTitle?: string
}>()

const emit = defineEmits<{
  close: []
  created: []
}>()

const workspaceStore = useWorkspaceStore()
const title = ref(props.initialTitle || '')
const description = ref('')
const priority = ref<TaskPriority>('medium')
const loading = ref(false)
const error = ref<string | null>(null)

async function submit() {
  if (!title.value.trim()) {
    error.value = 'Название задачи обязательно'
    return
  }

  loading.value = true
  error.value = null

  try {
    await tasksApi.create({
      message_id: props.messageId,
      workspace_id: workspaceStore.currentWorkspaceId ?? '',
      title: title.value.trim(),
      description: description.value || undefined,
      priority: priority.value,
    })
    emit('created')
    emit('close')
  } catch {
    error.value = 'Не удалось создать задачу'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-50 flex items-center justify-center px-4"
      role="dialog"
      aria-modal="true"
      aria-label="Создать задачу"
      @click.self="emit('close')"
    >
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div class="relative bg-dark-900 rounded-2xl border border-dark-700 shadow-2xl w-full max-w-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-dark-100">Создать задачу</h2>
          <button
            class="text-dark-500 hover:text-dark-300"
            aria-label="Закрыть"
            @click="emit('close')"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div v-if="messageId" class="mb-3 px-3 py-2 bg-atlas-600/10 border border-atlas-600/30 rounded-lg">
          <p class="text-xs text-atlas-300">Создаётся из сообщения</p>
        </div>

        <form @submit.prevent="submit" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-dark-300 mb-1" for="task-title">Название *</label>
            <input
              id="task-title"
              v-model="title"
              type="text"
              class="w-full px-3 py-2 bg-dark-800 border border-dark-700 rounded-lg text-dark-100 placeholder:text-dark-500 focus:outline-none focus:border-atlas-500"
              placeholder="Что нужно сделать?"
              autofocus
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-dark-300 mb-1" for="task-desc">Описание</label>
            <textarea
              id="task-desc"
              v-model="description"
              rows="3"
              class="w-full px-3 py-2 bg-dark-800 border border-dark-700 rounded-lg text-dark-100 placeholder:text-dark-500 focus:outline-none focus:border-atlas-500 resize-none"
              placeholder="Подробности..."
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-dark-300 mb-1">Приоритет</label>
            <div class="flex gap-2">
              <button
                v-for="p in (['low', 'medium', 'high', 'urgent'] as TaskPriority[])"
                :key="p"
                type="button"
                class="px-3 py-1 rounded-lg text-sm border transition-colors"
                :class="[
                  priority === p
                    ? 'bg-atlas-600 border-atlas-500 text-white'
                    : 'bg-dark-800 border-dark-700 text-dark-400 hover:border-dark-500'
                ]"
                @click="priority = p"
              >
                {{ { low: 'Низкий', medium: 'Средний', high: 'Высокий', urgent: 'Срочно' }[p] }}
              </button>
            </div>
          </div>

          <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

          <div class="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              class="px-4 py-2 text-sm text-dark-400 hover:text-dark-200 transition-colors"
              @click="emit('close')"
            >
              Отмена
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-atlas-600 hover:bg-atlas-500 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
              :disabled="loading"
            >
              {{ loading ? 'Создаём...' : 'Создать задачу' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>
