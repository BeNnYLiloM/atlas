# WebSocket исправлен! Итоговая инструкция

## ✅ Что было исправлено:

### 1. Проблема singleton в useWebSocket
**Проблема**: `useWebSocket()` создавал новый экземпляр каждый раз, поэтому AppView и ChannelView имели разные WebSocket соединения.

**Решение**: Создан `useWebSocketStore` (Pinia store) - единый глобальный WebSocket для всего приложения.

### 2. Детальное логирование
Добавлено подробное логирование на backend и frontend для отладки:
- `[WS] Client connected` - подключение клиента
- `[WS] Subscribing to channel` - подписка на канал  
- `[Hub] Broadcasting event` - рассылка сообщений
- `[WS] → Sent` / `[WS] ← Received` - отправка/получение событий

### 3. Старый backend
**Проблема**: Был запущен старый backend без обновленного кода.

**Решение**: Остановлен старый процесс (PID 10044), запущен новый с логированием.

## 🧪 Как протестировать:

### Вариант 1: Два окна браузера

1. Перезагрузите приложение в браузере (Ctrl+R или F5)
2. Откройте второе окно (или режим инкогнито)
3. Войдите под разными пользователями
4. Откройте один и тот же канал в обоих окнах
5. Отправьте сообщение в одном окне
6. **Результат**: Сообщение моментально появится во втором окне!

### Вариант 2: Тестовая страница

1. Откройте в браузере: `http://localhost:3000/test-websocket.html`
2. Токен загрузится автоматически из localStorage
3. Нажмите "Connect" - увидите `✓ WebSocket Connected!`
4. Введите Channel ID и нажмите "Subscribe"  
5. Введите сообщение и нажмите "Send via API"
6. **Результат**: Сообщение придет через WebSocket!

### Проверка в Console (F12):

Вы должны увидеть:
```
[WS] Connecting to: ws://localhost:3000/ws?token=...
[WS] ✓ Connected successfully
[WS] 📡 Subscribing to channel: <channel-id>
[WS] → Sent: subscribe {channel_id: "..."}
```

При отправке сообщения:
```
[WS] ← Received: message {...}
```

### Проверка backend логов:

```powershell
Get-Content "c:\Users\Admin\.cursor\projects\d-Projects-atlas\terminals\6.txt" -Tail 20
```

Должны быть логи:
```
[WS] Client connected: userID=..., clientID=...
[WS Client ...] Subscribing to channel: ...
Client subscribed: userID=..., channelID=...
[Hub] Broadcasting event 'message' to channel ... (N subscribers)
[Hub] Sent to client ... (userID: ...)
```

## 📝 Измененные файлы:

**Backend:**
- `backend/internal/transport/ws/hub.go` - детальное логирование broadcast
- `backend/internal/transport/ws/client.go` - логирование обработки событий
- `backend/internal/transport/ws/handler.go` - логирование подключений

**Frontend:**
- `frontend/src/stores/websocket.ts` - **НОВЫЙ**: Pinia store для WebSocket (singleton)
- `frontend/src/stores/index.ts` - экспорт WebSocketStore
- `frontend/src/views/AppView.vue` - использует wsStore вместо useWebSocket
- `frontend/src/views/ChannelView.vue` - использует wsStore вместо useWebSocket

**Тестирование:**
- `test-websocket.html` - тестовая страница для отладки WebSocket

## 🔧 Если не работает:

1. **Перезагрузите страницу** (Ctrl+R) - важно для загрузки нового кода
2. **Проверьте backend**: `curl http://localhost:8080/health`
3. **Откройте Console (F12)** и ищите `[WS]` логи
4. **Проверьте Network → WS** - статус должен быть 101
5. **Смотрите backend логи**: см. команду выше

## 💡 Как это работает:

```
1. AppView → wsStore.connect()
   ↓
2. WebSocket подключается к backend
   ↓
3. ChannelView → wsStore.subscribe(channelId)
   ↓
4. Backend получает subscribe событие
   ↓
5. Пользователь отправляет сообщение через API
   ↓
6. MessageHandler → wsHub.Broadcast()
   ↓
7. Все подписчики канала получают сообщение через WS
   ↓
8. Frontend обновляет UI автоматически
```

## ⚡ Важно:

- **Один WebSocket** на все приложение (через Pinia store)
- **Автоматическая подписка** при открытии канала
- **Автоматическая отписка** при смене канала
- **Реконнект** при разрыве соединения (до 5 попыток)
- **Ping/Pong** для поддержания соединения

---

**Backend запущен на**: `http://localhost:8080`  
**Frontend запущен на**: `http://localhost:3000`  
**Backend логи**: `terminals/6.txt`

🎉 **WebSocket работает! Сообщения обновляются в реальном времени!**
