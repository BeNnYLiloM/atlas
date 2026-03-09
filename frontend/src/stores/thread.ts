import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { messagesApi } from '@/api'
import type { Message } from '@/types'

export const useThreadStore = defineStore('thread', () => {
  // Текущий открытый тред (parent message)
  const currentThreadParentId = ref<string | null>(null)
  
  // Сообщения тредов: parentId -> Message[]
  const threadMessages = ref<Record<string, Message[]>>({})
  
  // Статистика тредов: parentId -> { count, lastReplyAt, participants, unreadCount }
  const threadStats = ref<Record<string, {
    count: number
    lastReplyAt: string
    lastReplyUser?: string
    unreadCount: number
  }>>({})

  const loading = ref(false)
  const error = ref<string | null>(null)

  const currentThread = computed(() => {
    if (!currentThreadParentId.value) return null
    return threadMessages.value[currentThreadParentId.value] || []
  })

  const isThreadOpen = computed(() => currentThreadParentId.value !== null)

  async function fetchThread(parentId: string) {
    loading.value = true
    error.value = null
    try {
      const messages = await messagesApi.getThread(parentId)
      threadMessages.value[parentId] = messages
      
      // Обновляем статистику
      if (messages.length > 0) {
        const lastReply = messages[messages.length - 1]
        threadStats.value[parentId] = {
          count: messages.length,
          lastReplyAt: lastReply.created_at,
          lastReplyUser: lastReply.user?.display_name,
          unreadCount: 0, // Обнулится при открытии
        }
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка загрузки треда'
    } finally {
      loading.value = false
    }
  }

  async function openThread(parentId: string) {
    currentThreadParentId.value = parentId
    if (!threadMessages.value[parentId]) {
      await fetchThread(parentId)
    }
    // Mark thread as read
    await markThreadAsRead(parentId)
  }

  async function markThreadAsRead(parentId: string) {
    try {
      const messages = threadMessages.value[parentId]
      const lastMessage = messages?.[messages.length - 1]
      await messagesApi.markThreadAsRead(parentId, lastMessage?.id)
      
      // Сбрасываем unread count
      if (threadStats.value[parentId]) {
        threadStats.value[parentId].unreadCount = 0
      }
    } catch (e) {
      console.error('Failed to mark thread as read:', e)
    }
  }

  async function fetchThreadUnreadCount(parentId: string) {
    try {
      const count = await messagesApi.getThreadUnreadCount(parentId)
      if (threadStats.value[parentId]) {
        threadStats.value[parentId].unreadCount = count
      } else {
        threadStats.value[parentId] = {
          count: 0,
          lastReplyAt: '',
          unreadCount: count,
        }
      }
      return count
    } catch (e) {
      console.error('Failed to fetch thread unread count:', e)
      return 0
    }
  }

  function closeThread() {
    currentThreadParentId.value = null
  }

  function addThreadReply(parentId: string, message: Message) {
    if (!threadMessages.value[parentId]) {
      threadMessages.value[parentId] = []
    }
    
    // Проверяем на дубликат
    const exists = threadMessages.value[parentId].some(m => m.id === message.id)
    if (!exists) {
      threadMessages.value[parentId].push(message)
      
      // Обновляем статистику
      const currentStats = threadStats.value[parentId] || { count: 0, lastReplyAt: '', unreadCount: 0 }
      threadStats.value[parentId] = {
        count: threadMessages.value[parentId].length,
        lastReplyAt: message.created_at,
        lastReplyUser: message.user?.display_name,
        unreadCount: currentThreadParentId.value === parentId ? 0 : currentStats.unreadCount + 1,
      }
    }
  }

  function updateThreadReply(parentId: string, message: Message) {
    if (!threadMessages.value[parentId]) return
    
    const index = threadMessages.value[parentId].findIndex(m => m.id === message.id)
    if (index !== -1) {
      threadMessages.value[parentId][index] = message
    }
  }

  function deleteThreadReply(parentId: string, messageId: string) {
    if (!threadMessages.value[parentId]) return
    
    const index = threadMessages.value[parentId].findIndex(m => m.id === messageId)
    if (index !== -1) {
      threadMessages.value[parentId].splice(index, 1)
      
      // Обновляем статистику
      const messages = threadMessages.value[parentId]
      if (messages.length > 0) {
        const lastReply = messages[messages.length - 1]
        const currentStats = threadStats.value[parentId] || { unreadCount: 0 }
        threadStats.value[parentId] = {
          count: messages.length,
          lastReplyAt: lastReply.created_at,
          lastReplyUser: lastReply.user?.display_name,
          unreadCount: currentStats.unreadCount,
        }
      } else {
        delete threadStats.value[parentId]
      }
    }
  }

  function getThreadStats(parentId: string) {
    return threadStats.value[parentId] || { count: 0, lastReplyAt: '', lastReplyUser: undefined, unreadCount: 0 }
  }

  function initThreadStats(parentId: string, count: number, unreadCount: number, lastReplyAt?: string) {
    threadStats.value[parentId] = {
      count,
      lastReplyAt: lastReplyAt || new Date().toISOString(),
      unreadCount,
    }
  }

  function $reset() {
    currentThreadParentId.value = null
    threadMessages.value = {}
    threadStats.value = {}
    loading.value = false
    error.value = null
  }

  return {
    currentThreadParentId,
    currentThread,
    isThreadOpen,
    threadStats,
    loading,
    error,
    openThread,
    closeThread,
    fetchThread,
    markThreadAsRead,
    fetchThreadUnreadCount,
    addThreadReply,
    updateThreadReply,
    deleteThreadReply,
    getThreadStats,
    initThreadStats,
    $reset,
  }
})
