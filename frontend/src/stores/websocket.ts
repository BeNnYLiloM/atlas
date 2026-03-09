import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useMessagesStore } from './messages'
import { useChannelsStore } from './channels'
import { useThreadStore } from './thread'
import { useWorkspaceStore } from './workspace'
import { useAuthStore } from './auth'
import { playNotificationSound, isSoundEnabled, isMentionSoundEnabled } from '@/utils/notificationSound'
import { createAuthenticatedWebSocket } from '@/api/websocket'
import { ensureAccessToken } from '@/api/session'
import type { Message, Channel, ChannelCategory } from '@/types'

function showBrowserNotification(title: string, body: string, channelId: string) {
  if (!('Notification' in window) || Notification.permission !== 'granted') return

  const n = new Notification(title, {
    body,
    icon: '/favicon.ico',
    tag: channelId,
  })
  n.onclick = () => {
    window.focus()
    n.close()
  }
}

interface WSEvent {
  type: 'message' | 'message_updated' | 'message_deleted' | 'thread_reply' | 'channel_created' | 'channel_updated' | 'channel_deleted' | 'member_added' | 'member_removed' | 'member_updated' | 'workspace_updated' | 'typing' | 'presence' | 'reaction_added' | 'reaction_removed' | 'category_created' | 'category_updated' | 'category_deleted'
  payload: unknown
}

