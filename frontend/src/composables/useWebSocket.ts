import { ref, onUnmounted } from 'vue'
import { useMessagesStore } from '@/stores'
import { createAuthenticatedWebSocket } from '@/api/websocket'
import type { Message } from '@/types'

interface WSEvent {
  type: 'message' | 'message_update' | 'message_delete' | 'typing' | 'presence'
  payload: unknown
}

export function useWebSocket() {
  const socket = ref<WebSocket | null>(null)
  const connected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5

  const messagesStore = useMessagesStore()

  function connect() {
    const token = localStorage.getItem('atlas_token')
    if (!token) return

    socket.value = createAuthenticatedWebSocket(token)

    socket.value.onopen = () => {
      connected.value = true
      reconnectAttempts.value = 0
      console.log('[WS] Connected')
    }

    socket.value.onmessage = (event) => {
      try {
        const data: WSEvent = JSON.parse(event.data)
        handleEvent(data)
      } catch (e) {
        console.error('[WS] Parse error:', e)
      }
    }

    socket.value.onclose = () => {
      connected.value = false
      console.log('[WS] Disconnected')
      attemptReconnect()
    }

    socket.value.onerror = (error) => {
      console.error('[WS] Error:', error)
    }
  }

  function handleEvent(event: WSEvent) {
    console.log('[WS] Received event:', event.type, event.payload)

    switch (event.type) {
      case 'message':
        messagesStore.addMessage(event.payload as Message)
        break
      case 'message_update':
        messagesStore.updateMessage(event.payload as Message)
        break
      case 'message_delete': {
        const { channel_id, message_id } = event.payload as { channel_id: string; message_id: string }
        messagesStore.deleteMessage(channel_id, message_id)
        break
      }
      case 'typing': {
        console.log('[WS] Typing event:', event.payload)
        break
      }
      case 'presence': {
        console.log('[WS] Presence event:', event.payload)
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
    setTimeout(connect, delay)
  }

  function disconnect() {
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }
  }

  function send(event: string, data: unknown) {
    if (socket.value?.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify({ event, data }))
    } else {
      console.warn('[WS] Cannot send, socket not open:', event)
    }
  }

  function subscribe(channelId: string) {
    console.log('[WS] Subscribing to channel:', channelId)
    send('subscribe', { channel_id: channelId })
  }

  function unsubscribe(channelId: string) {
    console.log('[WS] Unsubscribing from channel:', channelId)
    send('unsubscribe', { channel_id: channelId })
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    connect,
    disconnect,
    send,
    subscribe,
    unsubscribe,
  }
}
