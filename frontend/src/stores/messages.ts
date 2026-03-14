import { defineStore } from 'pinia'
import { ref } from 'vue'
import { messagesApi } from '@/api'
import type { Message, MessageCreate } from '@/types'
import { useThreadStore } from './thread'

export const useMessagesStore = defineStore('messages', () => {
  // Сообщения по каналам
  const messagesByChannel = ref<Record<string, Message[]>>({})
  const hasMoreByChannel = ref<Record<string, boolean>>({})
  const loading = ref(false)
  const loadingMore = ref(false)
  const sending = ref(false)
  const error = ref<string | null>(null)

  function getMessages(channelId: string): Message[] {
    const messages = messagesByChannel.value[channelId] ?? []
    // Сортируем по дате создания (старые сверху)
    return messages.sort((a, b) => 
      new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    )
  }

  async function fetchMessages(channelId: string, loadMore = false) {
    if (loadMore) {
      loadingMore.value = true
    } else {
      loading.value = true
    }
    error.value = null
    try {
      const limit = 50
      const currentMessages = messagesByChannel.value[channelId] ?? []
      const offset = loadMore ? currentMessages.length : 0
      const messages = await messagesApi.list(channelId, { limit, offset })

      // Если пришло меньше чем limit — больше нет
      hasMoreByChannel.value[channelId] = messages.length === limit

      if (loadMore) {
        // Подгрузка истории — добавляем в начало
        messagesByChannel.value[channelId] = [
          ...messages,
          ...currentMessages,
        ]
      } else {
        messagesByChannel.value[channelId] = messages
      }

      // Инициализируем threadStats из данных сообщений
      const threadStore = useThreadStore()
      for (const message of messages) {
        if (message.thread_replies_count && message.thread_replies_count > 0) {
          threadStore.initThreadStats(
            message.id,
            message.thread_replies_count,
            message.thread_unread_count || 0,
            message.created_at
          )
        }
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка загрузки сообщений'
    } finally {
      loading.value = false
      loadingMore.value = false
    }
  }

  async function sendMessage(data: MessageCreate) {
    sending.value = true
    error.value = null
    try {
      const message = await messagesApi.create(data)
      
      // Если это обычное сообщение (не тред), добавляем в messagesByChannel
      // Если тред - оно добавится через WebSocket в ThreadStore
      if (!data.parent_id) {
        addMessage(message)
      }
      
      return message
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка отправки сообщения'
      throw e
    } finally {
      sending.value = false
    }
  }

  function addMessage(message: Message) {
    const channelId = message.channel_id
    
    // Если это ответ в треде (есть parent_id), не добавляем в основной чат
    if (message.parent_id) {
      return
    }
    
    if (!messagesByChannel.value[channelId]) {
      messagesByChannel.value[channelId] = []
    }
    // Проверяем, нет ли уже такого сообщения (от WebSocket)
    const exists = messagesByChannel.value[channelId].some(m => m.id === message.id)
    if (!exists) {
      messagesByChannel.value[channelId].push(message)
    }
  }

  function updateMessage(message: Message) {
    const channelId = message.channel_id
    const messages = messagesByChannel.value[channelId]
    if (messages) {
      const index = messages.findIndex(m => m.id === message.id)
      if (index !== -1) {
        messages[index] = message
      }
    }
  }

  function deleteMessage(channelId: string, messageId: string) {
    const messages = messagesByChannel.value[channelId]
    if (messages) {
      const index = messages.findIndex(m => m.id === messageId)
      if (index !== -1) {
        messages.splice(index, 1)
      }
    }
  }

  // Обновляет поля call-сообщения по ID (без знания channel_id — ищем по всем каналам)
  function updateCallMessage(messageId: string, callStatus: 'ringing' | 'cancelled' | 'missed' | 'ongoing' | 'ended', durationSec: number | null) {
    for (const messages of Object.values(messagesByChannel.value)) {
      const msg = messages.find(m => m.id === messageId)
      if (msg) {
        msg.call_status = callStatus
        msg.call_duration_sec = durationSec ?? undefined
        break
      }
    }
  }

  function $reset() {
    messagesByChannel.value = {}
    loading.value = false
    sending.value = false
    error.value = null
  }

  return {
    messagesByChannel,
    hasMoreByChannel,
    loading,
    loadingMore,
    sending,
    error,
    getMessages,
    fetchMessages,
    sendMessage,
    addMessage,
    updateCallMessage,
    updateMessage,
    deleteMessage,
    $reset,
  }
})

