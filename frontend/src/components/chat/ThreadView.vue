<script setup lang="ts">
import { computed, watch } from 'vue'
import { useThreadStore, useMessagesStore } from '@/stores'
import MessageInput from './MessageInput.vue'
import MessageItem from './MessageItem.vue'

const threadStore = useThreadStore()
const messagesStore = useMessagesStore()

const parentMessage = computed(() => {
  if (!threadStore.currentThreadParentId) return null
  
  // Ищем родительское сообщение во всех каналах
  for (const messages of Object.values(messagesStore.messagesByChannel)) {
    const found = messages.find(m => m.id === threadStore.currentThreadParentId)
    if (found) return found
  }
  
  return null
})

// Следим за изменением треда и загружаем его
watch(
  () => threadStore.currentThreadParentId,
  (newParentId) => {
    if (newParentId) {
      threadStore.fetchThread(newParentId)
    }
  },
  { immediate: true }
)

function closeThread() {
  threadStore.closeThread()
}

</script>

<template>
  <div
    v-if="threadStore.isThreadOpen"
    class="w-96 border-l border-subtle bg-surface flex flex-col"
  >
    <!-- Header -->
    <div class="px-4 py-3 border-b border-subtle flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg
          class="w-5 h-5 text-muted"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
          />
        </svg>
        <h3 class="font-semibold text-primary">
          Тред
        </h3>
      </div>
      <button
        class="p-1 rounded hover:bg-elevated text-muted hover:text-primary transition-colors"
        title="Закрыть"
        @click="closeThread"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>

    <!-- Parent message -->
    <div
      v-if="parentMessage"
      class="px-4 py-3 border-b border-subtle bg-elevated"
    >
      <MessageItem :message="parentMessage" />
    </div>

    <!-- Thread messages -->
    <div class="flex-1 overflow-y-auto px-4 py-3 space-y-1">
      <MessageItem
        v-for="message in threadStore.currentThread"
        :key="message.id"
        :message="message"
      />

      <!-- Empty state -->
      <div
        v-if="threadStore.currentThread?.length === 0 && !threadStore.loading"
        class="text-center py-8"
      >
        <svg
          class="w-12 h-12 mx-auto text-faint mb-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
          />
        </svg>
        <p class="text-subtle text-sm">
          Пока нет ответов
        </p>
        <p class="text-faint text-xs mt-1">
          Начните обсуждение
        </p>
      </div>

      <!-- Loading -->
      <div
        v-if="threadStore.loading"
        class="flex justify-center py-8"
      >
        <svg
          class="animate-spin w-6 h-6 text-accent"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      </div>
    </div>

    <!-- Input для ответа в треде -->
    <div
      v-if="parentMessage"
      class="border-t border-subtle"
    >
      <MessageInput
        :channel-id="parentMessage.channel_id"
        :parent-id="parentMessage.id"
        placeholder="Ответить в треде..."
      />
    </div>
  </div>
</template>
