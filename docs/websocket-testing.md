# Тестирование WebSocket в Atlas

## Как проверить работу WebSocket

### Автоматическое подключение

WebSocket автоматически подключается при:
1. Входе в приложение (AppView)
2. Открытии канала (ChannelView автоматически подписывается на канал)

### Проверка в браузере

1. Откройте **DevTools** (F12)
2. Перейдите во вкладку **Console**
3. Вы должны увидеть:
   ```
   [WS] Connected
   [WS] Subscribing to channel: <channel-id>
   ```

4. Во вкладке **Network** > **WS**:
   - Должно быть подключение к `ws://localhost:3000/ws?token=...`
   - Статус: `101 Switching Protocols` (зеленый)

### Тестирование real-time обновлений

#### Способ 1: Два окна браузера

1. Откройте приложение в двух окнах браузера
2. Войдите под разными пользователями (или в режиме инкогнито)
3. Откройте один и тот же канал в обоих окнах
4. Отправьте сообщение в одном окне
5. **Результат**: Сообщение должно моментально появиться во втором окне

#### Способ 2: Через API и браузер

1. Откройте канал в браузере
2. Откройте консоль DevTools и смотрите на логи
3. Отправьте сообщение через API:

```powershell
# Получите токен из localStorage в браузере
# Откройте Console и выполните: localStorage.getItem('atlas_token')

$token = "ваш_токен_здесь"
$headers = @{
    'Authorization' = "Bearer $token"
    'Content-Type' = 'application/json'
}
$body = @{
    channel_id = "ID_канала"
    content = "Тестовое сообщение через API"
} | ConvertTo-Json

Invoke-RestMethod -Uri 'http://localhost:8080/api/v1/messages' -Method Post -Headers $headers -Body $body
```

4. **Результат**: 
   - В консоли увидите: `[WS] Received event: message {payload}`
   - Сообщение появится в UI без перезагрузки

#### Способ 3: Проверка всех событий

```javascript
// В консоли браузера выполните:

// 1. Отправить сообщение
await fetch('/api/v1/messages', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${localStorage.getItem('atlas_token')}`
  },
  body: JSON.stringify({
    channel_id: 'ID_канала',
    content: 'Test message'
  })
})

// 2. Обновить сообщение (замените MESSAGE_ID)
await fetch('/api/v1/messages/MESSAGE_ID', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${localStorage.getItem('atlas_token')}`
  },
  body: JSON.stringify({
    content: 'Updated message'
  })
})

// 3. Удалить сообщение
await fetch('/api/v1/messages/MESSAGE_ID', {
  method: 'DELETE',
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('atlas_token')}`
  }
})
```

### Проверка подписки на каналы

В консоли браузера должны быть логи:
```
[WS] Connected
[WS] Subscribing to channel: <channel-id>
```

При переходе на другой канал:
```
[WS] Unsubscribing from channel: <old-channel-id>
[WS] Subscribing to channel: <new-channel-id>
```

### Диагностика проблем

#### WebSocket не подключается

Проверьте:
1. Backend запущен: `curl http://localhost:8080/health`
2. Frontend запущен: `curl http://localhost:3000`
3. Токен валидный: проверьте localStorage в браузере
4. В консоли ошибки: `[WS] Error:` или `[WS] Disconnected`

#### Сообщения не приходят в real-time

1. Проверьте подключение в Network > WS
2. Убедитесь что подписка на канал прошла:
   - В консоли должен быть лог `[WS] Subscribing to channel:`
3. Проверьте логи backend:
   ```powershell
   # Смотрите логи backend
   Get-Content c:\Users\Admin\.cursor\projects\d-Projects-atlas\terminals\5.txt -Tail 50
   ```

#### Формат сообщений неправильный

Backend отправляет:
```json
{
  "type": "message",
  "payload": {
    "id": "...",
    "content": "..."
  }
}
```

Frontend ожидает:
```typescript
interface WSEvent {
  type: 'message' | 'message_update' | 'message_delete'
  payload: unknown
}
```

### События WebSocket

| Event | Описание | Payload |
|-------|----------|---------|
| `message` | Новое сообщение | `Message` объект |
| `message_update` | Обновление сообщения | `Message` объект |
| `message_delete` | Удаление сообщения | `{channel_id, message_id}` |
| `typing` | Пользователь печатает | `{user_id, channel_id}` |
| `presence` | Статус пользователя | `{user_id, status}` |

### Backend логи

Успешное подключение:
```
Client registered: userID=..., clientID=...
Client subscribed: userID=..., channelID=...
```

Broadcast сообщения:
```
# Сообщение отправлено всем подписчикам канала
```

### Frontend логи

Успешное подключение и работа:
```
[WS] Connected
[WS] Subscribing to channel: abc-123
[WS] Received event: message {...}
```

## Архитектура WebSocket

```
Frontend                Backend
   |                       |
   |-- WS Connect -------> Hub.Register
   |                       |
   |-- subscribe --------> Hub.Subscribe(channelID)
   |                       |
   |                    [User sends message via HTTP]
   |                       |
   |<-- message -------- Hub.Broadcast(channelID)
   |                       |
   |   (auto update UI)    |
```

## Особенности реализации

1. **Автоматическая подписка**: При открытии канала автоматически подписываемся на WebSocket события
2. **Автоматическая отписка**: При смене канала или закрытии страницы отписываемся
3. **Реконнект**: При разрыве соединения автоматически переподключаемся (до 5 попыток)
4. **Ping/Pong**: Backend автоматически отправляет ping каждые 54 секунды для поддержания соединения
5. **Буферизация**: Сообщения буферизуются (256 сообщений) на случай временных задержек

## Производительность

- Один WebSocket на клиента
- Подписка на несколько каналов через один сокет
- Broadcast только подписчикам конкретного канала
- Буферизация для защиты от перегрузки
