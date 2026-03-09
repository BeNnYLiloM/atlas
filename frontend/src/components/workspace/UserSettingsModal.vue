<script setup lang="ts">
import { ref, watch } from 'vue'
import { Modal } from '@/components/ui'
import {
  isSoundEnabled,
  isMentionSoundEnabled,
  setSoundEnabled,
  setMentionSoundEnabled,
  playNotificationSound,
} from '@/utils/notificationSound'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const soundEnabled = ref(isSoundEnabled())
const mentionSoundEnabled = ref(isMentionSoundEnabled())

watch(
  () => props.open,
  (open) => {
    if (open) {
      soundEnabled.value = isSoundEnabled()
      mentionSoundEnabled.value = isMentionSoundEnabled()
    }
  },
)

function toggleSound() {
  soundEnabled.value = !soundEnabled.value
  setSoundEnabled(soundEnabled.value)
  if (soundEnabled.value) playNotificationSound('message')
}

function toggleMentionSound() {
  mentionSoundEnabled.value = !mentionSoundEnabled.value
  setMentionSoundEnabled(mentionSoundEnabled.value)
  if (mentionSoundEnabled.value) playNotificationSound('mention')
}
</script>

<template>
  <Modal :open="open" title="Настройки уведомлений" @close="emit('close')">
    <div class="space-y-6 py-2">

      <!-- Звуки сообщений -->
      <div class="space-y-3">
        <h3 class="text-sm font-semibold text-dark-200 uppercase tracking-wide">Звуки</h3>

        <label class="flex items-center justify-between gap-4 cursor-pointer group">
          <div>
            <p class="text-sm font-medium text-dark-100">Звук новых сообщений</p>
            <p class="text-xs text-dark-400 mt-0.5">Тихий пинг при получении сообщения в канале</p>
          </div>
          <button
            type="button"
            class="relative inline-flex h-6 w-11 shrink-0 rounded-full transition-colors duration-200 focus:outline-none"
            :class="soundEnabled ? 'bg-atlas-600' : 'bg-dark-600'"
            @click="toggleSound"
          >
            <span
              class="inline-block h-5 w-5 mt-0.5 rounded-full bg-white shadow transform transition-transform duration-200"
              :class="soundEnabled ? 'translate-x-5' : 'translate-x-0.5'"
            />
          </button>
        </label>

        <label class="flex items-center justify-between gap-4 cursor-pointer group">
          <div>
            <p class="text-sm font-medium text-dark-100">Звук упоминаний</p>
            <p class="text-xs text-dark-400 mt-0.5">Двойной пинг когда вас упоминают (@имя, @everyone)</p>
          </div>
          <button
            type="button"
            class="relative inline-flex h-6 w-11 shrink-0 rounded-full transition-colors duration-200 focus:outline-none"
            :class="mentionSoundEnabled ? 'bg-atlas-600' : 'bg-dark-600'"
            @click="toggleMentionSound"
          >
            <span
              class="inline-block h-5 w-5 mt-0.5 rounded-full bg-white shadow transform transition-transform duration-200"
              :class="mentionSoundEnabled ? 'translate-x-5' : 'translate-x-0.5'"
            />
          </button>
        </label>
      </div>

      <p class="text-xs text-dark-500">
        При переключении воспроизводится предпросмотр звука.
        Настройки сохраняются в браузере.
      </p>
    </div>
  </Modal>
</template>
