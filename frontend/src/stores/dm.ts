import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { dmApi, type DMChannel } from '@/api/dm'
import apiClient from '@/api/client'
import { useNavigationStore } from './navigation'
import { useWorkspaceStore } from './workspace'

export const useDMStore = defineStore('dm', () => {
  const router = useRouter()
  const navigationStore = useNavigationStore()
  const workspaceStore = useWorkspaceStore()

  const dmList = ref<DMChannel[]>([])
  const loading = ref(false)

  // Выставляется из DMView — ID канала, который сейчас открыт и виден пользователю
  const activeChannelId = ref<string | null>(null)

  // Список отсортирован по last_message_at DESC (с бэкенда), при WS-обновлениях пересортируем
  const sortedDMList = computed(() =>
    [...dmList.value].sort((a, b) => {
      const ta = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0
      const tb = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0
      return tb - ta
    }),
  )

  const totalUnread = computed(() =>
    dmList.value.reduce((sum, dm) => sum + dm.unreadCount, 0),
  )

  async function fetchDMs() {
    const wsId = workspaceStore.currentWorkspaceId
    if (!wsId) return
    loading.value = true
    try {
      dmList.value = await dmApi.list(wsId)
    } finally {
      loading.value = false
    }
  }

  async function openDM(targetUserId: string) {
    const wsId = workspaceStore.currentWorkspaceId
    if (!wsId) return
    const channel = await dmApi.open(wsId, targetUserId)
    await fetchDMs()
    navigationStore.setSection('dm')
    await router.push({ name: 'dm-channel', params: { channelId: channel.id } })
  }

  // channelId + messageId — ID нового сообщения (нужен для сортировки и unread)
  function onDMMessage(channelId: string, messageId: string) {
    const dm = dmList.value.find((d) => d.channelId === channelId)
    if (dm) {
      dm.lastMessageAt = new Date().toISOString()
      // Не увеличиваем счётчик если пользователь прямо сейчас смотрит этот чат
      if (channelId !== activeChannelId.value) {
        dm.unreadCount++
      }
    } else {
      void fetchDMs()
    }
    void messageId
  }

  // Сбрасываем счётчик через API с передачей последнего сообщения канала
  async function clearUnread(channelId: string, lastMessageId?: string) {
    const dm = dmList.value.find((d) => d.channelId === channelId)
    if (!dm) return
    dm.unreadCount = 0
    try {
      await apiClient.post(`/channels/${channelId}/read`, lastMessageId ? { message_id: lastMessageId } : {})
    } catch {
      // silent
    }
  }

  function updatePeerStatus(userId: string, status: string) {
    for (const dm of dmList.value) {
      if (dm.peer.userId === userId) {
        dm.peer.status = status
      }
    }
  }

  return {
    dmList: sortedDMList,
    activeChannelId,
    totalUnread,
    loading,
    fetchDMs,
    openDM,
    onDMMessage,
    clearUnread,
    updatePeerStatus,
  }
})
