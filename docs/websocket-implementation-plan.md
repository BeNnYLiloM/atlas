# План реализации WebSocket архитектуры (Discord-подход)

## 🎯 Цель
Реализовать подписку на уровне workspace с получением всех событий: новые сообщения в любых каналах, создание/удаление каналов, изменения участников и т.д.

---

## 📋 Этап 1: Рефакторинг Hub на Backend

### 1.1 Изменить структуру подписок
**Файл:** `backend/internal/transport/ws/hub.go`

**Текущее состояние:**
```go
type Hub struct {
    channels map[string]map[string]*Client  // channelID -> clientID -> Client
    users    map[string][]*Client           // userID -> Clients
}
```

**Новое состояние:**
```go
type Hub struct {
    // Workspace-level subscriptions (основное)
    workspaces map[string]map[string]*Client  // workspaceID -> clientID -> Client
    
    // User-level subscriptions (для mentions, DMs)
    users map[string][]*Client                // userID -> Clients
    
    // Дополнительно: channel subscriptions для typing indicators
    channels map[string]map[string]*Client    // channelID -> clientID -> Client (опционально)
}
```

### 1.2 Расширить типы событий
**Добавить константы:**
```go
// Workspace-level события
EventWorkspaceUpdated    = "workspace_updated"
EventChannelCreated      = "channel_created"
EventChannelUpdated      = "channel_updated"
EventChannelDeleted      = "channel_deleted"
EventMemberAdded         = "member_added"
EventMemberRemoved       = "member_removed"
EventMemberUpdated       = "member_updated"

// Channel-level события (текущие)
EventMessage             = "message"
EventMessageUpdated      = "message_updated"
EventMessageDeleted      = "message_deleted"

// User-level события
EventTypingStart         = "typing_start"
EventTypingStop          = "typing_stop"
EventPresenceUpdate      = "presence_update"
EventMention             = "mention"
```

### 1.3 Новые методы Hub
```go
// SubscribeToWorkspace - подписать клиента на workspace
func (h *Hub) SubscribeToWorkspace(client *Client, workspaceID string)

// UnsubscribeFromWorkspace - отписать от workspace
func (h *Hub) UnsubscribeFromWorkspace(client *Client, workspaceID string)

// BroadcastToWorkspace - рассылка всем участникам workspace
func (h *Hub) BroadcastToWorkspace(workspaceID, event string, data interface{}, excludeID string)

// BroadcastToWorkspaceWithFilter - рассылка с фильтрацией по правам
func (h *Hub) BroadcastToWorkspaceWithFilter(workspaceID, event string, data interface{}, filter func(*Client) bool)
```

---

## 📋 Этап 2: Обновить Client обработку

### 2.1 Расширить handleMessage
**Файл:** `backend/internal/transport/ws/client.go`

**Добавить обработку новых событий:**
```go
case "subscribe_workspace":
    var data struct {
        WorkspaceID string `json:"workspace_id"`
    }
    // Подписать клиента на workspace
    c.hub.SubscribeToWorkspace(c, data.WorkspaceID)

case "unsubscribe_workspace":
    var data struct {
        WorkspaceID string `json:"workspace_id"`
    }
    c.hub.UnsubscribeFromWorkspace(c, data.WorkspaceID)

case "subscribe_channel":
    // Для typing indicators (опционально)
    var data struct {
        ChannelID string `json:"channel_id"`
    }
    c.hub.SubscribeToChannel(c, data.ChannelID)

case "typing":
    var data struct {
        ChannelID string `json:"channel_id"`
    }
    // Broadcast только в channel, не в весь workspace
    c.hub.BroadcastToChannel(data.ChannelID, "typing_start", ...)
```

### 2.2 Хранить workspace memberships в Client
```go
type Client struct {
    ID            string
    UserID        string
    hub           *Hub
    conn          *websocket.Conn
    send          chan []byte
    workspaces    map[string]bool  // Список подписанных workspace
    channels      map[string]bool  // Список подписанных каналов (для typing)
}
```

---

## 📋 Этап 3: Добавить Broadcast в Handlers

### 3.1 ChannelHandler
**Файл:** `backend/internal/transport/http/handler/channel.go`