export const useWebSocketStore = defineStore('websocket', () => {
  const socket = ref<WebSocket | null>(null)
  const connected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5

  async function connect() {
    if (socket.value?.readyState === WebSocket.OPEN) {
      console.log('[WS] Already connected')
      return
    }

    const token = await ensureAccessToken()
    if (!token) {
      console.warn('[WS] No access token available for realtime connection')
      return
    }

    console.log('[WS] Connecting to realtime endpoint')
    socket.value = createAuthenticatedWebSocket(token)

    socket.value.onopen = () => {
      connected.value = true
      reconnectAttempts.value = 0
      console.log('[WS] ✓ Connected successfully')
    }

    socket.value.onmessage = (event) => {
      const lines: string[] = (event.data as string).split('\n').filter((l: string) => l.trim())
      for (const line of lines) {
        try {
          const data: WSEvent = JSON.parse(line)
          handleEvent(data)
        } catch (e) {
          console.error('[WS] Parse error:', e, line)
        }
      }
    }

    socket.value.onclose = () => {
      connected.value = false
      console.log('[WS] ✗ Disconnected')
      attemptReconnect()
    }

    socket.value.onerror = (error) => {
      console.error('[WS] ✗ Error:', error)
    }
  }

  function handleEvent(event: WSEvent) {
    console.log('[WS] ← Received:', event.type, event.payload)

    const messagesStore = useMessagesStore()
    const channelsStore = useChannelsStore()
    const threadStore = useThreadStore()
    const workspaceStore = useWorkspaceStore()

    switch (event.type) {
      case 'message': {
        const { channel_id, message } = event.payload as { channel_id: string, message: Message }
        const authStore = useAuthStore()

        messagesStore.addMessage(message)

        if (message.user_id === authStore.user?.id) break

        const level = channelsStore.getNotificationLevel(channel_id)
        const isCurrentChannel = channel_id === channelsStore.currentChannelId
        const myDisplayName = authStore.user?.display_name ?? ''
        const isMentionedInMessage =
          (myDisplayName !== '' && message.content.includes(`@${myDisplayName}`)) ||
          message.content.includes('@everyone')

        if (!isCurrentChannel) {
          if (level === 'all') {
            channelsStore.incrementUnread(channel_id)
          } else if (level === 'mentions') {
            channelsStore.incrementUnread(channel_id)
          }
          if (isMentionedInMessage) {
            channelsStore.incrementMention(channel_id)
          }
        }

        const shouldNotify =
          level === 'all' ||
          (level === 'mentions' && isMentionedInMessage)

        if (shouldNotify) {
          if (isMentionedInMessage && isMentionSoundEnabled()) {
            playNotificationSound('mention')
          } else if (!isMentionedInMessage && isSoundEnabled()) {
            playNotificationSound('message')
          }

          if (document.visibilityState !== 'visible') {
            const channel = channelsStore.channels.find(c => c.id === channel_id)
            const channelName = channel ? `#${channel.name}` : 'канал'
            const senderName = message.user?.display_name ?? 'Кто-то'
            showBrowserNotification(
              `${senderName} в ${channelName}`,
              message.content.slice(0, 100),
              channel_id,
            )
          }
        }
        break
      }

      case 'thread_reply': {
        const { parent_id, message } = event.payload as { channel_id: string, parent_id: string, message: Message }
        threadStore.addThreadReply(parent_id, message)
        console.log('[WS] Thread reply added to parent:', parent_id)
        break
      }

      case 'message_updated': {
        const { message } = event.payload as { channel_id: string, message: Message }
        messagesStore.updateMessage(message)

        if (message.parent_id) {
          threadStore.updateThreadReply(message.parent_id, message)
        }
        break
      }

      case 'message_deleted': {
        const { channel_id, message_id } = event.payload as { channel_id: string; message_id: string }
        messagesStore.deleteMessage(channel_id, message_id)
        break
      }

      case 'channel_created': {
        const channel = event.payload as Channel
        channelsStore.addChannel(channel)
        console.log('[WS] Channel created:', channel.name)
        break
      }

      case 'channel_updated': {
        const channel = event.payload as Channel
        channelsStore.updateChannel(channel)
        console.log('[WS] Channel updated:', channel.name)
        break
      }

      case 'channel_deleted': {
        const { channel_id } = event.payload as { workspace_id: string, channel_id: string }
        channelsStore.removeChannel(channel_id)
        console.log('[WS] Channel deleted:', channel_id)
        break
      }

      case 'category_created': {
        const cat = event.payload as ChannelCategory
        channelsStore.addCategory(cat)
        break
      }

      case 'category_updated': {
        const cat = event.payload as ChannelCategory
        channelsStore.updateCategory(cat)
        break
      }

      case 'category_deleted': {
        const { category_id } = event.payload as { workspace_id: string; category_id: string }
        channelsStore.removeCategory(category_id)
        break
      }

      case 'reaction_added':
      case 'reaction_removed': {
        console.log('[WS] Reaction event:', event.type, event.payload)
        break
      }

      case 'workspace_updated': {
        const workspace = event.payload as import('@/types').Workspace
        workspaceStore.applyWorkspaceUpdate(workspace)
        break
      }

      case 'member_added': {
        const data = event.payload as { workspace_id: string; user_id: string; role: string }
        workspaceStore.addMember(data)
        console.log('[WS] Member added:', data.user_id)
        break
      }

      case 'member_removed': {
        const data = event.payload as { workspace_id: string; user_id: string }
        workspaceStore.applyMemberRemove(data)
        console.log('[WS] Member removed:', data.user_id)
        break
      }

      case 'member_updated': {
        const data = event.payload as { workspace_id: string; user_id: string; role?: string; nickname?: string | null }
        if (data.role || data.nickname !== undefined) {
          workspaceStore.applyMemberUpdate(data)
        } else {
          workspaceStore.fetchMembers(data.workspace_id)
        }
        break
      }

      case 'typing': {
        const { channel_id, user_id, typing } = event.payload as { channel_id: string, user_id: string, typing: boolean }
        channelsStore.setUserTyping(channel_id, user_id, typing)
        break
      }

      case 'presence': {
        const { user_id, status } = event.payload as { user_id: string, status: string }
        workspaceStore.setPresence(user_id, status)
        console.log('[WS] Presence update:', user_id, '->', status)
        break
      }

      default:
        console.warn('[WS] Unknown event type:', event)
    }
  }

  function attemptReconnect() {
    if (reconnectAttempts.value >= maxReconnectAttempts) {
      console.log('[WS] Max reconnect attempts reached')
      return
    }

    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), 30000)
    reconnectAttempts.value++

    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${reconnectAttempts.value})`)
    setTimeout(() => {
      void connect()
    }, delay)
  }

  function disconnect() {
    if (socket.value) {
      socket.value.close()
      socket.value = null
      connected.value = false
      console.log('[WS] Disconnected manually')
    }
  }

  function send(event: string, data: unknown) {
    if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
      console.warn('[WS] Cannot send, socket not ready. State:', socket.value?.readyState)
      return false
    }

    const message = JSON.stringify({ event, data })
    socket.value.send(message)
    console.log('[WS] → Sent:', event, data)
    return true
  }

  function subscribeToWorkspace(workspaceId: string) {
    console.log('[WS] 📡 Subscribing to workspace:', workspaceId)

    if (!connected.value || socket.value?.readyState !== WebSocket.OPEN) {
      console.warn('[WS] WebSocket not ready, waiting...')
      setTimeout(() => subscribeToWorkspace(workspaceId), 100)
      return
    }

    send('subscribe_workspace', { workspace_id: workspaceId })
  }

  function unsubscribeFromWorkspace(workspaceId: string) {
    console.log('[WS] 📡 Unsubscribing from workspace:', workspaceId)
    send('unsubscribe_workspace', { workspace_id: workspaceId })
  }

  function subscribe(channelId: string) {
    console.log('[WS] 📡 Subscribing to channel:', channelId)
    send('subscribe', { channel_id: channelId })
  }

  function unsubscribe(channelId: string) {
    console.log('[WS] 📡 Unsubscribing from channel:', channelId)
    send('unsubscribe', { channel_id: channelId })
  }

  return {
    socket,
    connected,
    connect,
    disconnect,
    send,
    subscribeToWorkspace,
    unsubscribeFromWorkspace,
    subscribe,
    unsubscribe,
  }
})
