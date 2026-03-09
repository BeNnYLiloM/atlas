<script setup lang="ts">
import { RouterView } from 'vue-router'
import { unlockAudioContext } from '@/utils/notificationSound'

// Разблокируем AudioContext при первом взаимодействии пользователя
// (требование браузеров — звук нельзя воспроизводить до первого gesture)
function handleFirstInteraction() {
  unlockAudioContext()
  // Запрашиваем разрешение на browser notifications
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission()
  }
  window.removeEventListener('click', handleFirstInteraction)
  window.removeEventListener('keydown', handleFirstInteraction)
}

window.addEventListener('click', handleFirstInteraction)
window.addEventListener('keydown', handleFirstInteraction)
</script>

<template>
  <RouterView />
</template>