**Create:**
```go
func (h *ChannelHandler) Create(c *gin.Context) {
    // ... создание канала ...
    
    // Broadcast в workspace
    h.wsHub.BroadcastToWorkspace(
        channel.WorkspaceID,
        "channel_created",
        channel,
        "",
    )
}
```

**Update:**
```go
func (h *ChannelHandler) Update(c *gin.Context) {
    // ... обновление канала ...
    
    h.wsHub.BroadcastToWorkspace(
        channel.WorkspaceID,
        "channel_updated",
        channel,
        "",
    )
}
```

**Delete:**
```go
func (h *ChannelHandler) Delete(c *gin.Context) {
    // ... удаление канала ...
    
    h.wsHub.BroadcastToWorkspace(
        workspaceID,
        "channel_deleted",
        map[string]string{
            "workspace_id": workspaceID,
            "channel_id":   channelID,
        },
        "",
    )
}
```

### 3.2 WorkspaceHandler
**Файл:** `backend/internal/transport/http/handler/workspace.go`

**AddMember:**
```go
func (h *WorkspaceHandler) AddMember(c *gin.Context) {
    // ... добавление участника ...
    
    h.wsHub.BroadcastToWorkspace(
        workspaceID,
        "member_added",
        map[string]interface{}{
            "workspace_id": workspaceID,
            "user":         user,
            "role":         role,
        },
        "",
    )
}
```

**RemoveMember:**
```go
// Аналогично - broadcast member_removed
```

### 3.3 MessageHandler
**Изменить broadcast:**
```go
func (h *MessageHandler) Create(c *gin.Context) {
    // ... создание сообщения ...
    
    // Получаем workspaceID из channel
    channel, _ := h.channelService.GetByID(ctx, message.ChannelID)
    
    // Broadcast в workspace, а не только в channel
    h.wsHub.BroadcastToWorkspace(
        channel.WorkspaceID,
        "message",
        map[string]interface{}{
            "channel_id": message.ChannelID,
            "message":    message,
        },
        "",
    )
}
```

---

## 📋 Этап 4: Обновить Frontend Stores

### 4.1 WebSocket Store
**Файл:** `frontend/src/stores/websocket.ts`

**Добавить методы:**
```typescript
// Подписка на workspace
function subscribeToWorkspace(workspaceId: string) {
  console.log('[WS] 📡 Subscribing to workspace:', workspaceId)
  
  if (!connected.value) {
    console.warn('[WS] Not connected, delaying subscription...')
    setTimeout(() => subscribeToWorkspace(workspaceId), 100)
    return
  }
  
  send('subscribe_workspace', { workspace_id: workspaceId })
  currentWorkspaceId.value = workspaceId
}

// Отписка от workspace
function unsubscribeFromWorkspace(workspaceId: string) {
  console.log('[WS] 📡 Unsubscribing from workspace:', workspaceId)
  send('unsubscribe_workspace', { workspace_id: workspaceId })
}

// Подписка на channel для typing (опционально)
function subscribeToChannel(channelId: string) {
  send('subscribe_channel', { channel_id: channelId })
}
```

**Обновить handleEvent:**
```typescript
function handleEvent(event: WSEvent) {
  console.log('[WS] ← Received:', event.type, event.payload)
  
  switch (event.type) {
    // Channel events
    case 'channel_created':
      channelsStore.addChannel(event.payload as Channel)
      break
    
    case 'channel_updated':
      channelsStore.updateChannel(event.payload as Channel)
      break
    
    case 'channel_deleted': {
      const { channel_id } = event.payload as { channel_id: string }
      channelsStore.removeChannel(channel_id)
      break
    }
    
    // Message events
    case 'message': {
      const { channel_id, message } = event.payload as { channel_id: string, message: Message }
      messagesStore.addMessage(message)
      
      // Обновить unread count если не текущий канал
      if (channel_id !== currentChannelId.value) {
        channelsStore.incrementUnread(channel_id)
      }
      break
    }
    
    // Member events
    case 'member_added':
      workspaceStore.addMember(event.payload)
      break
    
    case 'member_removed':
      workspaceStore.removeMember(event.payload)
      break
    
    // ... остальные события
  }
}
```

