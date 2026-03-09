<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { NotificationLevel } from '@/types'

defineProps<{
  x: number
  y: number
  channelId: string
  notificationLevel: NotificationLevel
  isAdmin: boolean
}>()

const emit = defineEmits<{
  close: []
  setNotification: [level: NotificationLevel]
  mute: [minutes: number | null]
  openSettings: []
  markRead: []
  deleteChannel: []
}>()

const submenu = ref<'mute' | 'notifications' | null>(null)

const muteOptions: { label: string; minutes: number | null }[] = [
  { label: '15 минут', minutes: 15 },
  { label: '1 час', minutes: 60 },
  { label: '3 часа', minutes: 180 },
  { label: '8 часов', minutes: 480 },
  { label: '24 часа', minutes: 1440 },
  { label: 'До отмены', minutes: null },
]

const notifOptions: { label: string; description: string; value: NotificationLevel }[] = [
  { label: 'Все сообщения', description: 'Уведомлять о каждом сообщении', value: 'all' },
  { label: 'Только упоминания', description: 'Только когда упоминают вас', value: 'mentions' },
  { label: 'Ничего', description: 'Не получать уведомлений', value: 'nothing' },
]

function onClickOutside(e: MouseEvent) {
  const el = document.getElementById('channel-ctx-menu')
  if (el && !el.contains(e.target as Node)) {
    emit('close')
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside))
onUnmounted(() => document.removeEventListener('mousedown', onClickOutside))
</script>

<template>
  <Teleport to="body">
    <div
      id="channel-ctx-menu"
      class="fixed z-[200] w-56 bg-dark-800 border border-dark-600 rounded-lg shadow-2xl py-1 text-sm"
      :style="{ top: `${y}px`, left: `${x}px` }"
    >
      <!-- Пометить как прочитанное -->
      <button
        class="w-full flex items-center px-3 py-1.5 text-dark-200 hover:bg-dark-700 hover:text-white transition-colors"
        @mouseenter="submenu = null"
        @click="emit('markRead'); emit('close')"
      >
        Пометить как прочитанное
      </button>

      <div class="my-1 border-t border-dark-700" />

      <!-- Заглушить канал -->
      <div
        class="relative"
        @mouseenter="submenu = 'mute'"
      >
        <button
          class="w-full flex items-center justify-between px-3 py-1.5 transition-colors"
          :class="submenu === 'mute' ? 'bg-dark-700 text-white' : 'text-dark-200 hover:bg-dark-700 hover:text-white'"
        >
          <span>Заглушить канал</span>
          <svg
            class="w-3.5 h-3.5 text-dark-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5l7 7-7 7"
            />
          </svg>
        </button>

        <!-- Подменю: заглушить -->
        <div
          v-if="submenu === 'mute'"
          class="absolute left-full top-0 ml-1 w-44 bg-dark-800 border border-dark-600 rounded-lg shadow-2xl py-1"
        >
          <button
            v-for="opt in muteOptions"
            :key="opt.label"
            class="w-full flex items-center px-3 py-1.5 text-dark-200 hover:bg-dark-700 hover:text-white transition-colors"
            @click="emit('mute', opt.minutes); emit('close')"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>

      <!-- Параметры уведомлений -->
      <div
        class="relative"
        @mouseenter="submenu = 'notifications'"
      >
        <button
          class="w-full flex items-center justify-between px-3 py-1.5 transition-colors"
          :class="submenu === 'notifications' ? 'bg-dark-700 text-white' : 'text-dark-200 hover:bg-dark-700 hover:text-white'"
        >
          <div class="text-left">
            <p>Параметры уведомлений</p>
            <p class="text-xs text-dark-400 mt-0.5">
              {{ notifOptions.find(o => o.value === notificationLevel)?.label }}
            </p>
          </div>
          <svg
            class="w-3.5 h-3.5 text-dark-400 shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5l7 7-7 7"
            />
          </svg>
        </button>

        <!-- Подменю: уведомления -->
        <div
          v-if="submenu === 'notifications'"
          class="absolute left-full top-0 ml-1 w-52 bg-dark-800 border border-dark-600 rounded-lg shadow-2xl py-1"
        >
          <button
            v-for="opt in notifOptions"
            :key="opt.value"
            class="w-full flex items-center gap-2 px-3 py-2 transition-colors"
            :class="notificationLevel === opt.value
              ? 'text-atlas-400 bg-atlas-600/10'
              : 'text-dark-200 hover:bg-dark-700 hover:text-white'"
            @click="emit('setNotification', opt.value); emit('close')"
          >
            <svg
              class="w-3.5 h-3.5 shrink-0"
              :class="notificationLevel === opt.value ? 'text-atlas-400' : 'text-transparent'"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
            <div class="text-left">
              <p class="text-sm">
                {{ opt.label }}
              </p>
              <p class="text-xs text-dark-500">
                {{ opt.description }}
              </p>
            </div>
          </button>
        </div>
      </div>

      <!-- Настройки и удаление канала (только admin/owner) -->
      <template v-if="isAdmin">
        <div class="my-1 border-t border-dark-700" />
        <button
          class="w-full flex items-center px-3 py-1.5 text-dark-200 hover:bg-dark-700 hover:text-white transition-colors"
          @mouseenter="submenu = null"
          @click="emit('openSettings'); emit('close')"
        >
          Настройки канала
        </button>
        <button
          class="w-full flex items-center px-3 py-1.5 text-red-400 hover:bg-red-500/10 transition-colors"
          @mouseenter="submenu = null"
          @click="emit('deleteChannel'); emit('close')"
        >
          Удалить канал
        </button>
      </template>
    </div>
  </Teleport>
</template>


