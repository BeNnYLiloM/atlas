<script setup lang="ts">
import { computed } from 'vue'
import { useCallsStore } from '@/stores/calls'
import { useChannelsStore } from '@/stores'
import { useDMStore } from '@/stores/dm'

const callsStore = useCallsStore()
const channelsStore = useChannelsStore()
const dmStore = useDMStore()

const callLabel = computed(() => {
  if (!callsStore.currentChannelId) return 'Голосовой канал'
  // Проверяем DM-чат
  const dm = dmStore.dmList.find(d => d.channelId === callsStore.currentChannelId)
  if (dm) return dm.peer.displayName
  // Обычный voice-канал
  return channelsStore.channels.find(c => c.id === callsStore.currentChannelId)?.name ?? 'Голосовой канал'
})

const callSubtitle = computed(() => {
  const dm = dmStore.dmList.find(d => d.channelId === callsStore.currentChannelId)
  return dm ? 'Личный звонок' : 'Голос подключён'
})
</script>

<template>
  <!-- Панель активного звонка — встраивается в нижнюю часть сайдбара -->
  <div
    v-if="callsStore.isInCall"
    class="border-t border-subtle bg-base p-2"
  >
    <!-- Статус подключения -->
    <div class="flex items-center gap-2 px-2 py-1 mb-1">
      <div class="w-2 h-2 rounded-full bg-green-400 animate-pulse flex-shrink-0" />
      <div class="flex-1 min-w-0">
        <p class="text-xs font-medium text-green-400 truncate">
          {{ callLabel }}
        </p>
        <p class="text-xs text-subtle">
          {{ callSubtitle }}
        </p>
      </div>
    </div>

    <!-- Кнопки управления -->
    <div class="flex items-center gap-1 px-1">
      <!-- Mute -->
      <button
        class="flex-1 flex items-center justify-center gap-1 py-1.5 rounded-lg text-xs transition-colors"
        :class="callsStore.isMuted ? 'bg-red-600/20 text-red-400' : 'text-muted hover:bg-elevated hover:text-primary'"
        :title="callsStore.isMuted ? 'Включить микрофон' : 'Выключить микрофон'"
        @click="callsStore.toggleMute()"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            v-if="!callsStore.isMuted"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
          />
          <path
            v-else
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2"
          />
        </svg>
        {{ callsStore.isMuted ? 'Вкл. mic' : 'Выкл. mic' }}
      </button>

      <!-- Disconnect -->
      <button
        class="flex items-center justify-center w-8 h-8 rounded-lg bg-red-600/20 text-red-400 hover:bg-red-600 hover:text-white transition-colors"
        title="Завершить звонок"
        @click="callsStore.leaveCall()"
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
            d="M16 8l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2M5 3a2 2 0 00-2 2v1c0 8.284 6.716 15 15 15h1a2 2 0 002-2v-3.28a1 1 0 00-.684-.948l-4.493-1.498a1 1 0 00-1.21.502l-1.13 2.257a11.042 11.042 0 01-5.516-5.517l2.257-1.128a1 1 0 00.502-1.21L9.228 3.683A1 1 0 008.279 3H5z"
          />
        </svg>
      </button>
    </div>
  </div>
</template>