### 4.2 Channels Store
**Файл:** `frontend/src/stores/channels.ts`

**Добавить методы для real-time обновлений:**
```typescript
function addChannel(channel: Channel) {
  const channels = channelsByWorkspace.value[channel.workspace_id] || []
  if (!channels.find(c => c.id === channel.id)) {
    channels.push(channel)
    channelsByWorkspace.value[channel.workspace_id] = channels
  }
}

function updateChannel(channel: Channel) {
  const channels = channelsByWorkspace.value[channel.workspace_id]
  if (channels) {
    const index = channels.findIndex(c => c.id === channel.id)
    if (index !== -1) {
      channels[index] = channel
    }
  }
}

function removeChannel(channelId: string) {
  Object.keys(channelsByWorkspace.value).forEach(workspaceId => {
    const channels = channelsByWorkspace.value[workspaceId]
    const index = channels.findIndex(c => c.id === channelId)
    if (index !== -1) {
      channels.splice(index, 1)
    }
  })
}

function incrementUnread(channelId: string) {
  // Увеличить счетчик непрочитанных
  unreadCounts.value[channelId] = (unreadCounts.value[channelId] || 0) + 1
}

function clearUnread(channelId: string) {
  unreadCounts.value[channelId] = 0
}
```

### 4.3 Workspace Store
**Файл:** `frontend/src/stores/workspace.ts`

**Добавить обработку участников:**
```typescript
function addMember(data: { workspace_id: string, user: User, role: string }) {
  // Добавить участника в список
}

function removeMember(data: { workspace_id: string, user_id: string }) {
  // Удалить участника из списка
}
```

---

## 📋 Этап 5: Обновить UI Components

### 5.1 AppView - подписка на workspace
**Файл:** `frontend/src/views/AppView.vue`

```vue
<script setup lang="ts">
onMounted(async () => {
  await authStore.fetchUser()
  await workspaceStore.fetchWorkspaces()
  
  // Подключаем WebSocket
  wsStore.connect()
})

// Подписываемся на workspace при его выборе
watch(
  () => workspaceStore.currentWorkspaceId,
  async (newId, oldId) => {
    // Отписываемся от старого
    if (oldId) {
      wsStore.unsubscribeFromWorkspace(oldId)
    }
    
    // Подписываемся на новый
    if (newId) {
      await channelsStore.fetchChannels(newId)
      wsStore.subscribeToWorkspace(newId)
    }
  },
  { immediate: true }
)
</script>
```

### 5.2 ChannelView - убрать подписку на channel
**Файл:** `frontend/src/views/ChannelView.vue`

```vue
<script setup lang="ts">
// Убрать wsStore.subscribe(channelId)
// Теперь подписка идет на workspace уровне

watch(
  () => route.params.channelId,
  async (channelId) => {
    if (channelId && typeof channelId === 'string') {
      channelsStore.setCurrentChannel(channelId)
      await messagesStore.fetchMessages(channelId)
      
      // Опционально: подписка на typing
      wsStore.subscribeToChannel(channelId)
      
      // Очистить непрочитанные
      channelsStore.clearUnread(channelId)
    }
  },
  { immediate: true }
)
</script>
```

### 5.3 Sidebar - показывать unread badges
**Файл:** `frontend/src/components/layout/Sidebar.vue`

```vue
<template>
  <div v-for="channel in channels" :key="channel.id">
    <router-link :to="`/channels/${channel.id}`">
      # {{ channel.name }}
      
      <!-- Badge с количеством непрочитанных -->
      <span v-if="channelsStore.unreadCounts[channel.id] > 0"
            class="badge">
        {{ channelsStore.unreadCounts[channel.id] }}
      </span>
    </router-link>
  </div>
</template>
```

---

## 📋 Этап 6: Добавить фильтрацию по правам

### 6.1 Проверка доступа на backend
**Файл:** `backend/internal/transport/ws/hub.go`

