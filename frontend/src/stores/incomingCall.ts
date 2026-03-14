import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useIncomingCallStore = defineStore('incomingCall', () => {
  const channelId = ref<string | null>(null)
  const callerName = ref('')
  const callerAvatar = ref<string | null>(null)
  const callerId = ref<string | null>(null)

  const isRinging = computed(() => channelId.value !== null)

  function ring(payload: { channelId: string; callerName: string; callerAvatar: string | null; callerId: string }) {
    channelId.value = payload.channelId
    callerName.value = payload.callerName
    callerAvatar.value = payload.callerAvatar
    callerId.value = payload.callerId
  }

  function clear() {
    channelId.value = null
    callerName.value = ''
    callerAvatar.value = null
    callerId.value = null
  }

  return { channelId, callerName, callerAvatar, callerId, isRinging, ring, clear }
})
