<script setup lang="ts">
import { watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCallsStore } from '@/stores/calls'
import { useIncomingCallStore } from '@/stores/incomingCall'
import { Avatar } from '@/components/ui'

const router = useRouter()
const callsStore = useCallsStore()
const incomingCall = useIncomingCallStore()

// Воспроизводим звонок пока показано уведомление
let audio: HTMLAudioElement | null = null

watch(
  () => incomingCall.isRinging,
  (ringing) => {
    if (ringing) {
      audio = new Audio('/sounds/ringtone.mp3')
      audio.loop = true
      audio.volume = 0.6
      audio.play().catch(() => {/* автовоспроизведение заблокировано — ок */})
    } else {
      audio?.pause()
      audio = null
    }
  },
)

async function accept() {
  if (!incomingCall.channelId) return
  const channelId = incomingCall.channelId
  incomingCall.clear()
  await callsStore.joinCall(channelId, true, false)
  // Переходим в DM чат если не там
  await router.push({ name: 'dm-channel', params: { channelId } })
}

function decline() {
  incomingCall.clear()
}
</script>

<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="opacity-0 translate-y-4 scale-95"
    enter-to-class="opacity-100 translate-y-0 scale-100"
    leave-active-class="transition-all duration-200 ease-in"
    leave-from-class="opacity-100 translate-y-0 scale-100"
    leave-to-class="opacity-0 translate-y-4 scale-95"
  >
    <div
      v-if="incomingCall.isRinging"
      class="fixed bottom-6 right-6 z-50 w-80 rounded-2xl bg-overlay border border-default shadow-2xl overflow-hidden"
    >
      <!-- Animated ring indicator -->
      <div class="h-1 bg-accent animate-pulse" />

      <div class="p-4">
        <p class="text-xs text-subtle uppercase tracking-wider font-semibold mb-3">Входящий звонок</p>

        <div class="flex items-center gap-3 mb-4">
          <div class="relative">
            <Avatar
              :name="incomingCall.callerName"
              :src="incomingCall.callerAvatar ?? undefined"
              size="md"
            />
            <!-- Пульсирующий ring -->
            <span class="absolute inset-0 rounded-full ring-2 ring-accent animate-ping opacity-50" />
          </div>
          <div class="min-w-0">
            <p class="font-semibold text-primary truncate">{{ incomingCall.callerName }}</p>
            <p class="text-sm text-subtle">Личный звонок</p>
          </div>
        </div>

        <div class="flex gap-2">
          <!-- Принять -->
          <button
            class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-xl bg-green-600 hover:bg-green-500 text-white font-medium text-sm transition-colors"
            @click="accept"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
            </svg>
            Принять
          </button>

          <!-- Отклонить -->
          <button
            class="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-xl bg-red-600/20 hover:bg-red-600 text-red-400 hover:text-white font-medium text-sm transition-colors"
            @click="decline"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2M5 3a2 2 0 00-2 2v1c0 8.284 6.716 15 15 15h1a2 2 0 002-2v-3.28a1 1 0 00-.684-.948l-4.493-1.498a1 1 0 00-1.21.502l-1.13 2.257a11.042 11.042 0 01-5.516-5.517l2.257-1.128a1 1 0 00.502-1.21L9.228 3.683A1 1 0 008.279 3H5z" />
            </svg>
            Отклонить
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>
