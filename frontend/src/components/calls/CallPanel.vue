<script setup lang="ts">
import { useCallsStore } from '@/stores/calls'
import { useChannelsStore } from '@/stores'

const callsStore = useCallsStore()
const channelsStore = useChannelsStore()

// Имя текущего голосового канала
function currentChannelName(): string {
  if (!callsStore.currentChannelId) return 'Голосовой канал'
  const ch = channelsStore.channels.find(c => c.id === callsStore.currentChannelId)
  return ch?.name ?? 'Голосовой канал'
}
</script>

<template>
  <!-- Панель активного голосового канала — встраивается в нижнюю часть сайдбара (как в Discord) -->
  <div
    v-if="callsStore.isInCall"
    class="border-t border-subtle bg-base p-2"
  >
    <!-- Статус подключения -->
    <div class="flex items-center gap-2 px-2 py-1 mb-1">
      <div class="w-2 h-2 rounded-full bg-green-400 animate-pulse flex-shrink-0" />
      <div class="flex-1 min-w-0">
        <p class="text-xs font-medium text-green-400 truncate">
          {{ currentChannelName() }}
        </p>
        <p class="text-xs text-subtle">
          Голос подключён
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
        title="Выйти из канала"
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
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
          />
        </svg>
      </button>
    </div>
  </div>
</template>