```go
type BroadcastMessage struct {
    WorkspaceID string
    Event       string
    Data        interface{}
    ExcludeID   string
    FilterFunc  func(*Client) bool  // Функция фильтрации
}

func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
    h.mu.RLock()
    clients := h.workspaces[msg.WorkspaceID]
    h.mu.RUnlock()
    
    for clientID, client := range clients {
        if clientID == msg.ExcludeID {
            continue
        }
        
        // Применяем фильтр если есть
        if msg.FilterFunc != nil && !msg.FilterFunc(client) {
            continue
        }
        
        // Отправляем сообщение
        select {
        case client.send <- data:
        default:
        }
    }
}
```

### 6.2 Использование фильтров
```go
// Пример: broadcast только для приватного канала
h.wsHub.BroadcastToWorkspaceWithFilter(
    workspaceID,
    "message",
    message,
    func(client *Client) bool {
        // Проверяем есть ли у пользователя доступ к приватному каналу
        return hasChannelAccess(client.UserID, channelID)
    },
)
```

---

## 📋 Этап 7: Тестирование

### 7.1 Unit тесты
- Тестирование Hub подписок
- Тестирование broadcast с фильтрацией
- Тестирование Client обработки событий

### 7.2 Integration тесты
- Подписка на workspace через WebSocket
- Получение событий создания канала
- Получение сообщений из разных каналов
- Unread counts

### 7.3 E2E тесты
- Два пользователя в одном workspace
- Создание канала одним - видит другой
- Сообщение в канале А - badge в UI пользователя Б

---

## 📋 Этап 8: Оптимизации (опционально)

### 8.1 Rate Limiting
```go
type Client struct {
    // ...
    rateLimiter *rate.Limiter
}
```

### 8.2 Event batching
```go
// Собирать события и отправлять батчами раз в 100ms
```

### 8.3 Compression
```go
// Сжатие больших payload (например списков участников)
```

### 8.4 Lazy loading
```typescript
// Загружать старые сообщения только при скролле
```

---

## 🗂 Файлы для изменения

### Backend (Go)
```
✏️  backend/internal/transport/ws/hub.go           - Изменить структуру подписок
✏️  backend/internal/transport/ws/client.go        - Добавить обработку workspace events
✏️  backend/internal/transport/http/handler/channel.go    - Добавить broadcast
✏️  backend/internal/transport/http/handler/workspace.go  - Добавить broadcast
✏️  backend/internal/transport/http/handler/message.go    - Изменить broadcast на workspace
🆕 backend/internal/transport/ws/events.go         - Константы типов событий
```

### Frontend (TypeScript/Vue)
```
✏️  frontend/src/stores/websocket.ts              - Добавить workspace подписку
✏️  frontend/src/stores/channels.ts               - Real-time updates
✏️  frontend/src/stores/workspace.ts              - Members updates
✏️  frontend/src/views/AppView.vue                - Подписка на workspace
✏️  frontend/src/views/ChannelView.vue            - Убрать channel подписку
✏️  frontend/src/components/layout/Sidebar.vue    - Unread badges
🆕 frontend/src/types/websocket.ts                - Типы WebSocket событий
```

---

## ⏱ Оценка времени

| Этап | Время | Сложность |
|------|-------|-----------|
| 1. Рефакторинг Hub | 2-3 часа | Средняя |
| 2. Обновление Client | 1-2 часа | Низкая |
| 3. Broadcast в Handlers | 2-3 часа | Средняя |
| 4. Frontend Stores | 2-3 часа | Средняя |
| 5. UI Components | 1-2 часа | Низкая |
| 6. Фильтрация прав | 2-3 часа | Высокая |
| 7. Тестирование | 2-4 часа | Средняя |
| **Итого** | **12-20 часов** | |

---

## 🎯 Приоритеты

### MVP (Must Have)
1. ✅ Подписка на workspace
2. ✅ Broadcast новых сообщений
3. ✅ Broadcast создания/удаления каналов
4. ✅ Real-time обновление списка каналов

### Phase 2 (Should Have)
5. ⭐ Unread badges/counts
6. ⭐ Member add/remove events
7. ⭐ Фильтрация по правам доступа

### Phase 3 (Nice to Have)
8. 💡 Typing indicators
9. 💡 Presence (online/offline)
10. 💡 @mentions notifications
11. 💡 Thread subscriptions

---

## 🚀 Готов начать реализацию?

План готов! Можно начинать с Этапа 1 когда скажете.
