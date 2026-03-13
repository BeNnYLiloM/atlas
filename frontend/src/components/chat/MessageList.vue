<script setup lang="ts">
import { computed, ref, nextTick, watch, onMounted } from 'vue'
import { useMessagesStore } from '@/stores'
import MessageItem from './MessageItem.vue'

const props = defineProps<{
  channelId: string
  highlightMessageId?: string | null
}>()

const messagesStore = useMessagesStore()
const containerRef = ref<HTMLElement | null>(null)

const messages = computed(() => messagesStore.getMessages(props.channelId))
const hasMore = computed(() => messagesStore.hasMoreByChannel[props.channelId] ?? false)

function scrollToBottom() {
  if (containerRef.value) {
    containerRef.value.scrollTop = containerRef.value.scrollHeight
  }
}

function scrollToMessage(messageId: string) {
  const tryScroll = () => {
    const el = containerRef.value?.querySelector(`[data-message-id="${messageId}"]`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
      return true
    }
    return false
  }
  // Сразу + с задержкой на случай если DOM ещё не готов
  if (!tryScroll()) {
    setTimeout(tryScroll, 150)
  }
}

function isNearBottom(): boolean {
  if (!containerRef.value) return true
  const { scrollTop, scrollHeight, clientHeight } = containerRef.value
  return scrollHeight - scrollTop - clientHeight < 100
}

async function loadMore() {
  if (!containerRef.value || messagesStore.loadingMore || !hasMore.value) return
  // Запоминаем высоту до загрузки чтобы сохранить позицию скролла
  const prevScrollHeight = containerRef.value.scrollHeight
  await messagesStore.fetchMessages(props.channelId, true)
  await nextTick()
  // Восстанавливаем позицию — скролл не прыгает вверх
  containerRef.value.scrollTop = containerRef.value.scrollHeight - prevScrollHeight
}

function onScroll() {
  if (!containerRef.value) return
  if (containerRef.value.scrollTop < 80) {
    loadMore()
  }
}

// При новых сообщениях — скроллим только если уже были внизу
watch(
  () => messages.value.length,
  async (newLen, oldLen) => {
    if (newLen === oldLen) return
    await nextTick()
    if (isNearBottom()) {
      scrollToBottom()
    }
  },
)

// При первом монтировании (перезагрузка страницы) — скроллим вниз
onMounted(async () => {
  await nextTick()
  scrollToBottom()
})

defineExpose({ scrollToBottom, scrollToMessage })

// Группировка сообщений по дате
const groupedMessages = computed(() => {
  const groups: { date: string; messages: typeof messages.value }[] = []
  let currentDate = ''

  for (const message of messages.value) {
    const date = new Date(message.created_at).toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })

    if (date !== currentDate) {
      currentDate = date
      groups.push({ date, messages: [] })
    }

    groups[groups.length - 1].messages.push(message)
  }

  return groups
})
</script>

<template>
  <div
    ref="containerRef"
    class="flex-1 overflow-y-auto px-4 py-4 space-y-4"
    @scroll="onScroll"
  >
    <!-- Загрузка истории (lazy load) -->
    <div
      v-if="messagesStore.loadingMore"
      class="flex justify-center py-3"
    >
      <svg
        class="animate-spin w-5 h-5 text-accent"
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
    <p
      v-else-if="!hasMore && messages.length > 0"
      class="text-center text-xs text-faint py-2"
    >
      Начало истории канала
    </p>

    <!-- Loading -->
    <div
      v-if="messagesStore.loading"
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

    <!-- Messages grouped by date -->
    <template v-else-if="groupedMessages.length > 0">
      <div
        v-for="group in groupedMessages"
        :key="group.date"
      >
        <!-- Date separator -->
        <div class="flex items-center gap-4 my-6">
          <div class="flex-1 h-px bg-elevated" />
          <span class="text-xs text-subtle font-medium">{{ group.date }}</span>
          <div class="flex-1 h-px bg-elevated" />
        </div>

        <!-- Messages -->
        <div class="space-y-1">
          <div
            v-for="message in group.messages"
            :key="message.id"
            :data-message-id="message.id"
          >
            <MessageItem
              :message="message"
              :highlighted="message.id === props.highlightMessageId"
            />
          </div>
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <div
      v-else
      class="flex flex-col items-center justify-center h-full text-center"
    >
      <div class="w-16 h-16 mb-4 rounded-2xl bg-elevated flex items-center justify-center">
        <svg
          class="w-8 h-8 text-accent"
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
      </div>
      <h3 class="text-lg font-semibold text-secondary mb-1">
        Начните общение!
      </h3>
      <p class="text-subtle text-sm max-w-xs">
        Это начало канала. Отправьте первое сообщение, чтобы начать разговор.
      </p>
    </div>
  </div>
</template>

